package response

import (
	"encoding/json"
	"net/http"

	"github.com/pedy4000/noker/pkg/logger"
)

func JSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func Error(w http.ResponseWriter, msg string, code int) {
	JSON(w, code, map[string]string{"error": msg})
	logger.Debug("HTTP Error", "code", code, "msg", msg)
}
