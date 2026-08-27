FROM cd-access:dev-wine

RUN apt-get update \
    && apt-get install -y --no-install-recommends wine64 \
    && rm -rf /var/lib/apt/lists/*

ENTRYPOINT []
CMD ["/usr/lib/wine/wine64", "--version"]
