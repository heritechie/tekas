package handlers

import "gorm.io/gorm"

type AppHandler struct {
	DB *gorm.DB
}

type SuccessResponse struct {
	Status string      `json:"status"`
	Result interface{} `json:"result"` // Result can hold any type of data
}

type FailedResponse struct {
	Message string `json:"message"`
}

func NewSuccessResponse(result interface{}) SuccessResponse {
	return SuccessResponse{
		Status: "SUCCESS", // Default status
		Result: result,    // Dynamic result data
	}
}

func NewFailedResponse(failedMessage string) FailedResponse {
	return FailedResponse{
		Message: failedMessage,
	}
}
