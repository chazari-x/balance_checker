FROM golang:1.18 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.10

RUN adduser -DH ledger

WORKDIR /app

COPY --from=builder /app/main /app/

COPY ./config/config.docker.yaml /app/config/config.yaml
COPY ./files/proxy.txt /app/files/proxy.txt
RUN mkdir -p /app/files && chown -R ledger:ledger /app/files
RUN chown ledger:ledger /app/main
RUN chmod +x /app/main

USER ledger

CMD ["/app/main"]
