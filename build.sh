#!/bin/bash

# Create bin directory if it doesn't exist
mkdir -p bin

# Build for Linux (amd64)
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o bin/miner-linux-amd64 cmd/miner/main.go cmd/miner/client.go

# Build for Windows (amd64)
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o bin/miner-windows-amd64.exe cmd/miner/main.go cmd/miner/client.go

# Build for macOS (amd64)
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o bin/miner-macos-amd64 cmd/miner/main.go cmd/miner/client.go

# Build for macOS (arm64) - For M1/M2 Macs
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o bin/miner-macos-arm64 cmd/miner/main.go cmd/miner/client.go

echo "Build complete! Binaries are in the bin directory:"
ls -l bin/miner-* 