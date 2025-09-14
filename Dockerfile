# React frontend
FROM node:lts-alpine as frontend-builder
WORKDIR /app
COPY app/package.json yarn.lock ./
RUN yarn install --immutable
COPY frontend/ .
RUN yarn build

# Go backend
FROM golang:1.25-alpine as backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN go build -o /server

# Stage 3: Production image
FROM nginx:stable-alpine
COPY --from=frontend-builder /app/build /usr/share/nginx/html
COPY --from=backend-builder /server /usr/bin/server
COPY nginx/default.conf /etc/nginx/conf.d/default.conf

# Expose port 80 for nginx and 8080 for backend
EXPOSE 80
EXPOSE 8080

# Run backend & nginx concurrently (use a process manager or script)
CMD sh -c "/usr/bin/server & nginx -g 'daemon off;'"