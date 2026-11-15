FROM golang:1.26.11-bookworm

RUN groupadd --gid 1000 dev \
    && useradd --uid 1000 --gid dev --create-home --shell /bin/bash dev

ENV CGO_ENABLED=0
USER dev
