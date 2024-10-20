// cmd/rest-api/models/auth.go
package models

import (
	"gorm.io/gorm"
)

// User represents a user in the system.
type User struct {
	gorm.Model
	UID         string `gorm:"type:uuid;uniqueIndex"`
	PhoneNumber string `json:"phone_number" gorm:"unique;not null"`
	FirstName   string `json:"first_name" gorm:"not null;default:''"`
	LastName    string `json:"last_name" gorm:"not null;default:''"`
	Address     string `json:"address" gorm:"not null;default:''"`
	Pin         string `json:"pin" gorm:"not null;default:''"`

	Accounts []UserAccount `gorm:"foreignKey:UserID"`
}
