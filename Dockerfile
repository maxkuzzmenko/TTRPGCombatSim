# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY dnd.go .
RUN CGO_ENABLED=0 go build -o dnd dnd.go

# Stage 2: Minimal SSH server
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends openssh-server && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir -p /run/sshd

# Lock root account: disable password login and set shell to nologin
RUN passwd -l root && \
    usermod -s /usr/sbin/nologin root

# Create the dnd user (password set at runtime via entrypoint)
RUN useradd -m -s /usr/local/bin/run-dnd.sh dnd

# Copy compiled binary and scripts
COPY --from=builder /build/dnd /opt/dnd
COPY sshd_config /etc/ssh/sshd_config
COPY run-dnd.sh /usr/local/bin/run-dnd.sh
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /opt/dnd /usr/local/bin/run-dnd.sh /entrypoint.sh

EXPOSE 22

ENTRYPOINT ["/entrypoint.sh"]
