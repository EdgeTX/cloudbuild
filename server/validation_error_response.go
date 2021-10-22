package server

type ValidationErrorResponse struct {
	Error            string   `json:"error"`
	ValidationErrors []string `json:"validation_errors"`
}

func NewValidationErrorResponse(message string, errs []error) *ValidationErrorResponse {
	validationErrors := make([]string, 0)
	for _, err := range errs {
		validationErrors = append(validationErrors, err.Error())
	}
	return &ValidationErrorResponse{
		Error:            message,
		ValidationErrors: validationErrors,
	}
}
