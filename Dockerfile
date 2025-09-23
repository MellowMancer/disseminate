FROM node:lts-alpine AS frontend-build

WORKDIR /app/frontend

COPY frontend/package.json frontend/yarn.lock ./
RUN corepack enable && yarn install --immutable

COPY frontend/ ./

COPY entrypoint.sh ./
RUN chmod +x entrypoint.sh &&\
    ./entrypoint.sh 



FROM golang:1.25-alpine AS backend-build

RUN apk add --no-cache git bash &&\
    go install github.com/air-verse/air@latest

WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN go build -o server main.go



FROM golang:1.25-alpine

RUN apk add --no-cache bash

WORKDIR /app/backend

COPY --from=backend-build /app/backend/ ./
COPY --from=frontend-build /app/frontend/dist ../frontend/dist

EXPOSE 8080

CMD ["./server"]
