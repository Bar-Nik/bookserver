package api

import (
	"log/slog"
	"net/http"
	"server/internal/logger"

	"github.com/gorilla/mux"
)

func Logging(log *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log = log.With(
				slog.String("ip", r.RemoteAddr),
				slog.String("URL", r.URL.Path),
			)

			ctx := logger.NewContext(r.Context(), log)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// func Logging(Logger *slog.Logger) mux.MiddlewareFunc {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			var metod string
// 			switch r.Method {
// 			case "GET":
// 				metod = "Get book(s)"
// 			case "DELETE":
// 				metod = "Delete book"
// 			case "POST":
// 				metod = "Post book(s)"
// 			case "PUT":
// 				metod = "Update book"
// 			}
// 			Logger.Info("Request",
// 				slog.String("msg", metod),
// 			)
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }
