FROM golang:1.21 as builder

WORKDIR /app

COPY . .

RUN go build -o funeypot .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

VOLUME /etc/funeypot

WORKDIR /data

COPY --from=builder /app/funeypot /usr/local/bin/funeypot

EXPOSE 22
EXPOSE 80
EXPOSE 21

ENTRYPOINT ["funeypot"]

CMD ["-c", "/etc/funeypot/funeypot.yaml"]
