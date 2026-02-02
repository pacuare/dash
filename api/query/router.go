package query

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"app.pacuare.dev/shared"
	"github.com/jackc/pgx/v5"
)

func queryEndpointArgs(w http.ResponseWriter, r *http.Request, query string, params []any) {
	email, err := shared.GetUser(r)

	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"Not authorized"}`))
		return
	}

	// e.g. keeper@farthergate.com -> user_keeper__farthergate_com
	var conn *pgx.Conn
	var dbName string
	if fullAccess, err := shared.QueryOne[bool]("select fullAccess from AuthorizedUsers where email=$1", email); err != nil {
		slog.Error("querying access level", "error", err)

		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Internal server error"}`))
		return
	} else if fullAccess {
		dbName = "pacuare_data"
	} else {
		dbName = shared.GetUserDatabase(*email)

	}

	dbUrl := fmt.Sprintf("%s/%s", os.Getenv("DATABASE_URL_BASE"), dbName)
	conn, err = pgx.Connect(r.Context(), dbUrl)

	if err != nil {
		slog.Error("open database", "error", err)

		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Internal server error"}`))
		return
	}

	defer conn.Close(r.Context())

	res, err := conn.Query(r.Context(), query, params...)
	if err != nil {
		slog.Error("run query", "query", query, "error", err)

		w.WriteHeader(500)
		jsonError, _ := json.Marshal(map[string]string{"error": err.Error()})
		w.Write(jsonError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("["))

	firstRow := true

	for res.Next() {
		if firstRow {
			firstRow = false
		} else {
			w.Write([]byte(","))
		}
		values, _ := pgx.RowToMap(res)
		marshalled, _ := json.Marshal(values)
		w.Write(marshalled)
	}
	w.Write([]byte("]"))
}

func Mount() {
	http.HandleFunc("/api/query", func(w http.ResponseWriter, r *http.Request) {
		if !(r.Method == http.MethodPost || r.Method == http.MethodOptions) {
			w.WriteHeader(405)
			fmt.Fprint(w, "Method not allowed")
			return
		}

		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(200)
			w.Write([]byte(`{"preflight": "ok"}`)) // send the preflight on its way before we hit logic
			return
		}

		var params []any
		var query string

		reqBuf := new(strings.Builder)

		_, err := io.Copy(reqBuf, r.Body)

		if err != nil {
			slog.Error("get query", "error", err)

			w.WriteHeader(500)
			w.Write([]byte(`{"error":"Internal server error"}`))
			return
		}

		if r.Header.Get("Content-Type") == "application/json" {
			var jsonBody struct {
				Query  string `json:"query"`
				Params []any  `json:"params"`
			}

			err = json.Unmarshal([]byte(reqBuf.String()), &jsonBody)

			if err != nil {
				slog.Error("parse json", "error", err)
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte(`{"error":"Parse error"}`))
				return
			}
			params = jsonBody.Params
			query = jsonBody.Query
		} else {
			params = []any{}
			query = reqBuf.String()
		}

		queryEndpointArgs(w, r, query, params)
	})
}
