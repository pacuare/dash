package auth

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"app.pacuare.dev/shared"
	"app.pacuare.dev/shared/mailer"
	"app.pacuare.dev/templates"
)

func Mount() {
	conn := shared.DB
	http.HandleFunc("GET /auth/login", func(w http.ResponseWriter, r *http.Request) {
		failedEmail := r.URL.Query().Get("failed-email")
		var pFailedEmail *string = nil

		if r.URL.Query().Has("failed-email") {
			pFailedEmail = &failedEmail
		}

		templates.Login(pFailedEmail).Render(r.Context(), w)
	})

	http.HandleFunc("GET /auth/verify", func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")

		hasUser, err := shared.QueryOne[bool]("select (count(*) > 0) from AuthorizedUsers where email=$1", email)

		if err != nil {
			slog.Error("verify", "error", err)
		}

		if !hasUser {
			http.Redirect(w, r, fmt.Sprintf("/auth/login?failed-email=%s", email), http.StatusSeeOther)
		} else {
			_, err := mailer.SendConfirmation(conn, email)

			if err != nil {
				slog.Error("create verification code", "email", email)
				http.Redirect(w, r, fmt.Sprintf("/auth/login?failed-email=%s", email), http.StatusSeeOther)
				return
			}

			templates.Verify(email).Render(r.Context(), w)
		}
	})

	http.HandleFunc("POST /auth/verify", func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")

		err := r.ParseForm()
		if err != nil {
			slog.Error("parse form", "form", "verify", "error", err)
			http.Redirect(w, r, fmt.Sprintf("/auth/login?failed-email=%s&login-failed", email), http.StatusSeeOther)
			return
		}

		inputCode := r.Form.Get("otp")

		slog.Info("otp input", "email", email, "code", inputCode)

		code, err := shared.QueryOne[string]("select code from LoginCodes where email=$1", email)

		if err != nil {
			slog.Error("get code", "error", err)
			http.Redirect(w, r, fmt.Sprintf("/auth/login?failed-email=%s&login-failed", email), http.StatusSeeOther)
			return
		}

		if strings.EqualFold(code, inputCode) {
			enc, err := shared.Encrypt([]byte(email))
			if err != nil {
				slog.Error("encrypt", "error", err)
				http.Redirect(w, r, fmt.Sprintf("/auth/login?failed-email=%s&login-failed", email), http.StatusSeeOther)
				return
			}

			c := &http.Cookie{Name: "AuthStatus", Value: hex.EncodeToString(enc), MaxAge: 259200, Path: "/", HttpOnly: true}
			http.SetCookie(w, c)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})

	http.HandleFunc("GET /auth/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "AuthStatus", Path: "/", Expires: time.Unix(0, 0)})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}
