# syntax=docker/dockerfile:1.3-labs
FROM golang:1.18-alpine as go
WORKDIR /app
COPY . ./
RUN <<EOF
go mod download
go build -o amizone-api-server cmd/amizone-api-server/server.go
EOF

FROM alpine:latest as tailscale
WORKDIR /app
COPY . ./
ENV TSFILE=tailscale_1.26.1_amd64.tgz
RUN wget https://pkgs.tailscale.com/stable/${TSFILE} && tar xzf ${TSFILE} --strip-components=1
COPY . ./

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates iptables ip6tables
COPY --from=go /app/amizone-api-server ./
COPY --from=tailscale /app/tailscale ./
COPY --from=tailscale /app/tailscaled ./
RUN mkdir -p /var/run/tailscale /var/cache/tailscale /var/lib/tailscale

COPY <<EOF ./start.sh
#!/bin/sh
/app/tailscaled --state=/var/lib/tailscale/tailscaled.state --socket=/var/run/tailscale/tailscaled.sock &
/app/tailscale up --authkey="\${TAILSCALE_AUTHKEY}" --hostname=fly-app
/app/amizone-api-server
EOF
RUN chmod +x start.sh

ENV GRPC_GO_LOG_SEVERITY_LEVEL=info
ENV GRPC_GO_LOG_VERBOSITY_LEVEL=99

EXPOSE 8081

CMD ["./start.sh"]
