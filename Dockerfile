# ---- Stage 1: Build Backend ----
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app/backend

# Copy Go module files and download dependencies first
# This leverages Docker cache
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the rest of the backend source code
COPY backend/ ./

# Build the backend application
# Assuming the main package is in cmd/api
# If your main.go is directly in backend/, change the path accordingly
# Build the backend application
# Assuming the main package is in cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api/main.go
# Based on the file structure, main.go is in the root of backend/
# RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./main.go

# ---- Stage 2: Build Frontend ----
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy package.json and package-lock.json
COPY frontend/package.json frontend/package-lock.json ./

# Install dependencies (including devDependencies needed for build)
RUN npm ci

# Ensure scripts in node_modules/.bin are executable
RUN chmod -R +x node_modules/.bin

# Copy the rest of the frontend source code
COPY frontend/ ./

# Build the frontend application
RUN node node_modules/react-scripts/bin/react-scripts.js build

# ---- Stage 3: Final Image ----
FROM nginx:1.27-alpine

# Install gettext for envsubst used in entrypoint.sh
RUN apk add --no-cache gettext

# Copy the built backend binary from the backend-builder stage
COPY --from=backend-builder /app/api /app/api

# Copy the built frontend static files from the frontend-builder stage
COPY --from=frontend-builder /app/frontend/build /usr/share/nginx/html

# Copy the Nginx configuration file
COPY nginx.conf /etc/nginx/nginx.conf

# Copy the entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose the default Cloud Run port (though Cloud Run uses the PORT env var)
EXPOSE 8080

# Set the entrypoint script as the command to run
ENTRYPOINT ["/entrypoint.sh"]
