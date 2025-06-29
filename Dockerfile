# Build stage (Go app)
FROM golang:1.24.3-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/api/main.go

# Prod stage (Debian-based with Python)
FROM python:3.11-slim AS prod

WORKDIR /app

# Install TensorFlow and Pillow
RUN pip install --no-cache-dir tensorflow pillow

# Copy the Go binary from the build stage
COPY --from=build /app/main .

# Copy Python scripts and models
COPY ml/ ./ml

EXPOSE ${PORT}

CMD ["./main"]
