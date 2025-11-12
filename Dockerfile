# --- Stage 1: Build frontend ---
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build   # SPA to ./dist

# --- Stage 2: Build Go binary ---
FROM golang:1.25-alpine AS go-builder
RUN apk add --no-cache build-base git

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . ./

RUN go build -trimpath -ldflags "-s -w" -o /binary

# --- Stage 3: Final image ---
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=go-builder /binary /binary
COPY --from=frontend-builder /frontend/dist /app/res/dist

EXPOSE 80

ENTRYPOINT ["/binary"]
