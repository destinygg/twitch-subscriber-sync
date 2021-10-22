FROM golang:latest AS builder

RUN mkdir /build
COPY go.mod /build
COPY go.sum /build

COPY ./internal /build/internal
COPY ./twitchscrape /build/twitchscrape

WORKDIR /build/twitchscrape

RUN go get
RUN go build


# runner
FROM debian:buster-slim

RUN mkdir /app
WORKDIR /app/
COPY --from=builder /build/twitchscrape/twitchscrape .

CMD ["./twitchscrape"]