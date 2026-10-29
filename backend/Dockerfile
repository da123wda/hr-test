FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

COPY hr-server /app/hr-server

RUN chmod +x /app/hr-server && mkdir -p /app/data/uploads /app/config

EXPOSE 8080

CMD ["/app/hr-server"]
