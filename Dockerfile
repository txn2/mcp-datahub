FROM alpine:3.23@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659

# Install CA certificates for TLS
RUN apk add --no-cache ca-certificates

# Copy binary from goreleaser build context
ARG TARGETARCH
COPY linux/${TARGETARCH}/mcp-datahub /usr/local/bin/mcp-datahub

# Run as non-root user
RUN adduser -D -u 1000 mcp
USER mcp

ENTRYPOINT ["/usr/local/bin/mcp-datahub"]
