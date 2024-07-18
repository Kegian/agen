############################
# STEP 1 build executable binary
############################

FROM golang:1.19-alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

RUN go install -tags 'postgres clickhouse' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/server main.go

############################
# STEP 2 build a small image
############################

FROM alpine

RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime && \
    echo Europe/Moscow /etc/timezone

WORKDIR /app

COPY migrations migrations

COPY --from=builder /go/bin/migrate /bin/migrate
COPY --from=builder /app/server /app/server

EXPOSE 8080

CMD ["./server"]
