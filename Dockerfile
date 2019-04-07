FROM golang:1.12.2-alpine3.9 AS build

WORKDIR /go/pickme

COPY *.go ./
COPY go.mod .
COPY go.sum .
COPY vendor ./vendor

RUN go build -mod=vendor -o /go/bin/pickme

FROM alpine:3.9

COPY --from=build /go/bin/pickme /usr/local/bin/pickme
COPY etc /etc/

RUN \
  adduser -DH user \
  \
  && apk add --no-cache \
    ca-certificates

USER user

ENV \
  PICKME_REDIS_ADDRS= \
  PICKME_TELEGRAM_PATH= \
  PICKME_TELEGRAM_PROXY= \
  PICKME_TELEGRAM_TOKEN= \
  PICKME_TELEGRAM_URL=

EXPOSE 8080

CMD ["/usr/local/bin/pickme"]
