package httpx

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
)

func PathUUIDOrRespond(w http.ResponseWriter, r *http.Request, key string) (string, bool) {
	value := r.PathValue(key)
	if uuid.Validate(value) != nil {
		RespondError(r.Context(), w, http.StatusNotFound, errors.New("nie znaleziono"))
		return "", false
	}
	return value, true
}
