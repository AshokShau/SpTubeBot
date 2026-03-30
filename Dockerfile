FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go generate

RUN CGO_ENABLED=0 GOOS=linux go build -o bot main.go

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    ca-certificates \
    zlib1g \
    tzdata \
    curl \
    wget \
    && rm -rf /var/lib/apt/lists/*


RUN wget -O /usr/local/bin/yt-dlp \
    https://github.com/yt-dlp/yt-dlp-nightly-builds/releases/latest/download/yt-dlp_linux \
    && chmod +x /usr/local/bin/yt-dlp

RUN curl -fsSL https://deno.land/install.sh | sh \
    && ln -sf /root/.deno/bin/deno /usr/local/bin/deno

ENV DENO_INSTALL="/root/.deno"
ENV PATH="${DENO_INSTALL}/bin:${PATH}"

COPY --from=builder /app/bot .
COPY --from=builder /app/libtdjson.so.* ./

CMD ["./bot"]
