package handler

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"simple-tasks/internal/middleware"
)

type ErrorDetail struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type ErrorInfo struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorResponse struct {
	Error     ErrorInfo `json:"error"`
	RequestId string    `json:"requestId"`
}

type ErrType = int

const (
	errorInvalidJson ErrType = iota
	errorValidation
	errorNotFound
	errorBadRequest
	errorInternal
)

var codeMap = map[int]string{
	errorInvalidJson: "invalid_json",
	errorValidation:  "validation_error",
	errorNotFound:    "not_found",
	errorBadRequest:  "bad_request",
	errorInternal:    "errorInternal",
}

func newError(ctx context.Context, errType ErrType, err error) *ErrorResponse {
	reqId, ok := ctx.Value(middleware.RequestId).(string)
	if !ok {
		reqId = ""
	}

	errResponse := &ErrorResponse{
		Error: ErrorInfo{
			Code:    codeMap[errType],
			Message: err.Error(),
			Details: []ErrorDetail{},
		},
		RequestId: reqId,
	}

	if errType == errorValidation {
		var errFields validator.ValidationErrors
		if !errors.As(err, &errFields) {
			return errResponse
		}

		details := make([]ErrorDetail, 0, len(errFields))
		for _, err := range errFields {
			details = append(details, ErrorDetail{
				Field:   err.StructField(),
				Rule:    err.ActualTag() + " " + err.Param(),
				Message: err.Error(),
			})
		}
		errResponse.Error.Details = details
	}

	return errResponse
}
