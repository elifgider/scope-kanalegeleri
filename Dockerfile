FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY go-app/go.mod go-app/go.sum ./go-app/
WORKDIR /src/go-app
RUN go mod download

WORKDIR /src
COPY go-app ./go-app
COPY public ./public

WORKDIR /src/go-app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/kanalegeleri-go .

FROM alpine:3.21
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/kanalegeleri-go /app/kanalegeleri-go
COPY --from=builder /src/go-app/config /app/config
COPY --from=builder /src/go-app/templates /app/templates
COPY --from=builder /src/public/static /app/public/static

RUN mkdir -p /app/uploads && chown -R app:app /app

USER app

EXPOSE 8080

CMD ["/app/kanalegeleri-go"]
