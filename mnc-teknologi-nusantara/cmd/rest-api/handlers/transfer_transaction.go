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

type TransferRequest struct {
	TargetUser string  `json:"target_user"`
	Amount     float64 `json:"amount"`
	Remarks    string  `json:"remarks"`
}

type TransferResult struct {
	TransferID    string    `json:"transfer_id"`
	Amount        float64   `json:"amount"`
	Remarks       string    `json:"remarks"`
	BalanceBefore float64   `json:"balance_before"`
	BalanceAfter  float64   `json:"balance_after"`
	CreatedDate   time.Time `json:"created_date"`
}

func (h *AppHandler) HandleTransfer(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(NewFailedResponse("User ID not found in context"))
		return
	}

	// Parse request body
	var req TransferRequest
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

	transferTrxID := uuid.New().String()

	transferResult := TransferResult{
		TransferID: transferTrxID,
		Amount:     req.Amount,
		Remarks:    req.Remarks,
	}

	baseTransactionLog := models.AccountTransactionLog{
		TransactionCategory: "TRANSFER",
		Amount:              req.Amount,
		Remarks:             req.Remarks,
		TransactionReff:     transferTrxID,
	}

	var userAccount models.UserAccount

	// Start a new transaction
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("uid = ?", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found") // Return an error if the user does not exist
			}
			return err // Return the error to rollback the transaction for other errors
		}

		if err := tx.Where("user_id = ?", user.ID).First(&userAccount).Error; err != nil {
			return err
		}

		var targetUser models.User
		if err := tx.Where("uid = ?", req.TargetUser).First(&targetUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("target user not found") // Return an error if the user does not exist
			}
			return err // Return the error to rollback the transaction for other errors
		}

		if userAccount.CurrentBalance < req.Amount {
			return errors.New("insufficient balance")
		}

		// Deduct user balances
		userAccount.LastBalance = userAccount.CurrentBalance // Save the previous balance
		userAccount.CurrentBalance -= req.Amount             // Update curr

		// Save the updated account
		if err := tx.Save(&userAccount).Error; err != nil {
			return err // Return the error to rollback the transaction
		}

		// Fetch the user's account
		var updatedUserAccount models.UserAccount
		if err := tx.Where("user_id = ?", user.ID).First(&updatedUserAccount).Error; err != nil {
			return err
		}

		transactionLog := baseTransactionLog
		transactionLog.UserAccountID = userAccount.ID
		transactionLog.TransactionType = "DEBIT"
		transactionLog.BalanceBefore = updatedUserAccount.LastBalance
		transactionLog.BalanceAfter = updatedUserAccount.CurrentBalance

		if err := tx.Create(&transactionLog).Error; err != nil {
			return err // Return the error to rollback the transaction
		}

		var targetAccount models.UserAccount
		if err := tx.Where("user_id = ?", targetUser.ID).First(&targetAccount).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// If not found, create a new UserAccount
				targetAccount = models.UserAccount{
					UserID:         targetUser.ID, // Use the user's ID from the User model
					CurrentBalance: 0,             // Set initial balance or any other default values
					LastBalance:    0,
				}

				if err := tx.Create(&targetAccount).Error; err != nil {
					return err // Return the error to rollback the transaction
				}
			} else {
				return err // Return the error to rollback the transaction for other errors
			}
		}

		// Update tager balances
		targetAccount.LastBalance = targetAccount.CurrentBalance // Save the previous balance
		targetAccount.CurrentBalance += req.Amount               // Update curr

		// Save the updated account
		if err := tx.Save(&targetAccount).Error; err != nil {
			return err // Return the error to rollback the transaction
		}

		// Fetch the user's account
		var updatedTargetAccount models.UserAccount
		if err := tx.Where("user_id = ?", targetAccount.ID).First(&updatedTargetAccount).Error; err != nil {
			return err
		}

		transactionLogTarget := baseTransactionLog
		transactionLogTarget.UserAccountID = updatedTargetAccount.ID
		transactionLogTarget.TransactionType = "CREDIT"
		transactionLogTarget.BalanceBefore = updatedTargetAccount.LastBalance
		transactionLogTarget.BalanceAfter = updatedTargetAccount.CurrentBalance

		if err := tx.Create(&transactionLogTarget).Error; err != nil {
			return err // Return the error to rollback the transaction
		}

		transferTransactionPayoad := models.TransferTransaction{
			UID:           transferTrxID,
			UserAccountID: userAccount.ID, // Reference to the user's account
			TargetUserID:  targetAccount.ID,
			Remarks:       req.Remarks,
			Amount:        req.Amount, // The amount topped up
			Status:        "SUCCESS",  // Status of the transaction
		}

		if err := tx.Create(&transferTransactionPayoad).Error; err != nil {
			return err // Return the error to rollback the transaction
		}

		var transferTransaction models.TransferTransaction
		if err := tx.Where("uid = ?", transferTrxID).First(&transferTransaction).Error; err != nil {
			return err
		}

		transferResult.BalanceBefore = updatedUserAccount.LastBalance
		transferResult.BalanceAfter = updatedUserAccount.CurrentBalance
		transferResult.CreatedDate = transferTransaction.CreatedAt

		return nil // Commit the transaction if everything is successful
	})

	if err != nil {
		if err.Error() == "insufficient balance" {
			errMessage := "Balance is not enough"
			transactionLog := baseTransactionLog
			transactionLog.UserAccountID = userAccount.ID
			transactionLog.TransactionType = "DEBIT"
			transactionLog.BalanceBefore = userAccount.LastBalance
			transactionLog.BalanceAfter = userAccount.CurrentBalance
			transactionLog.Status = "FAILED"
			transactionLog.ErrMessage = errMessage

			if err := h.DB.Create(&transactionLog).Error; err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(NewFailedResponse(errMessage))
			}

			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(NewFailedResponse(errMessage))
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
	json.NewEncoder(w).Encode(NewSuccessResponse(transferResult))
}
