package httpx

import "net/http"

type MappedError struct {
	Status  int
	Code    string
	Message string
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) error {
	return WriteJSON(
		w,
		status,
		ErrorResponse{
			Error: ErrorBody{
				Code:    code,
				Message: message,
			},
		},
	)
}

func AppendLogError(attrs []any, err error) []any {
	result := make([]any, 0, len(attrs)+2)
	result = append(result, attrs...)
	return append(result, "error", err)
}
