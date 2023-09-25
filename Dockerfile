FROM --platform=$TARGETPLATFORM golang:latest as builder
ARG TARGETOS
ARG TARGETARCH

ADD . /usr/src/playlist-bot
WORKDIR /usr/src/playlist-bot
RUN --mount=type=cache,target=/go/pkg/mod \
      --mount=type=bind,source=go.mod,target=go.mod \
      --mount=type=bind,source=go.sum,target=go.sum \
      go mod download -x
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /usr/bin/playlist-bot .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /usr/bin/dca internal/dca/dca.go

FROM --platform=$TARGETPLATFORM debian:stable-slim
RUN apt-get update && apt-get install ffmpeg ca-certificates python3 wget -y
RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/bin/yt-dlp
RUN chmod +x /usr/bin/yt-dlp
COPY --from=builder /usr/bin/playlist-bot /usr/bin/playlist-bot
COPY --from=builder /usr/bin/dca /usr/bin/dca
RUN sha1sum /usr/bin/playlist-bot > /metadata-sha.sum
ADD web web

CMD ["/usr/bin/playlist-bot"]
