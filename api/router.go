package api

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"app.pacuare.dev/api/auth"
	"app.pacuare.dev/api/query"
	"app.pacuare.dev/shared"
)

//go:embed openapi.yml
var apiSpec string

func Mount() {
	auth.Mount()
	query.Mount()

	http.HandleFunc("/api/openapi.yml", func(w http.ResponseWriter, r *http.Request) {
		if !(r.Method == http.MethodGet || r.Method == http.MethodOptions) {
			w.WriteHeader(405)
			fmt.Fprint(w, "Method not allowed")
			return
		}
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Content-Type", "application/yaml")
		w.WriteHeader(200)
		fmt.Fprint(w, apiSpec)
	})

	http.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprint(w, `{"ok":"true"}`)
	})

	http.HandleFunc("POST /api/refresh", func(w http.ResponseWriter, r *http.Request) {
		email, err := shared.GetUser(r)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte("Not authorized"))
			return
		}

		db, err := shared.QueryOne[string]("select InitUserDatabase($1, $2, $3)", os.Getenv("DATABASE_URL_BASE"), os.Getenv("DATABASE_DATA"), *email)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
			return
		}

		w.WriteHeader(200)
		w.Write([]byte(db))
	})

	http.HandleFunc("POST /api/recreate", func(w http.ResponseWriter, r *http.Request) {
		email, err := shared.GetUser(r)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte("Not authorized"))
			return
		}

		_, err = shared.DB.Exec(context.Background(), fmt.Sprintf("drop database %s", shared.GetUserDatabase(*email)))

		if err != nil {
			slog.Error("recreate db", "user", email, "error", err)
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
			return
		}

		w.WriteHeader(200)
		w.Write([]byte("Deleted successfully"))
	})

	http.HandleFunc("POST /api/key", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			slog.Error("parse key form", "error", err)
			http.Redirect(w, r, "/?settings", http.StatusSeeOther)
			return
		}

		email, err := shared.GetUser(r)
		description := r.FormValue("description")

		if err != nil {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		key, err := shared.QueryOne[string]("insert into APIKeys (email, description) values ($1, $2) returning key", email, description)

		if err != nil {
			slog.Error("create api key", "error", err)
			http.Redirect(w, r, "/?settings", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/?settings&key=%s", key), http.StatusSeeOther)
	})

	http.HandleFunc("POST /api/key/delete", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			slog.Error("delete api key", "error", err)
			http.Redirect(w, r, "/?settings", http.StatusSeeOther)
			return
		}

		email, err := shared.GetUser(r)
		id := r.FormValue("id")

		if err != nil {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		_, err = shared.DB.Exec(context.Background(), "delete from APIKeys where id = $1 and email = $2", id, email)

		if err != nil {
			slog.Error("delete api key", "error", err)
			http.Redirect(w, r, "/?settings", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/?settings", http.StatusSeeOther)
	})
}
