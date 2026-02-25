#!/bin/bash
# Quick test script to verify Cultivator works

set -e

echo "Testing Cultivator Build..."
echo ""

# Build
echo "Building..."
go build -o cultivator ./cmd/cultivator

# Version
echo ""
echo "Version:"
./cultivator version

# Validate
echo ""
echo "Validating config:"
./cultivator validate

# Help
echo ""
echo "Help:"
./cultivator --help

echo ""
echo "Build successful! Binary created: ./cultivator"
echo ""
echo "Next steps:"
echo "  1. Install: sudo mv cultivator /usr/local/bin/"
echo "  2. Run: cultivator --help"
echo "  3. Test in a Terragrunt repo"
