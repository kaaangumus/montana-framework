#!/bin/sh
set -eu

INSTALL_DIR="/usr/local/share/montana"
BIN_DIR="/usr/local/bin"

if [ "$(id -u)" -ne 0 ]; then
  echo "Please run with sudo: sudo ./install.sh" >&2
  exit 1
fi

if [ ! -f "index.json" ]; then
  echo "index.json not found. Run this installer from the Montana repository root." >&2
  exit 1
fi

if [ ! -d "exploits" ]; then
  echo "exploits directory not found. Run this installer from the Montana repository root." >&2
  exit 1
fi

echo "Building montana..."
go build -o montana .

echo "Installing binary to $BIN_DIR/montana..."
mkdir -p "$BIN_DIR"
install -m 0755 montana "$BIN_DIR/montana"

echo "Installing metadata index to $INSTALL_DIR/index.json..."
mkdir -p "$INSTALL_DIR"
install -m 0644 index.json "$INSTALL_DIR/index.json"

echo "Installing exploit archive to $INSTALL_DIR/exploits..."
rm -rf "$INSTALL_DIR/exploits"
cp -R exploits "$INSTALL_DIR/exploits"
find "$INSTALL_DIR/exploits" -type d -exec chmod 0755 {} \;
find "$INSTALL_DIR/exploits" -type f -exec chmod 0644 {} \;

echo "Installation complete. Run: montana -q \"wordpress rce\""
