FROM jetpackio/devbox:latest
USER root
COPY docker-daemon.json /etc/docker/daemon.json
ARG CODEX_VERSION=0.162.0
ARG HUNK_VERSION=0.25.0
RUN apt-get update && apt-get install -y --no-install-recommends \
    git procps build-essential pkg-config ca-certificates curl nodejs npm bat \
    && ln -s /usr/bin/batcat /usr/local/bin/bat \
    && rm -rf /var/lib/apt/lists/*
RUN curl -fsSL "https://raw.githubusercontent.com/modem-dev/hunk/v${HUNK_VERSION}/install.sh" -o /tmp/install-hunk.sh \
    && HUNK_INSTALL_DIR=/usr/local/bin HUNK_NO_MODIFY_PATH=1 sh /tmp/install-hunk.sh "${HUNK_VERSION}" \
    && rm /tmp/install-hunk.sh \
    && hunk --version
ENV NPM_CONFIG_PREFIX=/home/devbox/.local/npm
ENV PATH="/home/devbox/.local/npm/bin:${PATH}"
RUN mkdir -p "$NPM_CONFIG_PREFIX" && chown devbox:devbox "$NPM_CONFIG_PREFIX"
USER devbox
RUN npm install -g "@openai/codex@${CODEX_VERSION}"
