# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY sim.go .
RUN CGO_ENABLED=0 go build -o sim sim.go

# Stage 2: Minimal SSH server
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends openssh-server && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir -p /run/sshd

# Lock root account: disable password login and set shell to nologin
RUN passwd -l root && \
    usermod -s /usr/sbin/nologin root

# Create the sim user (password set at runtime via entrypoint)
RUN useradd -m -s /usr/local/bin/run-sim.sh sim

# Copy compiled binary and scripts
COPY --from=builder /build/sim /opt/sim
COPY sshd_config /etc/ssh/sshd_config
COPY run-sim.sh /usr/local/bin/run-sim.sh
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /opt/sim /usr/local/bin/run-sim.sh /entrypoint.sh

EXPOSE 22

ENTRYPOINT ["/entrypoint.sh"]
