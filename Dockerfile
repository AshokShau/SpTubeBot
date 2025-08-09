FROM python:3.13-slim

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ffmpeg \
        bash \
        vorbis-tools \
        file \
        coreutils \
        gawk \
        xxd \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir uv

COPY . .

RUN chmod +x cover_gen.sh

RUN uv pip install -e . --system

CMD ["start"]
