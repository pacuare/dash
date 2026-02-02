# Run templ generation in watch mode
templ:
	go tool templ generate --watch --proxy="http://localhost:8090" --open-browser=false -v

# Run air for Go hot reload
server:
	go tool air \
	    --build.cmd "go build -o tmp/bin/main ./main.go" \
        --build.bin "tmp/bin/main" \
        --build.delay "100" \
        --build.exclude_dir "node_modules" \
        --build.include_ext "go" \
        --build.stop_on_error "false" \
        --misc.clean_on_exit true

# Watch Tailwind CSS changes
tailwind:
	tailwindcss -i ./assets/css/input.css -o ./public/output.css --watch

# Start development server with all watchers
dev:
	make -j3 tailwind templ server

migrate:
	./goose postgres \"$DATABASE_URL\" up -dir migrations/
