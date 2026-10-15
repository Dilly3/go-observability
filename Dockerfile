FROM golang:1.27-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -o /out/wallet-api .

FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/wallet-api /app/wallet-api

USER 10001:10001

EXPOSE 8080

ENTRYPOINT ["/app/wallet-api"]