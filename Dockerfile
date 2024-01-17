FROM golang:1.21-alpine AS build

WORKDIR /usr/src/app

COPY go.mod ./
COPY cmd ./cmd

RUN go build -ldflags="-s -w" -o /usr/local/bin/app cmd/globalshutdown/main.go

FROM alpine:3.19

COPY --from=build /usr/local/bin/app /app

CMD ["/app", "-server"]