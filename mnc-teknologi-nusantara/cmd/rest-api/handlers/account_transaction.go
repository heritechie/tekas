package handlers

import (
	"encoding/json"
	"errors"
	"mnctech-restapi/cmd/rest-api/auth"
	"mnctech-restapi/cmd/rest-api/models"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TopUpRequest represents the expected payload for a top-up request.
type TopUpRequest struct {
	Amount float64 `json:"amount" validate:"required"` // Amount to top up
}

// TopUpResult represents the structure of the top-up response result.
type TopUpResult struct {
	TopUpID       string    `json:"top_up_id"`      // Unique identifier for the top-up transaction
	AmountTopUp   float64   `json:"amount_top_up"`  // Amount that was topped up
	BalanceBefore float64   `json:"balance_before"` // Balance before the top-up
	BalanceAfter  float64   `json:"balance_after"`  // Balance after the top-up
	CreatedDate   time.Time `json:"created_date"`   // Date of the top-up
}

// TopUp handles user top-ups.
func (h *AppHandler) HandleTopUp(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(NewFailedResponse("User ID not found in context"))
		return
	}

	var req TopUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(NewFailedResponse("Invalid request payload"))
		return
	}

	// Validate input
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	topupTrxID := uuid.New().String()

	// Create the response
	topupResult := TopUpResult{
		TopUpID:     topupTrxID, // Use the transaction ID here if needed
		AmountTopUp: req.Amount,
	}

	// Start a new transaction
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("uid = ?", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found") // Return an error if the user does not exist
			}
			return err // Return the error to rollback the transaction for other errors
		}
		// Fetch the user's account
		var userAccount models.UserAccount
		if err := tx.Where("user_id = ?", user.ID).First(&userAccount).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// If not found, create a new UserAccount
				userAccount = models.UserAccount{
					UserID:         user.ID, // Use the user's ID from the User model
					CurrentBalance: 0,       // Set initial balance or any other default values
					LastBalance:    0,
				}

				if err := tx.Create(&userAccount).Error; err != nil {
					return err // Return the error to rollback the transaction
				}
			} else {
				return err // Return the error to rollback the transaction for other errors
			}
		}

		// Update balances
		userAccount.LastBalance = userAccount.CurrentBalance // Save the previous balance
		userAccount.CurrentBalance += req.Amount             // Update current balance

		// Save the updated account
		if err := tx.Save(&userAccount).Error; err != nil {
			return err // Return the error to rollback the transaction
		}

		// Create a top-up transaction record
		topUpTransactionPayload := models.TopUpTransaction{
			UID:           topupTrxID,
			UserAccountID: userAccount.ID, // Reference to the user's account
			Amount:        req.Amount,     // The amount topped up
			Status:        "SUCCESS",      // Status of the transaction
		}

		if err := tx.Create(&topUpTransactionPayload).Error; err != nil {
			return err // Return the error to rollback the transaction
		}

		// Fetch the user's account
		var updatedUserAccount models.UserAccount
		if err := tx.Where("user_id = ?", user.ID).First(&updatedUserAccount).Error; err != nil {
			return err
		}

		// Create an account transaction log record
		transactionLog := models.AccountTransactionLog{
			UserAccountID: userAccount.ID,
			Amount:        req.Amount,
			BalanceBefore: updatedUserAccount.LastBalance,
			BalanceAfter:  updatedUserAccount.CurrentBalance,
		}

		if err := tx.Create(&transactionLog).Error; err != nil {
			return err // Return the error to rollback the transaction
		}

		var topupTransaction models.TopUpTransaction
		if err := tx.Where("uid = ?", topupTrxID).First(&topupTransaction).Error; err != nil {
			return err
		}

		topupResult.BalanceBefore = updatedUserAccount.LastBalance
		topupResult.BalanceAfter = updatedUserAccount.CurrentBalance
		topupResult.CreatedDate = topupTransaction.CreatedAt

		return nil // Commit the transaction if everything is successful
	})

	if err != nil {
		if err.Error() == "user not found" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(NewFailedResponse("User not found"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(NewFailedResponse("Transaction failed"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NewSuccessResponse(topupResult))
}
