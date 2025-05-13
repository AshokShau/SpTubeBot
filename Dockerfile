FROM golang:1.24.0-alpine3.20 AS builder

WORKDIR /app

RUN apk add --no-cache git gcc musl-dev

COPY . .
RUN go build -ldflags="-w -s" -o songBot

FROM alpine:3.20.2

WORKDIR /app

RUN apk add --no-cache ffmpeg bash vorbis-tools file coreutils gawk

COPY --from=builder /app/songBot /app/songBot
COPY cover_gen.sh /app/cover_gen.sh
RUN chmod +x /app/songBot /app/cover_gen.sh

ENTRYPOINT ["/app/songBot"]
