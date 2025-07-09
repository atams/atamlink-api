FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git libc-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy && go mod download

# Install swag binary (pastikan versinya sesuai kebutuhan Anda)
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

# Generate swagger docs
RUN /go/bin/swag init -generalInfo ./docs.go

# Build binary
RUN go build -o catalogd ./cmd/catalogd

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/catalogd .

EXPOSE 8080

CMD ["./catalogd"]
