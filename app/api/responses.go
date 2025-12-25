package api

import (
	"encoding/json"
	"net/http"
)

// Response представляет стандартный JSON ответ
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    int    `json:"code"`
}

// SendJSON отправляет JSON ответ
func SendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// SendSuccess отправляет успешный ответ
func SendSuccess(w http.ResponseWriter, data interface{}) {
	SendJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// SendCreated отправляет ответ о создании ресурса
func SendCreated(w http.ResponseWriter, data interface{}) {
	SendJSON(w, http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// SendError отправляет ответ с ошибкой
func SendError(w http.ResponseWriter, statusCode int, message string) {
	SendJSON(w, statusCode, ErrorResponse{
		Success: false,
		Error:   message,
		Code:    statusCode,
	})
}

// SendBadRequest отправляет ответ 400
func SendBadRequest(w http.ResponseWriter, message string) {
	SendError(w, http.StatusBadRequest, message)
}

// SendNotFound отправляет ответ 404
func SendNotFound(w http.ResponseWriter, message string) {
	SendError(w, http.StatusNotFound, message)
}

// SendInternalError отправляет ответ 500
func SendInternalError(w http.ResponseWriter, message string) {
	SendError(w, http.StatusInternalServerError, message)
}

// SendUnauthorized отправляет ответ 401
func SendUnauthorized(w http.ResponseWriter, message string) {
	SendError(w, http.StatusUnauthorized, message)
}

// SendNoContent отправляет ответ 204
func SendNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
