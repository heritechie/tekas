package models

import (
	"gorm.io/gorm"
)

// TopUpTransaction represents a transaction for a top-up.
type TopUpTransaction struct {
	gorm.Model
	UID           string  `gorm:"type:uuid;uniqueIndex"`
	UserAccountID uint    `json:"user_account_id" gorm:"not null"`          // Foreign key referencing UserAccount
	Amount        float64 `json:"amount" gorm:"not null"`                   // The amount of the top-up
	Status        string  `json:"status" gorm:"not null;default:'PENDING'"` // Status of the transaction (e.g., SUCCESS, FAILED)

	UserAccount UserAccount `gorm:"foreignKey:UserAccountID"`
}

// PaymentTransaction represents a transaction for a payment.
type PaymentTransaction struct {
	gorm.Model
	UID           string      `gorm:"type:uuid;uniqueIndex"`
	UserAccountID uint        `json:"user_account_id" gorm:"not null"` // Foreign key referencing UserAccount
	Amount        float64     `json:"amount" gorm:"not null"`          // The amount of the top-up
	Remarks       string      `json:"remarks" gorm:"not null"`
	Status        string      `json:"status" gorm:"not null;default:'PENDING'"` // Status of the transaction (e.g., SUCCESS, FAILED)
	UserAccount   UserAccount `gorm:"foreignKey:UserAccountID"`
}

// PaymentTransaction represents a transaction for a payment.
type TransferTransaction struct {
	gorm.Model
	UID           string      `gorm:"type:uuid;uniqueIndex"`
	UserAccountID uint        `json:"user_account_id" gorm:"not null"` // Foreign key referencing UserAccount
	TargetUserID  uint        `json:"target_account_id" gorm:"not null"`
	Amount        float64     `json:"amount" gorm:"not null"` // The amount of the top-up
	Remarks       string      `json:"remarks" gorm:"not null"`
	Status        string      `json:"status" gorm:"not null;default:'PENDING'"` // Status of the transaction (e.g., SUCCESS, FAILED)
	UserAccount   UserAccount `gorm:"foreignKey:UserAccountID"`
	TargetAccount UserAccount `gorm:"foreignKey:UserAccountID"`
}
