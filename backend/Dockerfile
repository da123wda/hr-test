FROM golang:1.27-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/hr-server ./src

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

COPY --from=builder /app/hr-server /app/hr-server

RUN mkdir -p /app/data/uploads /app/config

EXPOSE 8080

CMD ["/app/hr-server"]
