package types

type APIResponse struct {
	Title   string
	Message string
	Status  string
	Data    interface{}
}

func NewSuccessResponse(title, message string, data interface{}) APIResponse {
	return APIResponse{
		Title:   title,
		Message: message,
		Status:  "success",
		Data:    data,
	}
}

func NewErrorResponse(title, message string) APIResponse {
	return APIResponse{
		Title:   title,
		Message: message,
		Status:  "error",
		Data:    nil,
	}
}
