# PowerShell test script to verify Cultivator works

Write-Host "Testing Cultivator Build..." -ForegroundColor Green
Write-Host ""

# Build
Write-Host "Building..." -ForegroundColor Yellow
go build -o cultivator.exe ./cmd/cultivator

if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}

# Version
Write-Host ""
Write-Host "Version:" -ForegroundColor Cyan
./cultivator.exe version

# Validate
Write-Host ""
Write-Host "Validating config:" -ForegroundColor Cyan
./cultivator.exe validate

# Help
Write-Host ""
Write-Host "Help:" -ForegroundColor Cyan
./cultivator.exe --help

Write-Host ""
Write-Host "Build successful! Binary created: cultivator.exe" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Move to PATH: Move-Item cultivator.exe C:\Windows\System32\"
Write-Host "  2. Run: cultivator --help"
Write-Host "  3. Test in a Terragrunt repo"
