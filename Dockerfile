FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git libc-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy && go mod download
RUN swag init -generalInfo ./docs.go


COPY . .

RUN go build -o catalogd ./cmd/catalogd

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/catalogd .

EXPOSE 8080

CMD ["./catalogd"]
