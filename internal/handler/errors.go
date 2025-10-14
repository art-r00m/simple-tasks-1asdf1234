package handler

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"simple-tasks/internal/middleware"
)

type errorDetail struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type errorInfo struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []errorDetail `json:"details,omitempty"`
}

type errorResponse struct {
	Error     errorInfo `json:"error"`
	RequestId string    `json:"requestId"`
}

type errType = int

const (
	errorInvalidJson errType = iota
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

func newError(ctx context.Context, errType errType, err error) *errorResponse {
	reqId, ok := ctx.Value(middleware.RequestId).(string)
	if !ok {
		reqId = ""
	}

	errResponse := &errorResponse{
		Error: errorInfo{
			Code:    codeMap[errType],
			Message: err.Error(),
			Details: []errorDetail{},
		},
		RequestId: reqId,
	}

	if errType == errorValidation {
		var errFields validator.ValidationErrors
		if !errors.As(err, &errFields) {
			return errResponse
		}

		details := make([]errorDetail, 0, len(errFields))
		for _, err := range errFields {
			details = append(details, errorDetail{
				Field:   err.StructField(),
				Rule:    err.ActualTag() + " " + err.Param(),
				Message: err.Error(),
			})
		}
		errResponse.Error.Details = details
	}

	return errResponse
}
