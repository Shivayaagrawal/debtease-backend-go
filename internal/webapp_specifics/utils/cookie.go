package utils

import (
	"net/http"

	"github.com/google/uuid"
)

const SessionCookieName = "calc_session"

func GetSessionID(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return uuid.Nil, err
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		return uuid.Nil, err
	}

	return sessionID, nil
}

func SetSessionCookie(w http.ResponseWriter, sessionID uuid.UUID) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID.String(),
		Path:     "/",
		MaxAge:   31536000, // 1 year
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // Set to true in production with HTTPS
	}
	http.SetCookie(w, cookie)
}
