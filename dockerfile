FROM ubuntu:latest

RUN apt-get update && apt-get install -y \
    iptables \
    iproute2 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /veilnet

COPY ./veilnet-portal ./veilnet-portal

RUN chmod +x ./veilnet-portal

EXPOSE 3000

CMD ["./veilnet-portal"]