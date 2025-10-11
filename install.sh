#!/bin/bash

# Create directories
INSTALL_DIR="/usr/local/share/montana-framework"
BIN_DIR="/usr/local/bin"

echo "Creating directories..."
mkdir -p $INSTALL_DIR
mkdir -p $BIN_DIR

# Build the application
echo "Building montana-framework..."
go build -o montana-framework .

# Install the binary
echo "Installing montana-framework binary to $BIN_DIR..."
cp montana-framework $BIN_DIR

# Install data files
echo "Installing data files to $INSTALL_DIR..."
cp -r exploits $INSTALL_DIR
cp index.json $INSTALL_DIR

echo "Installation complete. You can now run 'montana-framework' from your terminal."
