# Use the specified Ubuntu base image
FROM mirror.gcr.io/library/ubuntu:24.04

# Set environment variables for Go
ENV GO_VERSION=1.20 \
    GOROOT=/usr/local/go \
    GOPATH=/go \
    PATH=/usr/local/go/bin:/go/bin:$PATH

# Update the package list and install dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    curl \
    make \
    git \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Install Docker CLI
RUN curl -fsSL https://download.docker.com/linux/static/stable/x86_64/docker-20.10.24.tgz | tar -xzC /usr/local/bin --strip-components=1 docker/docker

# Download and install Go
RUN curl -fsSL https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz | tar -xzC /usr/local

# Create Go workspace directories
RUN mkdir -p /go/src /go/bin /go/pkg

# Verify installations
RUN go version && \
    docker --version && \
    make --version && \
    git --version

# Set default working directory
WORKDIR /app

# Default command
CMD ["bash"]
