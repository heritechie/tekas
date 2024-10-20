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

type PaymentRequest struct {
	Amount  float64 `json:"amount"`
	Remarks string  `json:"remarks"`
}

type PaymentResult struct {
	PaymentID     string    `json:"payment_id"`
	Amount        float64   `json:"amount"`
	Remarks       string    `json:"remarks"`
	BalanceBefore float64   `json:"balance_before"`
	BalanceAfter  float64   `json:"balance_after"`
	CreatedDate   time.Time `json:"created_date"`
}

func (h *AppHandler) HandlePayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(NewFailedResponse("User ID not found in context"))
		return
	}

	// Parse request body
	var req PaymentRequest
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

	paymentTrxID := uuid.New().String()

	paymentResult := PaymentResult{
		PaymentID: paymentTrxID,
		Amount:    req.Amount,
		Remarks:   req.Remarks,
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
			return err
		}

		if userAccount.CurrentBalance < req.Amount {
			return errors.New("balance is not enough")
		}

		// Update balances
		userAccount.LastBalance = userAccount.CurrentBalance // Save the previous balance
		userAccount.CurrentBalance -= req.Amount             // Update curr

		// Save the updated account
		if err := tx.Save(&userAccount).Error; err != nil {
			return err // Return the error to rollback the transaction
		}

		paymentTransactionPayoad := models.PaymentTransaction{
			UID:           paymentTrxID,
			UserAccountID: userAccount.ID, // Reference to the user's account
			Remarks:       req.Remarks,
			Amount:        req.Amount, // The amount topped up
			Status:        "SUCCESS",  // Status of the transaction
		}

		if err := tx.Create(&paymentTransactionPayoad).Error; err != nil {
			return err // Return the error to rollback the transaction
		}

		// Fetch the user's account
		var updatedUserAccount models.UserAccount
		if err := tx.Where("user_id = ?", user.ID).First(&updatedUserAccount).Error; err != nil {
			return err
		}

		// Create an account transaction log record
		transactionLog := models.AccountTransactionLog{
			UserAccountID: updatedUserAccount.ID,
			Amount:        req.Amount,
			BalanceBefore: updatedUserAccount.LastBalance,
			BalanceAfter:  updatedUserAccount.CurrentBalance,
		}

		if err := tx.Create(&transactionLog).Error; err != nil {
			return err // Return the error to rollback the transaction
		}

		var paymentTransaction models.PaymentTransaction
		if err := tx.Where("uid = ?", paymentTrxID).First(&paymentTransaction).Error; err != nil {
			return err
		}

		paymentResult.BalanceBefore = updatedUserAccount.LastBalance
		paymentResult.BalanceAfter = updatedUserAccount.CurrentBalance
		paymentResult.CreatedDate = paymentTransaction.CreatedAt

		return nil // Commit the transaction if everything is successful
	})

	if err != nil {
		if err.Error() == "balance is not enough" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(NewFailedResponse("Insufficient balance"))
		} else if err.Error() == "user not found" {
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
	json.NewEncoder(w).Encode(NewSuccessResponse(paymentResult))
}
