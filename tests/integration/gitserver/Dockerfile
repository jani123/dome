FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends git apache2 openssh-server tini \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home --shell /bin/sh git \
    && passwd -d git
COPY httpd.conf sshd_config /etc/git-test-server/
COPY entrypoint.sh /usr/local/bin/git-test-server
RUN chmod 0755 /usr/local/bin/git-test-server
EXPOSE 8080 2222
ENTRYPOINT ["/usr/bin/tini", "-g", "--", "/usr/local/bin/git-test-server"]
