#!/bin/sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CLIENTS_DIR="$SCRIPT_DIR/clients"
UI_DIR="$CLIENTS_DIR/shock-ui"
EMBED_DIR="$SCRIPT_DIR/shock-server/ui/dist"

echo "==> Installing dependencies..."
cd "$CLIENTS_DIR"
npm install

echo "==> Building shock-ts..."
cd "$CLIENTS_DIR/shock-ts"
npm run build

echo "==> Building shock-ui..."
cd "$UI_DIR"
npm run build

echo "==> Copying dist to shock-server/ui/dist..."
rm -rf "$EMBED_DIR"
cp -r "$UI_DIR/dist" "$EMBED_DIR"

echo "==> UI build complete. Run ./compile-server.sh to embed in binary."
