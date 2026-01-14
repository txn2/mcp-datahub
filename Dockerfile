FROM alpine:3.21

# Install CA certificates for TLS
RUN apk add --no-cache ca-certificates

# Copy binary from goreleaser build context
ARG TARGETARCH
COPY mcp-datahub /usr/local/bin/mcp-datahub

# Run as non-root user
RUN adduser -D -u 1000 mcp
USER mcp

ENTRYPOINT ["/usr/local/bin/mcp-datahub"]
