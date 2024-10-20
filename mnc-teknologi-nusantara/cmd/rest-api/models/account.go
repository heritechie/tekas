package models

import (
	"gorm.io/gorm"
)

type UserAccount struct {
	gorm.Model
	UserID         uint    `gorm:"not null;foreignKey:UserID;references:ID"`
	CurrentBalance float64 `json:"current_balance" gorm:"not null;default:0"`
	LastBalance    float64 `json:"last_balance" gorm:"not null;default:0"`

	User User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:UserID"`
}

// AccountTransaction represents a transaction for a user's account.
type AccountTransactionLog struct {
	gorm.Model
	UserAccountID   uint    `json:"user_account_id" gorm:"not null"`  // Foreign key referencing UserAccount
	Amount          float64 `json:"amount" gorm:"not null"`           // The amount of the transaction
	TransactionType string  `json:"transaction_type" gorm:"not null"` // Type: "topup", "payment", "transfer" etc.
	TransactionReff string  `json:"transaction_reff"`
	BalanceBefore   float64 `json:"balance_before"` // Balance before the transaction
	BalanceAfter    float64 `json:"balance_after"`  // Balance after the transaction

	UserAccount UserAccount `gorm:"foreignKey:UserAccountID"`
}
