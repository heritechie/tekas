package handlers

import (
	"encoding/json"
	"mnctech-restapi/cmd/rest-api/auth"
	"mnctech-restapi/cmd/rest-api/models"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	*AppHandler
	AccessTokenKey  []byte // Include the access token key
	RefreshTokenKey []byte // Include the refresh token key
}

// RegisterRequest represents the expected payload for registration.
type RegisterRequest struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required"`
	Address     string `json:"address" validate:"required"`
	PIN         string `json:"pin" validate:"required"`
}

type RegisterResponse struct {
	UserID      string `json:"user_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	CreatedDate string `json:"created_date"`
}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	PIN         string `json:"pin" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// DB variable to be initialized in main
var DB *gorm.DB

// Register handles user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate input
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the phone number already exists
	var existingUser models.User
	if err := h.DB.Where("phone_number = ?", req.PhoneNumber).First(&existingUser).Error; err == nil {
		http.Error(w, "Phone number already registered", http.StatusConflict)
		return
	}

	// Hash the PIN before saving it
	hashedPin, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash PIN", http.StatusInternalServerError)
		return
	}

	// Create user
	user := models.User{
		UID:         uuid.New().String(), // Generate UUID here
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		Pin:         string(hashedPin), // Store the hashed PIN
	}

	// Insert user into the database using GORM
	if err := h.DB.Create(&user).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(NewFailedResponse("Gagal mendaftarkan pengguna"))
		return
	}

	userResponse := RegisterResponse{
		UserID:      user.UID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
		Address:     user.Address,
		CreatedDate: user.CreatedAt.Format("2006-01-02 15:04:05"), // Format as desired
	}

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(NewSuccessResponse(userResponse))
}

// Login handles user login and generates a JWT token.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate input
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user models.User
	// Find user by phone number
	if err := h.DB.Where("phone_number = ?", req.PhoneNumber).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Check if PIN matches
	if err := bcrypt.CompareHashAndPassword([]byte(user.Pin), []byte(req.PIN)); err != nil {
		http.Error(w, "Invalid PIN", http.StatusUnauthorized)
		return
	}

	// Create JWT token
	accessExpirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours
	accessClaims := &auth.CustomClaims{
		UID: user.UID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpirationTime),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	// Sign the access token with the secret key
	accessTokenString, err := accessToken.SignedString(h.AccessTokenKey)
	if err != nil {
		http.Error(w, "Could not create access token", http.StatusInternalServerError)
		return
	}

	// Create JWT refresh token
	refreshExpirationTime := time.Now().Add(7 * 24 * time.Hour) // Token expires in 7 days
	refreshClaims := &auth.CustomClaims{
		UID: user.UID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpirationTime),
		}}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	// Sign the refresh token with the secret key
	refreshTokenString, err := refreshToken.SignedString(h.RefreshTokenKey)
	if err != nil {
		http.Error(w, "Could not create refresh token", http.StatusInternalServerError)
		return
	}

	// Respond with the tokens
	response := LoginResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NewSuccessResponse(response))
}
