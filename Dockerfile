# Stage 1: Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --legacy-peer-deps
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.23-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/internal/server/static ./internal/server/static
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /dashboard ./cmd/dashboard

# Stage 3: Final image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend /dashboard /usr/local/bin/dashboard
RUN mkdir -p /data
EXPOSE 80
CMD ["dashboard"]
