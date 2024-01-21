FROM golang:1.21-alpine AS build

WORKDIR /usr/src/app

COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd

RUN apk add build-base
ARG CGO_ENABLED=1
RUN go mod download
RUN go build -ldflags="-s -w" -o /usr/local/bin/app cmd/globalshutdown/main.go

FROM alpine:3.19

COPY --from=build /usr/local/bin/app /app

CMD ["/app"]