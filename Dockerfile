FROM alpine:3.24@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

# Install CA certificates for TLS
RUN apk add --no-cache ca-certificates

# Copy binary from goreleaser build context
ARG TARGETARCH
COPY linux/${TARGETARCH}/mcp-datahub /usr/local/bin/mcp-datahub

# Run as non-root user
RUN adduser -D -u 1000 mcp
USER mcp

ENTRYPOINT ["/usr/local/bin/mcp-datahub"]
