FROM jetpackio/devbox:latest
USER root
COPY docker-daemon.json /etc/docker/daemon.json
ARG CODEX_VERSION=0.162.0
RUN apt-get update && apt-get install -y --no-install-recommends \
    git procps build-essential pkg-config ca-certificates curl nodejs npm \
    && rm -rf /var/lib/apt/lists/*
ENV NPM_CONFIG_PREFIX=/home/devbox/.local/npm
ENV PATH="/home/devbox/.local/npm/bin:${PATH}"
RUN mkdir -p "$NPM_CONFIG_PREFIX" && chown devbox:devbox "$NPM_CONFIG_PREFIX"
USER devbox
RUN npm install -g "@openai/codex@${CODEX_VERSION}"
