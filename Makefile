.PHONY: all server agent db
all: server agent db
server:
	cd cmd/server && go build -o server *.go

agent:
	cd cmd/agent && go build -o agent *.go

db:
	podman stop metrics || true
	podman rm metrics || true
	podman run --name metrics -p 5432:5432 -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} -d postgres:16.1
	until podman exec metrics pg_isready -h localhost -p 5432; do \
  		echo "waiting Postgres to be ready..."; \
  		sleep 1; \
  	done
	podman exec -e PGPASSWORD=${POSTGRES_PASSWORD} metrics psql -U postgres -c "CREATE DATABASE metrics;"
