package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sebnow/chud/platform/constants"
	"github.com/sebnow/chud/platform/log"
)

const maxJSONBodySize = 1 << 20

func RespondJSON(ctx context.Context, w http.ResponseWriter, statusCode int, data any) {
	logger := log.FromContext(ctx)

	w.Header().Set(constants.HTTPHeaderContentType, constants.HTTPContentTypeJSON)
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

type errorResponse struct {
	Message string `json:"message"`
}

func RespondError(ctx context.Context, w http.ResponseWriter, statusCode int, err error) {
	logger := log.FromContext(ctx)
	if statusCode >= http.StatusInternalServerError {
		logger.Error().Err(err).Int("status", statusCode).Msg("Responding with error")
	} else {
		logger.Debug().Err(err).Int("status", statusCode).Msg("Responding with error")
	}

	w.Header().Set(constants.HTTPHeaderContentType, constants.HTTPContentTypeJSON)
	w.WriteHeader(statusCode)

	message := err.Error()
	if statusCode >= http.StatusInternalServerError {
		message = "Internal server error"
	}

	response := map[string]any{
		"status": statusCode,
		"error":  errorResponse{Message: message},
	}

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		logger.Error().Err(encodeErr).Msg("Failed to encode JSON error response")
	}
}

func DecodeJSONOrRespond(ctx context.Context, w http.ResponseWriter, r *http.Request, v any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBodySize))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		log.FromContext(ctx).Debug().Err(err).Msg("Invalid JSON body")
		if _, tooLarge := errors.AsType[*http.MaxBytesError](err); tooLarge {
			RespondError(ctx, w, http.StatusRequestEntityTooLarge, errors.New("żądanie jest za duże"))
			return false
		}
		RespondError(ctx, w, http.StatusBadRequest, errors.New("nieprawidłowe dane żądania"))
		return false
	}
	return true
}
