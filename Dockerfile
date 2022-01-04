FROM golang:1.17.5-alpine3.15 AS build

WORKDIR /go/pickme

COPY *.go ./
COPY go.mod .
COPY go.sum .

RUN go build -o /go/bin/pickme

FROM alpine:3.15

COPY --from=build /go/bin/pickme /usr/local/bin/pickme

RUN \
  adduser -DH user \
  \
  && apk add --no-cache \
    ca-certificates

USER user

ENV \
  PICKME_REDISADDRS= \
  PICKME_TELEGRAM_PATH= \
  PICKME_TELEGRAM_PROXY= \
  PICKME_TELEGRAM_TOKEN= \
  PICKME_TELEGRAM_URL=

EXPOSE 8080

CMD ["/usr/local/bin/pickme"]
