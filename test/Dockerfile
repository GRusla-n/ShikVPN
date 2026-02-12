FROM ubuntu:22.04
RUN apt-get update && apt-get install -y \
    iproute2 iptables curl iputils-ping net-tools \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY build/vpn-server build/vpn-client build/vpn-keygen ./
COPY test/run-test.sh ./
RUN chmod +x vpn-server vpn-client vpn-keygen run-test.sh
