# Development environment setup script for Cultivator (PowerShell)
# This script sets up pre-commit hooks and development dependencies

$ErrorActionPreference = "Stop"

# Helper functions
function Print-Success {
    param($message)
    Write-Host "✓ $message" -ForegroundColor Green
}

function Print-Error {
    param($message)
    Write-Host "✗ $message" -ForegroundColor Red
}

function Print-Info {
    param($message)
    Write-Host "ℹ $message" -ForegroundColor Yellow
}

function Check-Command {
    param($command)
    $exists = Get-Command $command -ErrorAction SilentlyContinue
    if ($exists) {
        Print-Success "$command is installed"
        return $true
    } else {
        Print-Error "$command is not installed"
        return $false
    }
}

# Main setup
Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "  Cultivator Development Environment Setup" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Check required tools
Print-Info "Checking required tools..."
Write-Host ""

$missingTools = 0

# Check Go
if (Check-Command "go") {
    $goVersion = (go version)
    Write-Host "  Version: $goVersion"
} else {
    $missingTools++
    Write-Host "  Install: https://golang.org/dl/"
}
Write-Host ""

# Check Git
if (Check-Command "git") {
    $gitVersion = (git --version)
    Write-Host "  Version: $gitVersion"
} else {
    $missingTools++
    Write-Host "  Install: https://git-scm.com/download/win"
}
Write-Host ""

# Exit if required tools are missing
if ($missingTools -gt 0) {
    Print-Error "Missing $missingTools required tool(s). Please install them and try again."
    exit 1
}

Print-Success "All required tools are installed!"
Write-Host ""

# Install optional tools
Print-Info "Checking optional tools..."
Write-Host ""

# Check Python (for pre-commit)
$pythonCmd = $null
if (Get-Command "python" -ErrorAction SilentlyContinue) {
    $pythonCmd = "python"
} elseif (Get-Command "python3" -ErrorAction SilentlyContinue) {
    $pythonCmd = "python3"
} elseif (Get-Command "py" -ErrorAction SilentlyContinue) {
    $pythonCmd = "py"
}

if ($null -eq $pythonCmd) {
    Print-Error "Python is not installed (required for pre-commit)"
    Write-Host "  Install: https://www.python.org/downloads/"
    Write-Host "  Or use: winget install Python.Python.3"
    exit 1
} else {
    Print-Success "Python is installed"
    $pythonVersion = (& $pythonCmd --version)
    Write-Host "  Version: $pythonVersion"
}
Write-Host ""

# Install pre-commit
Print-Info "Installing pre-commit..."
if (Check-Command "pre-commit") {
    $precommitVersion = (pre-commit --version)
    Write-Host "  Already installed: $precommitVersion"
} else {
    try {
        & $pythonCmd -m pip install --user pre-commit
        Print-Success "pre-commit installed successfully"
        
        # Add to PATH if needed
        $userScriptsPath = & $pythonCmd -c "import site; print(site.USER_BASE + '\\Scripts')"
        if ($env:PATH -notlike "*$userScriptsPath*") {
            Print-Info "Adding pre-commit to PATH for this session..."
            $env:PATH = "$userScriptsPath;$env:PATH"
            
            Print-Info "To make this permanent, add to User PATH in System Environment Variables:"
            Write-Host "  $userScriptsPath" -ForegroundColor Cyan
        }
    } catch {
        Print-Error "Failed to install pre-commit"
        Write-Host "  Error: $_"
        exit 1
    }
}
Write-Host ""

# Check golangci-lint
Print-Info "Checking golangci-lint..."
if (Check-Command "golangci-lint") {
    $golangciVersion = (golangci-lint --version)
    Write-Host "  Version: $golangciVersion"
} else {
    Print-Info "golangci-lint not found. Installing..."
    try {
        # Download and install golangci-lint
        $golangciVersion = "v1.55.2"
        $downloadUrl = "https://github.com/golangci/golangci-lint/releases/download/$golangciVersion/golangci-lint-$golangciVersion-windows-amd64.zip"
        $tempZip = "$env:TEMP\golangci-lint.zip"
        $goPath = (go env GOPATH)
        $binPath = Join-Path $goPath "bin"
        
        # Create bin directory if it doesn't exist
        if (!(Test-Path $binPath)) {
            New-Item -ItemType Directory -Path $binPath | Out-Null
        }
        
        # Download
        Print-Info "Downloading golangci-lint..."
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempZip
        
        # Extract
        Expand-Archive -Path $tempZip -DestinationPath $env:TEMP -Force
        $extractedDir = Join-Path $env:TEMP "golangci-lint-$golangciVersion-windows-amd64"
        $exePath = Join-Path $extractedDir "golangci-lint.exe"
        
        # Move to GOPATH/bin
        Copy-Item $exePath -Destination (Join-Path $binPath "golangci-lint.exe") -Force
        
        # Cleanup
        Remove-Item $tempZip -Force
        Remove-Item $extractedDir -Recurse -Force
        
        Print-Success "golangci-lint installed successfully"
        
        # Add GOPATH/bin to PATH if not already there
        if ($env:PATH -notlike "*$binPath*") {
            Print-Info "Adding GOPATH/bin to PATH for this session..."
            $env:PATH = "$binPath;$env:PATH"
            
            Print-Info "To make this permanent, add to User PATH in System Environment Variables:"
            Write-Host "  $binPath" -ForegroundColor Cyan
        }
    } catch {
        Print-Error "Failed to install golangci-lint"
        Write-Host "  Error: $_"
        Write-Host "  Manual install: https://golangci-lint.run/usage/install/"
    }
}
Write-Host ""

# Download Go dependencies
Print-Info "Downloading Go dependencies..."
try {
    go mod download
    go mod tidy
    Print-Success "Go dependencies downloaded"
} catch {
    Print-Error "Failed to download Go dependencies"
    Write-Host "  Error: $_"
    exit 1
}
Write-Host ""

# Install pre-commit hooks
Print-Info "Installing pre-commit hooks..."
try {
    pre-commit install --install-hooks
    pre-commit install --hook-type commit-msg
    Print-Success "Pre-commit hooks installed"
} catch {
    Print-Error "Failed to install pre-commit hooks"
    Write-Host "  Error: $_"
    Write-Host "  Try running: pre-commit install --install-hooks"
    exit 1
}
Write-Host ""

# Verify setup
Print-Info "Verifying setup..."
Write-Host ""

try {
    $testBinary = Join-Path $env:TEMP "cultivator-test.exe"
    go build -o $testBinary .\cmd\cultivator
    if (Test-Path $testBinary) {
        Print-Success "Build test passed"
        Remove-Item $testBinary -Force
    }
} catch {
    Print-Error "Build test failed"
    Write-Host "  Error: $_"
    exit 1
}
Write-Host ""

# Summary
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "  Setup Complete!" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:"
Write-Host "  1. Run 'make check' or 'go test ./...' to verify everything works"
Write-Host "  2. Read CONTRIBUTING.md for development guidelines"
Write-Host "  3. Start coding! Pre-commit hooks will run automatically"
Write-Host ""
Write-Host "Useful commands:"
Write-Host "  make build           - Build the binary"
Write-Host "  make test            - Run tests"
Write-Host "  make check           - Run all checks (fmt, vet, lint, test)"
Write-Host "  make pre-commit-run  - Manually run pre-commit hooks"
Write-Host "  make help            - Show all available commands"
Write-Host ""
Print-Success "Happy coding!"
