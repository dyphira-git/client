# DYP Chain Miner

This is the miner client for the DYP blockchain. It connects to a mining server and performs proof-of-work mining to create new blocks.

## Pre-built Binaries

Pre-built binaries are available in the `bin` directory:
- Windows: `miner-windows-amd64.exe`
- Linux: `miner-linux-amd64`
- macOS (Intel): `miner-macos-amd64`
- macOS (M1/M2): `miner-macos-arm64`

## Running the Miner

### Windows
```cmd
miner-windows-amd64.exe -miner YOUR_WALLET_ADDRESS -server dyphira-chain.fly.dev:50051
```

### Linux
```bash
./miner-linux-amd64 -miner YOUR_WALLET_ADDRESS -server dyphira-chain.fly.dev:50051
```

### macOS (Intel)
```bash
./miner-macos-amd64 -miner YOUR_WALLET_ADDRESS -server dyphira-chain.fly.dev:50051
```

### macOS (M1/M2)
```bash
./miner-macos-arm64 -miner YOUR_WALLET_ADDRESS -server dyphira-chain.fly.dev:50051
```

Replace `YOUR_WALLET_ADDRESS` with your Ethereum-format wallet address (starting with 0x).

## Building from Source

### Prerequisites
- Go 1.24 or later
- Git

### Windows
```cmd
build.bat
```

### Linux/macOS
```bash
chmod +x build.sh
./build.sh
```

## Command Line Options

- `-miner`: Your wallet address (required)
- `-server`: Mining server address (default: localhost:50051)

Example:
```bash
./miner -miner 0x1234567890123456789012345678901234567890 -server dyphira-chain.fly.dev:50051
```

## Troubleshooting

1. **Connection Issues**
   - Verify the server address is correct
   - Check your internet connection
   - Ensure the port 50051 is not blocked by your firewall

2. **Invalid Wallet Address**
   - Make sure your wallet address is in the correct Ethereum format
   - It should start with "0x" followed by 40 hexadecimal characters

3. **Performance Issues**
   - The miner will automatically use available CPU cores
   - Mining speed depends on your CPU performance
   - Consider running multiple instances if you have multiple wallet addresses 