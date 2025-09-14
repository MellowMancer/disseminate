# Only for development ( I really dont want to give up hot reloading )
FROM golang:1.25-alpine
RUN apk add --no-cache git bash
RUN go install github.com/air-verse/air@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]


