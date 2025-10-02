
# --- Stage 1: Build the Frontend ---
# This stage uses a Node.js environment to build the static frontend assets.
FROM node:lts-alpine AS build-frontend

WORKDIR /app/frontend

# Copy package manifests and install dependencies. This step is cached by Docker.
COPY ./frontend/package.json ./frontend/yarn.lock ./
RUN corepack enable && yarn install --immutable

# Copy the rest of the frontend source code.
COPY ./frontend/ ./

# Run the production build script from your package.json (e.g., "tsc -b && vite build").
# This creates the optimized static files in the /app/frontend/dist/ directory.
COPY entrypoint.sh ./
RUN chmod +x entrypoint.sh &&\
    ./entrypoint.sh 
# At the end of this stage, the Node.js environment is discarded, but the /app/frontend/dist folder is kept for the next stage.


# --- Stage 2: Build the Backend ---
# This stage uses a Go environment to compile the backend application.
FROM golang:1.25-alpine AS build-backend

WORKDIR /app

# Copy Go module files and download dependencies. This step is cached by Docker.
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire backend source code.
COPY . .

# --- The Magic Step ---
# Copy the compiled frontend assets from the 'build-frontend' stage into this stage.
COPY --from=build-frontend /app/frontend/dist ./frontend/dist

# Compile the Go application into a single, static binary.
# THIS IS THE MOST IMPORTANT LINE FOR CI/CD:
#   - APP_ENV=production: This tells your main.go to enable the embedded file server and DISABLE the dev proxy.
#   - CGO_ENABLED=0: Creates a fully static binary that can run on a minimal image like Alpine.
RUN CGO_ENABLED=0 APP_ENV=production go build -o /server ./main.go
# At the end of this stage, the Go compiler environment is discarded, but the final '/server' binary is kept.


# --- Stage 3: Final Production Image ---
# This stage creates the final, minimal image that you will deploy.
FROM alpine:3.18

# Copy ONLY the compiled server binary from the 'build-backend' stage.
# The final image does not contain Node.js, the Go compiler, or any source code.
COPY --from=build-backend /server /server

# Expose the port that our Go application listens on.
EXPOSE 8080

# The command to run the application when a container is started from this image.
CMD ["/server"]