#BUILD
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# -s убирает таблицу символов (информацию о функциях и переменных)
# -w убирате DWARF (данные, которые используют отладчики (например, gdb))
# -trimpath убирает абсолютные пути из бинарника
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o gophermart ./cmd/gophermart

#RUN
FROM alpine:3.20
COPY --from=builder /app/gophermart /gophermart
# used embed.FS
# COPY --from=builder /app/internal/adapter/postgres/migrations /migrations 

CMD ["/gophermart"]