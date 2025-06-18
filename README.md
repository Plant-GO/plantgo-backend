# Project plantgo-backend

## Getting Started

### Prerequisites
- [Docker](https://www.docker.com/)
- [Make (optional)](https://www.gnu.org/software/make/)

## MakeFile

Run bckend on docker with containerized DB
```bash
make docker-run
```

Live reload the application:
```bash
make watch
```

Shutdown DB Container
```bash
make docker-down
```

DB Integrations Test:
```bash
make itest
```

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```

## Swagger UI for API Documentation

Test the API endpoints here :
```bash
http://localhost:8080/swagger/index.html
```

## Wanna interact with db ?

Connect to db
```bash
docker exec -it plantgo-backend-plantgo_postgres-1 psql -U gogo -d plantgo_db
```

## Swagger UI Docs

Endpoints interaction 
```bash
swag init -g ./cmd/api/main.go -o ./cmd/api/docs
```