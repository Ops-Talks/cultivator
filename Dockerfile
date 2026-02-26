# Dockerfile for Cultivator

FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cultivator ./cmd/cultivator

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates git bash

# Install OpenTofu (open-source alternative to Terraform)
ENV OPENTOFU_VERSION=1.6.1
RUN wget -q https://github.com/opentofu/opentofu/releases/download/v${OPENTOFU_VERSION}/tofu_${OPENTOFU_VERSION}_linux_amd64.zip && \
    unzip tofu_${OPENTOFU_VERSION}_linux_amd64.zip && \
    mv tofu /usr/local/bin/ && \
    rm tofu_${OPENTOFU_VERSION}_linux_amd64.zip && \
    ln -s /usr/local/bin/tofu /usr/local/bin/terraform

# Install Terragrunt
ENV TERRAGRUNT_VERSION=0.55.0
RUN wget -q https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64 && \
    chmod +x terragrunt_linux_amd64 && \
    mv terragrunt_linux_amd64 /usr/local/bin/terragrunt

# Copy binary from builder
COPY --from=builder /app/cultivator /usr/local/bin/cultivator

# Create working directory
WORKDIR /workspace

ENTRYPOINT ["cultivator"]
CMD ["--help"]
