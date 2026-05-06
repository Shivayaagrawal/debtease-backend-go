package models

type PdfRequestRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type PdfRequestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type WaitlistRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type WaitlistResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
