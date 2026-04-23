FROM alpine:3.23@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11

# Install CA certificates for TLS
RUN apk add --no-cache ca-certificates

# Copy binary from goreleaser build context
ARG TARGETARCH
COPY linux/${TARGETARCH}/mcp-datahub /usr/local/bin/mcp-datahub

# Run as non-root user
RUN adduser -D -u 1000 mcp
USER mcp

ENTRYPOINT ["/usr/local/bin/mcp-datahub"]
