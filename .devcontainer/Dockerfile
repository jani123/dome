FROM jetpackio/devbox:latest
USER root
COPY docker-daemon.json /etc/docker/daemon.json
ARG CODEX_VERSION=0.162.0
ARG HUNK_VERSION=0.25.0
ARG BROOT_VERSION=1.62.0
ARG BROOT_SHA256=13630ae5d3b011ddc81fca05cb80dbee6189dad3ba393dd477f5f687c7b5106d
RUN apt-get update && apt-get install -y --no-install-recommends \
    git procps build-essential pkg-config ca-certificates curl nodejs npm bat unzip util-linux \
    && ln -s /usr/bin/batcat /usr/local/bin/bat \
    && rm -rf /var/lib/apt/lists/*
RUN curl -fsSL "https://raw.githubusercontent.com/modem-dev/hunk/v${HUNK_VERSION}/install.sh" -o /tmp/install-hunk.sh \
    && HUNK_INSTALL_DIR=/usr/local/bin HUNK_NO_MODIFY_PATH=1 sh /tmp/install-hunk.sh "${HUNK_VERSION}" \
    && rm /tmp/install-hunk.sh \
    && hunk --version
RUN case "$(dpkg --print-architecture)" in \
        amd64) broot_target=x86_64-unknown-linux-musl ;; \
        arm64) broot_target=aarch64-unknown-linux-musl ;; \
        *) echo "Unsupported broot architecture" >&2; exit 1 ;; \
    esac \
    && curl -fsSL "https://github.com/Canop/broot/releases/download/v${BROOT_VERSION}/broot_${BROOT_VERSION}.zip" -o /tmp/broot.zip \
    && echo "${BROOT_SHA256}  /tmp/broot.zip" | sha256sum -c - \
    && unzip -p /tmp/broot.zip "${broot_target}/broot" > /usr/local/bin/broot \
    && chmod 0755 /usr/local/bin/broot \
    && rm /tmp/broot.zip \
    && broot --version
ENV NPM_CONFIG_PREFIX=/home/devbox/.local/npm
ENV PATH="/home/devbox/.local/npm/bin:${PATH}"
RUN mkdir -p "$NPM_CONFIG_PREFIX" && chown devbox:devbox "$NPM_CONFIG_PREFIX"
USER devbox
RUN npm install -g "@openai/codex@${CODEX_VERSION}"
