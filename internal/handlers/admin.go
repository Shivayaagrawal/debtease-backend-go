package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/logger"
	"fmt"
	"net/http"
)

func HandleMetrics(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Logger.Infow("Admin metrics requested",
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)

		hits := cfg.FileserverHits.Load()
		html := fmt.Sprintf(`
		<html>
		<body>
			<h1>Welcome, DebtEase Admin</h1>
			<p>DebtEase has been visited %d times!</p>
		</body>
		</html>`, hits)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))

		logger.Logger.Infow("Admin metrics served", "hits", hits)
	}
}
