# Dev Dockerfile
FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN go build -o /bin/todo-api ./cmd/todo-api

EXPOSE 8080
CMD ["/bin/todo-api"]