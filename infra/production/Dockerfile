# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/todo-api ./cmd/todo-api

# Runtime stage
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /out/todo-api /todo-api

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/todo-api"]