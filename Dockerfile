FROM node:lts-alpine AS build-frontend

WORKDIR /app/frontend

COPY ./frontend/package.json ./frontend/yarn.lock ./
RUN corepack enable && yarn install --immutable

COPY ./frontend/ ./

COPY entrypoint.sh ./
RUN chmod +x entrypoint.sh &&\
    ./entrypoint.sh 



FROM golang:1.25-alpine AS build-backend

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY handlers/ ./handlers/
COPY middlewares/ ./middlewares/
COPY models/ ./models/
COPY repositories/ ./repositories/
COPY routes/ ./routes/
COPY services/ ./services/



COPY --from=build-frontend /app/frontend/dist ./frontend/dist

#   - APP_ENV=production: This tells main.go to enable the embedded file server and DISABLE the dev proxy.
#   - CGO_ENABLED=0: Creates a fully static binary that can run on a minimal image like Alpine.
RUN CGO_ENABLED=0 APP_ENV=production go build -o /server ./main.go


FROM alpine:3.18
RUN addgroup -S nonroot \
    && adduser -S appuser -G nonroot
USER appuser

COPY --from=build-backend /server /server

EXPOSE 8080

CMD ["/server"]