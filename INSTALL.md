# Installation Guide

Multiple ways to install Cultivator.

## Option 1: Download Pre-built Binary (Recommended)

### Linux / macOS

```bash
# Download latest release
curl -LO https://github.com/cultivator-dev/cultivator/releases/latest/download/cultivator-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)

# Make executable
chmod +x cultivator-*

# Move to PATH
sudo mv cultivator-* /usr/local/bin/cultivator

# Verify installation
cultivator version
```

### Windows (PowerShell)

```powershell
# Download latest release
Invoke-WebRequest -Uri "https://github.com/cultivator-dev/cultivator/releases/latest/download/cultivator-windows-amd64.exe" -OutFile "cultivator.exe"

# Add to PATH or move to a directory in PATH
Move-Item cultivator.exe C:\Windows\System32\

# Verify installation
cultivator version
```

## Option 2: Install via Go

```bash
go install github.com/cultivator-dev/cultivator/cmd/cultivator@latest
```

## Option 3: Build from Source

### Prerequisites
- Go 1.25 or higher
- Git

### Steps

```bash
# Clone repository
git clone https://github.com/cultivator-dev/cultivator.git
cd cultivator

# Build
make build

# Install
sudo make install

# Or just copy to PATH
sudo cp cultivator /usr/local/bin/

# Verify
cultivator version
```

## Option 4: Docker

### Pull Image

```bash
docker pull cultivator/cultivator:latest
```

### Run with Docker

```bash
docker run -v $(pwd):/workspace cultivator/cultivator:latest plan
```

### Build Docker Image

```bash
# Clone repository
git clone https://github.com/cultivator-dev/cultivator.git
cd cultivator

# Build image
docker build -t cultivator:local .

# Run
docker run -v $(pwd):/workspace cultivator:local plan
```

### Docker Compose

```bash
# Clone repository
git clone https://github.com/cultivator-dev/cultivator.git
cd cultivator

# Run with docker-compose
docker-compose run cultivator plan
```

## Option 5: GitHub Action (No Installation)

Use directly in GitHub Actions without local installation:

```yaml
- uses: cultivator-dev/cultivator-action@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
```

## Verify Installation

```bash
# Check version
cultivator version

# Check help
cultivator --help

# Run a test command
cultivator detect
```

## Update Cultivator

### Binary Installation

Re-download and replace the binary:

```bash
curl -LO https://github.com/cultivator-dev/cultivator/releases/latest/download/cultivator-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)
chmod +x cultivator-*
sudo mv cultivator-* /usr/local/bin/cultivator
```

### Go Installation

```bash
go install github.com/cultivator-dev/cultivator/cmd/cultivator@latest
```

### Docker

```bash
docker pull cultivator/cultivator:latest
```

## Dependencies

Cultivator requires these tools to be installed:

### Required
- **Git** - For detecting changes
- **Terragrunt** >= 0.55.0
- **Terraform** >= 1.7.0

### Optional
- **Docker** - For containerized execution
- **Make** - For build automation

### Install Dependencies

#### Terragrunt

```bash
# Linux
wget -q https://github.com/gruntwork-io/terragrunt/releases/download/v0.55.0/terragrunt_linux_amd64
chmod +x terragrunt_linux_amd64
sudo mv terragrunt_linux_amd64 /usr/local/bin/terragrunt

# macOS
brew install terragrunt

# Windows
choco install terragrunt
```

#### Terraform

```bash
# Linux
wget -q https://releases.hashicorp.com/terraform/1.7.0/terraform_1.7.0_linux_amd64.zip
unzip terraform_1.7.0_linux_amd64.zip
sudo mv terraform /usr/local/bin/

# macOS
brew install terraform

# Windows
choco install terraform
```

## Uninstall

### Binary Installation

```bash
sudo rm /usr/local/bin/cultivator
```

### Go Installation

```bash
rm $(go env GOPATH)/bin/cultivator
```

### Docker

```bash
docker rmi cultivator/cultivator:latest
```

## Troubleshooting

### Command not found

Ensure `/usr/local/bin` is in your PATH:

```bash
echo $PATH
export PATH="/usr/local/bin:$PATH"
```

### Permission denied

Make sure the binary is executable:

```bash
chmod +x /usr/local/bin/cultivator
```

### Dependencies not found

Install Terraform and Terragrunt first, then verify:

```bash
terraform version
terragrunt version
```

## Next Steps

- Follow the [Quick Start Guide](docs/quickstart.md)
- Configure [cultivator.yml](docs/configuration.md)
- Check [Examples](examples/)

## Get Help

- [Documentation](docs/)
- [GitHub Issues](https://github.com/cultivator-dev/cultivator/issues)
- [Contributing Guide](CONTRIBUTING.md)
