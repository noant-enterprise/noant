# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.26-alpine AS backend
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/main main.go

# Stage 3: Runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=backend /app/backend/bin/main ./backend/main
COPY --from=frontend /app/frontend/dist ./backend/static/
EXPOSE 8080
ENV GIN_MODE=release
ENTRYPOINT ["./backend/main"]
