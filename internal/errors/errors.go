package errors

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrNotFound            = NewAppError("NOT_FOUND", "Recurso não encontrado", http.StatusNotFound)
	ErrUnauthorized        = NewAppError("UNAUTHORIZED", "Não autorizado", http.StatusUnauthorized)
	ErrForbidden           = NewAppError("FORBIDDEN", "Acesso negado", http.StatusForbidden)
	ErrBadRequest          = NewAppError("BAD_REQUEST", "Requisição inválida", http.StatusBadRequest)
	ErrInternalServer      = NewAppError("INTERNAL_SERVER_ERROR", "Erro interno do servidor", http.StatusInternalServerError)
	ErrConflict            = NewAppError("CONFLICT", "Conflito de recursos", http.StatusConflict)
	ErrValidation          = NewAppError("VALIDATION_ERROR", "Erro de validação", http.StatusBadRequest)
	ErrDatabase            = NewAppError("DATABASE_ERROR", "Erro no banco de dados", http.StatusInternalServerError)
	ErrInvalidCredentials  = NewAppError("INVALID_CREDENTIALS", "Credenciais inválidas", http.StatusUnauthorized)
	ErrEmailAlreadyExists  = NewAppError("EMAIL_ALREADY_EXISTS", "Email já cadastrado", http.StatusConflict)
	ErrUserNotFound        = NewAppError("USER_NOT_FOUND", "Usuário não encontrado", http.StatusNotFound)
	ErrTransactionNotFound = NewAppError("TRANSACTION_NOT_FOUND", "Transação não encontrada", http.StatusNotFound)
	ErrGoalNotFound        = NewAppError("GOAL_NOT_FOUND", "Meta não encontrada", http.StatusNotFound)
	ErrInvestmentNotFound  = NewAppError("INVESTMENT_NOT_FOUND", "Investimento não encontrado", http.StatusNotFound)
	ErrCategoryNotFound    = NewAppError("CATEGORY_NOT_FOUND", "Categoria não encontrada", http.StatusNotFound)
	ErrResourceNotOwned    = NewAppError("RESOURCE_NOT_OWNED", "Recurso não pertence ao usuário", http.StatusForbidden)
)

type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Details    map[string]interface{}
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	clone := e.clone()
	if details == nil {
		clone.Details = make(map[string]interface{})
		return clone
	}
	clone.Details = make(map[string]interface{}, len(details))
	for k, v := range details {
		clone.Details[k] = v
	}
	return clone
}

func (e *AppError) WithError(err error) *AppError {
	clone := e.clone()
	clone.Err = err
	return clone
}

func NewAppError(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

func WrapError(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
		Details:    make(map[string]interface{}),
	}
}

func (e *AppError) clone() *AppError {
	if e == nil {
		return nil
	}
	clone := *e
	if e.Details != nil {
		clone.Details = make(map[string]interface{}, len(e.Details))
		for k, v := range e.Details {
			clone.Details[k] = v
		}
	} else {
		clone.Details = make(map[string]interface{})
	}
	return &clone
}

func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

func AsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

func FromError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	if errors.Is(err, ErrNotFound.Err) {
		return ErrNotFound.WithError(err)
	}

	if errors.Is(err, context.Canceled) {
		return WrapError(err, "REQUEST_CANCELED", "Requisição cancelada pelo cliente", http.StatusRequestTimeout)
	}

	return WrapError(err, "UNKNOWN_ERROR", "Erro desconhecido", http.StatusInternalServerError)
}

func NewAuthError(code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Details:    make(map[string]interface{}),
	}
}

func NewValidationError(field, message string) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Details:    make(map[string]interface{}),
	}
}

func NewDatabaseError(err error) *AppError {
	return WrapError(err, "DATABASE_ERROR", "Erro ao executar operação no banco de dados", http.StatusInternalServerError)
}

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s não encontrado", resource),
		StatusCode: http.StatusNotFound,
		Details: map[string]interface{}{
			"resource": resource,
		},
	}
}

func NewConflictError(resource string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    fmt.Sprintf("%s já existe", resource),
		StatusCode: http.StatusConflict,
		Details: map[string]interface{}{
			"resource": resource,
		},
	}
}

func ParseValidationErrors(err error) *AppError {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return ErrBadRequest.WithError(err)
	}

	fieldErrors := make([]map[string]string, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		translatedField := translateFieldName(fieldErr.Field())
		fieldErrors = append(fieldErrors, map[string]string{
			"field":   translatedField,
			"message": translateValidationError(fieldErr),
		})
	}

	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Erro de validação nos campos",
		StatusCode: http.StatusBadRequest,
		Details: map[string]interface{}{
			"fields": fieldErrors,
		},
	}
}

func translateFieldName(field string) string {
	fieldLower := strings.ToLower(field)
	fieldMap := map[string]string{
		"amount":       "valor",
		"account_id":   "conta",
		"accountid":    "conta",
		"category_id":  "categoria",
		"categoryid":   "categoria",
		"type":         "tipo",
		"description":  "descrição",
		"name":         "nome",
		"email":        "email",
		"password":     "senha",
		"date":         "data",
		"target":       "valor alvo",
		"targetamount": "valor alvo",
	}
	if translated, ok := fieldMap[fieldLower]; ok {
		return translated
	}
	return field
}

func translateValidationError(fe validator.FieldError) string {
	fieldName := translateFieldName(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s é obrigatório", fieldName)
	case "email":
		return "Email inválido"
	case "min":
		return fmt.Sprintf("%s deve ter no mínimo %s caracteres", fieldName, fe.Param())
	case "max":
		return fmt.Sprintf("%s deve ter no máximo %s caracteres", fieldName, fe.Param())
	case "gte":
		return fmt.Sprintf("%s deve ser maior ou igual a %s", fieldName, fe.Param())
	case "lte":
		return fmt.Sprintf("%s deve ser menor ou igual a %s", fieldName, fe.Param())
	case "gt":
		return fmt.Sprintf("%s deve ser maior que %s", fieldName, fe.Param())
	case "lt":
		return fmt.Sprintf("%s deve ser menor que %s", fieldName, fe.Param())
	case "ne":
		return fmt.Sprintf("%s deve ser diferente de %s", fieldName, fe.Param())
	case "len":
		return fmt.Sprintf("%s deve ter exatamente %s caracteres", fieldName, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s deve ser um dos valores: %s", fieldName, fe.Param())
	case "uuid":
		return fmt.Sprintf("%s deve ser um UUID válido", fieldName)
	case "url":
		return fmt.Sprintf("%s deve ser uma URL válida", fieldName)
	case "datetime":
		return fmt.Sprintf("%s deve ser uma data/hora válida", fieldName)
	case "numeric":
		return fmt.Sprintf("%s deve ser um valor numérico", fieldName)
	case "alphanum":
		return fmt.Sprintf("%s deve conter apenas letras e números", fieldName)
	default:
		return fmt.Sprintf("Validação '%s' falhou para %s", fe.Tag(), fieldName)
	}
}
