printf "SECRET_KEY_BASE=%s\n" "$(openssl rand -hex 16)" > .env
cat >> .env <<EOF
DATABASE_URL=postgresql://localhost:5432/pacuare
DATABASE_URL_BASE=postgresql://localhost:5432
DATABASE_DATA=pacuare_data
RESEND_API_KEY=console
EOF
