FROM golang:1.23.0 AS build
WORKDIR /
COPY . .

RUN go build -o /app /cmd/server/main.go

FROM ubuntu:20.04

ENV GIN_MODE release

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app /app
ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO /zoneinfo.zip
ENTRYPOINT [ "/app" ]
