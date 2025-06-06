@echo off
echo Creating bin directory...
if not exist bin mkdir bin

echo Building for Windows (amd64)...
set GOOS=windows
set GOARCH=amd64
go build -o bin\miner-windows-amd64.exe cmd/miner/main.go cmd/miner/client.go

echo Build complete! Binary is in the bin directory.
dir bin\miner-windows-amd64.exe 