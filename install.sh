#!/bin/sh
set -eu

INSTALL_DIR="/usr/local/share/montana"
BIN_DIR="/usr/local/bin"

echo "Building montana..."
go build -o montana .

echo "Installing binary to $BIN_DIR/montana..."
install -m 0755 montana "$BIN_DIR/montana"

echo "Installing metadata index to $INSTALL_DIR/index.json..."
mkdir -p "$INSTALL_DIR"
install -m 0644 index.json "$INSTALL_DIR/index.json"

echo "Installation complete. Run: montana -q \"wordpress rce\""
