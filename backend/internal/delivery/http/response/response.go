package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type PaginatedMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Data    any           `json:"data"`
	Meta    PaginatedMeta `json:"meta"`
}

func JSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func Success(w http.ResponseWriter, statusCode int, message string, data any) {
	JSON(w, statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, statusCode int, message string, errDetails any) {
	JSON(w, statusCode, APIResponse{
		Success: false,
		Message: message,
		Errors:  errDetails,
	})
}

func Paginated(w http.ResponseWriter, statusCode int, message string, data any, page, limit int, totalItems int64) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	totalPages := int((totalItems + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	JSON(w, statusCode, PaginatedResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta: PaginatedMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	})
}
