#!/bin/bash
set -e

# Set password from environment variable (default: dragon20sided)
SIM_PASSWORD="${SIM_PASSWORD:-dragons!!1!}"
echo "sim:${SIM_PASSWORD}" | chpasswd

# Generate host keys if they don't exist
if [ ! -f /etc/ssh/ssh_host_ed25519_key ]; then
    ssh-keygen -t ed25519 -f /etc/ssh/ssh_host_ed25519_key -N ""
    ssh-keygen -t rsa -b 4096 -f /etc/ssh/ssh_host_rsa_key -N ""
fi

echo "=== Sim SSH Server Ready ==="
echo "Connect with: ssh sim@<host> -p 2222"
echo "Password: ${SIM_PASSWORD}"

# Start sshd in foreground
exec /usr/sbin/sshd -D -e
