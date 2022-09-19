FROM golang:latest as builder

ADD . /usr/src/playlist-bot
WORKDIR /usr/src/playlist-bot/cmd/discord
RUN go build -o /usr/bin/playlist-bot .

FROM debian:stable-slim
RUN apt-get update && apt-get install ffmpeg ca-certificates -y
COPY --from=builder /usr/bin/playlist-bot /usr/bin/playlist-bot

CMD ["/usr/bin/playlist-bot"]
