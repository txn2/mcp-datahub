FROM alpine:3.23@sha256:865b95f46d98cf867a156fe4a135ad3fe50d2056aa3f25ed31662dff6da4eb62

# Install CA certificates for TLS
RUN apk add --no-cache ca-certificates

# Copy binary from goreleaser build context
ARG TARGETARCH
COPY linux/${TARGETARCH}/mcp-datahub /usr/local/bin/mcp-datahub

# Run as non-root user
RUN adduser -D -u 1000 mcp
USER mcp

ENTRYPOINT ["/usr/local/bin/mcp-datahub"]
