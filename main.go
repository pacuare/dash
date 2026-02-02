package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"app.pacuare.dev/api"
	"app.pacuare.dev/shared"
	"app.pacuare.dev/templates"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	_, json := os.LookupEnv("LOG_FORMAT_JSON")
	if json {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}

	slog.Info("Connecting to database")
	conn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database")
	}
	slog.Info("Connected")
	shared.DB = conn
	defer conn.Close()

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	api.Mount()
	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		returnedApiKey := r.URL.Query().Get("key")
		email, err := shared.GetUser(r)

		if err != nil {
			slog.Error("get user", "error", err)
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		fullAccess, err := shared.QueryOne[bool]("select fullAccess from AuthorizedUsers where email=$1", email)

		if err != nil {
			slog.Error("get access", "email", email, "error", err)
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		if !fullAccess {
			if databaseExists, err :=
				shared.QueryOne[bool]("select count(*)>0 from pg_catalog.pg_database where datname = GetUserDatabase($1)", email); !databaseExists || err != nil {
				http.Redirect(w, r, "/createdb", http.StatusSeeOther)
				if err != nil {
					slog.Error("failed to check database existence", "error", err)
				}
				return
			}
		}

		templates.Index(*email, fullAccess, returnedApiKey).Render(r.Context(), w)
	})

	http.HandleFunc("GET /createdb", func(w http.ResponseWriter, r *http.Request) {
		email, err := shared.GetUser(r)

		if err != nil {
			slog.Error("get user", "error", err)
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		if databaseExists, _ :=
			shared.QueryOne[bool]("select count(*)>0 from pg_catalog.pg_database where datname = GetUserDatabase($1)", email); databaseExists {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		templates.CreateDB(*email).Render(r.Context(), w)
		_, err = shared.DB.Exec(context.Background(), fmt.Sprintf("create database %s", shared.GetUserDatabase(*email)))

		if err != nil {
			slog.Error("create user database", "email", email, "error", err)
		}

		db, err := shared.QueryOne[string]("select InitUserDatabase($1, $2, $3)", os.Getenv("DATABASE_URL_BASE"), os.Getenv("DATABASE_DATA"), *email)

		if err != nil {
			slog.Error("init user database", "email", email, "error", err)
		}

		slog.Info("created database", "email", email, "db", db)

		w.Write([]byte("<meta http-equiv=\"Refresh\" content=\"0;url=/\">"))
	})

	slog.Info("Starting server on :8080")
	http.ListenAndServe(":8080", nil)
}
