FROM golang:1.24.0-alpine3.20 AS builder

WORKDIR /app

RUN apk add --no-cache --virtual .build-deps \
    git \
    gcc \
    musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o songBot

RUN apk del .build-deps

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache \
    ffmpeg \
    bash \
    vorbis-tools \
    file \
    coreutils \
    gawk


COPY --from=builder /app/songBot /app/cover_gen.sh ./
RUN chmod +x /app/songBot /app/cover_gen.sh

ENTRYPOINT ["/app/songBot"]
