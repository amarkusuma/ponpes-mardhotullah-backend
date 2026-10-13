// Package apperror mendefinisikan tipe error standar untuk seluruh aplikasi.
package apperror

import (
	"fmt"
	"net/http"
)

// AppError adalah custom error yang membawa HTTP status code.
type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Detail)
	}
	return e.Message
}

// Sentinel errors — clone sebelum mengubah Detail agar tidak race condition.
var (
	ErrBadRequest   = &AppError{Code: http.StatusBadRequest, Message: "permintaan tidak valid"}
	ErrUnauthorized = &AppError{Code: http.StatusUnauthorized, Message: "tidak terautentikasi"}
	ErrForbidden    = &AppError{Code: http.StatusForbidden, Message: "akses ditolak"}
	ErrNotFound     = &AppError{Code: http.StatusNotFound, Message: "data tidak ditemukan"}
	ErrConflict     = &AppError{Code: http.StatusConflict, Message: "data sudah ada"}
	ErrInternal     = &AppError{Code: http.StatusInternalServerError, Message: "kesalahan internal server"}
)

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func ErrValidation(detail string) *AppError {
	return &AppError{Code: http.StatusUnprocessableEntity, Message: "validasi gagal", Detail: detail}
}

func ErrNotFoundMsg(resource string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: fmt.Sprintf("%s tidak ditemukan", resource)}
}

func ErrConflictMsg(resource string) *AppError {
	return &AppError{Code: http.StatusConflict, Message: fmt.Sprintf("%s sudah ada", resource)}
}
