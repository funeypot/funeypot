FROM golang:1.23 as builder

ARG VERSION=unknown

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -v -buildvcs=false -ldflags="-s -w -X main.Version=${VERSION}" -o funeypot

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

VOLUME /config

WORKDIR /data

COPY --from=builder /app/funeypot /usr/local/bin/funeypot

EXPOSE 22
EXPOSE 80
EXPOSE 21

ENTRYPOINT ["funeypot"]

CMD ["-c", "/config/funeypot.yaml"]
