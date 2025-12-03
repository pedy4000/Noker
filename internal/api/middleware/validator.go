package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pedy4000/noker/internal/api/response"
	"github.com/pedy4000/noker/pkg/validate"
)

func Validate[T any](next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input T
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			response.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := validate.Do(input); err != nil {
			response.Error(w, validate.Format(err), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "Body", input)
		next(w, r.WithContext(ctx))
	}
}
