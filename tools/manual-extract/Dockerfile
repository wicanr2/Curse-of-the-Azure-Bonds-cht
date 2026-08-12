FROM debian:bookworm-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends unar poppler-utils imagemagick ca-certificates \
 && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/bin/sh", "-c"]
