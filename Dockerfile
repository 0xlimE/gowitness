FROM ghcr.io/go-rod/rod

# Install runtime dependencies
RUN apt-get update && apt-get install -y libpcap0.8 ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the pre-built gowitness binary
COPY ./gowitness /usr/local/bin/gowitness
RUN chmod +x /usr/local/bin/gowitness

# Set working directory
WORKDIR /app

# Expose ports for gowitness server
EXPOSE 7171 8080

# Create volume mount point
VOLUME ["/app/targets"]

ENTRYPOINT ["dumb-init", "--", "gowitness"]