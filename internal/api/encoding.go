package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"example.com/permission-selector/internal/domain"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, domain.ErrNotFound) {
		status = http.StatusNotFound
	} else if errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrDuplicate) {
		status = http.StatusConflict
	} else if errors.Is(err, domain.ErrInactiveRecord) {
		status = http.StatusGone
	}
	writeJSON(writer, status, errorResponse{Error: err.Error()})
}

func decodeBody(request *http.Request, target any) error {
	if request.Body == nil {
		return domain.ErrInvalidInput
	}
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func pathValue(path, prefix string) string {
	if len(path) <= len(prefix) || path[:len(prefix)] != prefix {
		return ""
	}
	return path[len(prefix):]
}
