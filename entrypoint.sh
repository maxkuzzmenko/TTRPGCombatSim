#!/bin/bash
set -e

# Set password from environment variable (default: dragon20sided)
DND_PASSWORD="${DND_PASSWORD:-dragon20sided}"
echo "dnd:${DND_PASSWORD}" | chpasswd

# Generate host keys if they don't exist
if [ ! -f /etc/ssh/ssh_host_ed25519_key ]; then
    ssh-keygen -t ed25519 -f /etc/ssh/ssh_host_ed25519_key -N ""
    ssh-keygen -t rsa -b 4096 -f /etc/ssh/ssh_host_rsa_key -N ""
fi

echo "=== DnD SSH Server Ready ==="
echo "Connect with: ssh dnd@<host> -p 2222"
echo "Password: ${DND_PASSWORD}"

# Start sshd in foreground
exec /usr/sbin/sshd -D -e
