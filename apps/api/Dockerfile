FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY apps/api/go.mod apps/api/go.sum ./apps/api/

WORKDIR /src/apps/api
RUN go mod download

WORKDIR /src
COPY apps/api ./apps/api

WORKDIR /src/apps/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/server

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/api ./api

EXPOSE 8080

CMD ["./api"]
