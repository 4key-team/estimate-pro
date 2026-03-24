FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /server /server
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations /migrations
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/server"]
