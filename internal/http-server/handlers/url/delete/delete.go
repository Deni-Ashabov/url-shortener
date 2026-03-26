package delete

import (
	"log/slog"
	"net/http"
	"url_shorter/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type DeleteURL interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, deleteUrl DeleteURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			http.Error(w, "alias is required", http.StatusBadRequest)
			return
		}

		err := deleteUrl.DeleteURL(alias)

		if err != nil {
			logger.Error("failed to delete url", sl.Err(err))

			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		logger.Info("deleting", slog.String("alias", alias))

		w.WriteHeader(http.StatusNoContent)
	}
}
