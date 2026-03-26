package redirect

import (
	"errors"
	"log/slog"
	"net/http"
	"url_shorter/internal/lib/logger/sl"
	"url_shorter/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type GetURL interface {
	GetUrl(alias string) (string, error)
}

func New(log *slog.Logger, getUrl GetURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			http.Error(w, "alias is required", http.StatusBadRequest)
			return
		}

		url, err := getUrl.GetUrl(alias)

		if errors.Is(err, storage.ErrUrlNotFound) {
			logger.Info("url not found", slog.String("alias", alias))

			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if err != nil {
			logger.Error("failed to get url", sl.Err(err))

			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		logger.Info("redirecting", slog.String("alias", alias))

		http.Redirect(w, r, url, http.StatusFound)
	}
}
