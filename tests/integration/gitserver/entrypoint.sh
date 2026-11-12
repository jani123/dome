#!/bin/bash
set -euo pipefail

: "${TEST_GIT_AUTHORIZED_KEY:?provide the test SSH public key}"
install -d -m 0755 /run/sshd /run/apache2 /run/git-test-server /srv/git
chown git:git /srv/git
install -d -m 0700 -o git -g git /home/git/.ssh
printf '%s\n' "no-agent-forwarding,no-port-forwarding,no-X11-forwarding,no-pty ${TEST_GIT_AUTHORIZED_KEY}" > /home/git/.ssh/authorized_keys
chmod 0600 /home/git/.ssh/authorized_keys
chown git:git /home/git/.ssh/authorized_keys
if [[ ! -f /run/git-test-server/ssh_host_ed25519_key ]]; then
    ssh-keygen -q -t ed25519 -N '' -f /run/git-test-server/ssh_host_ed25519_key
fi

/usr/sbin/sshd -D -e -f /etc/git-test-server/sshd_config &
ssh_pid=$!
/usr/sbin/apache2 -f /etc/git-test-server/httpd.conf -DFOREGROUND &
http_pid=$!
trap 'kill "$ssh_pid" "$http_pid" 2>/dev/null || true; wait || true' EXIT
trap 'exit 0' TERM INT
wait -n "$ssh_pid" "$http_pid"
