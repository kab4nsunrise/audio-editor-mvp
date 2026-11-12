# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY backend/go.mod ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o /audio-editor .

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ffmpeg ca-certificates
WORKDIR /app
COPY --from=builder /audio-editor .
COPY static/ ./static/
RUN mkdir -p uploads
EXPOSE 8080
CMD ["./audio-editor"]
