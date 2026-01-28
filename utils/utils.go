package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, APIResponse{Success: false, Error: message})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func ParseIntParam(r *http.Request, key string) (int, error) {
	vars := mux.Vars(r)
	idStr, ok := vars[key]
	if !ok {
		return 0, errors.New("missing path parameter")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.New("invalid path parameter")
	}

	return id, nil
}

func DecodeJSONBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func IsValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func HashPassword(password string) (string, error) {
	return "", nil
}

func VerifyPassword(hashPassword, plainPassword string) bool {
	return false
}

func GenerateToken(userID string, expiryHours int) (string, error) {
	return "", nil
}

func ValidateToken(tokenString string) (string, error) {
	return "", nil
}

type PaginationParams struct {
	Page      int    `json:"page" validate:"required,min=1"`
	PageSize  int    `json:"page_size" validate:"required,min=1,max=100"`
	SortBy    string `json:"sort_by" validate:"omitempty"`
	SortOrder string `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}
