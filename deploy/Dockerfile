#BUILD
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o gophermart ./cmd/gophermart

#RUN
FROM alpine:3.20
COPY --from=builder /app//gophermart /gophermart
COPY --from=builder /app/internal/adapter/postgres/migrations /migrations

CMD ["/gophermart"]