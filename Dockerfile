FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN ln -snf /usr/share/zoneinfo/America/Sao_Paulo /etc/localtime && echo "America/Sao_Paulo" > /etc/timezone

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8084

CMD ["./main"]
