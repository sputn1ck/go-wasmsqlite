#!/bin/bash
set -euo pipefail

# Configuration. By default, fetch the latest official SQLite WASM bundle from
# the script-friendly metadata embedded in sqlite.org/download.html.
ASSETS_DIR="${ASSETS_DIR:-assets}"
SQLITE_DOWNLOAD_PAGE="${SQLITE_DOWNLOAD_PAGE:-https://sqlite.org/download.html}"
SQLITE_BASE_URL="${SQLITE_BASE_URL:-https://sqlite.org}"
REQUESTED_VERSION="${SQLITE_VERSION:-}"
REQUESTED_URL="${SQLITE_URL:-}"
EXPECTED_SHA="${EXPECTED_SHA:-}"
TEMP_DIR=$(mktemp -d "${TMPDIR:-/tmp}/sqlite-wasm-download.XXXXXX")
trap 'rm -rf "$TEMP_DIR"' EXIT

if [ -n "$REQUESTED_URL" ]; then
    SQLITE_URL="$REQUESTED_URL"
    SQLITE_ZIP="${SQLITE_URL##*/}"
    SQLITE_VERSION="${REQUESTED_VERSION:-${SQLITE_ZIP#sqlite-wasm-}}"
    SQLITE_VERSION="${SQLITE_VERSION%.zip}"
    SQLITE_RELEASE="${SQLITE_RELEASE:-$SQLITE_VERSION}"
else
    echo "📦 Fetching SQLite WASM metadata..."
    DOWNLOAD_HTML=$(curl -fsSL "$SQLITE_DOWNLOAD_PAGE")
    WASM_RECORD=$(printf '%s\n' "$DOWNLOAD_HTML" | awk -F, -v requested="$REQUESTED_VERSION" '
        $1 == "PRODUCT" && $3 ~ /^20[0-9][0-9]\/sqlite-wasm-[0-9]+\.zip$/ {
            if (requested == "" || $3 ~ "/sqlite-wasm-" requested "\\.zip$") {
                print
                exit
            }
        }
    ')

    if [ -z "$WASM_RECORD" ]; then
        if [ -n "$REQUESTED_VERSION" ]; then
            echo "❌ Could not find sqlite-wasm-${REQUESTED_VERSION}.zip in $SQLITE_DOWNLOAD_PAGE"
        else
            echo "❌ Could not find a sqlite-wasm download in $SQLITE_DOWNLOAD_PAGE"
        fi
        exit 1
    fi

    IFS=, read -r _ SQLITE_RELEASE RELATIVE_URL _ EXPECTED_SHA <<EOF
$WASM_RECORD
EOF
    SQLITE_ZIP="${RELATIVE_URL##*/}"
    SQLITE_VERSION="${SQLITE_ZIP#sqlite-wasm-}"
    SQLITE_VERSION="${SQLITE_VERSION%.zip}"
    SQLITE_URL="${SQLITE_BASE_URL%/}/$RELATIVE_URL"
fi

echo "📦 Fetching SQLite WASM ${SQLITE_RELEASE} (${SQLITE_VERSION})..."

# Create assets directory if it doesn't exist
mkdir -p "$ASSETS_DIR"
rm -f "$ASSETS_DIR/sqlite3-worker1.js" "$ASSETS_DIR/sqlite3-worker1-promiser.js"

# Download the zip file
echo "⬇️  Downloading from $SQLITE_URL..."
curl -L -o "$TEMP_DIR/sqlite-wasm.zip" "$SQLITE_URL"

# Verify SHA3-256 checksum
echo "🔐 Verifying checksum..."
if command -v sha3sum >/dev/null 2>&1; then
    ACTUAL_SHA=$(sha3sum -a 256 "$TEMP_DIR/sqlite-wasm.zip" | cut -d' ' -f1)
elif command -v openssl >/dev/null 2>&1; then
    ACTUAL_SHA=$(openssl dgst -sha3-256 "$TEMP_DIR/sqlite-wasm.zip" | sed 's/^.*= //')
else
    echo "⚠️  Warning: Neither sha3sum nor openssl found. Skipping checksum verification."
    ACTUAL_SHA=""
fi

if [ -n "$ACTUAL_SHA" ] && [ -n "$EXPECTED_SHA" ]; then
    if [ "$ACTUAL_SHA" != "$EXPECTED_SHA" ]; then
        echo "❌ Checksum mismatch!"
        echo "   Expected: $EXPECTED_SHA"
        echo "   Got:      $ACTUAL_SHA"
        exit 1
    fi
    echo "✅ Checksum verified"
elif [ -z "$EXPECTED_SHA" ]; then
    echo "⚠️  Warning: EXPECTED_SHA not set. Skipping checksum verification."
fi

# Extract the zip file
echo "📂 Extracting files..."
unzip -q "$TEMP_DIR/sqlite-wasm.zip" -d "$TEMP_DIR"

# Copy only the required files
echo "📋 Copying required files to $ASSETS_DIR..."
cp "$TEMP_DIR/sqlite-wasm-${SQLITE_VERSION}/jswasm/sqlite3.js" "$ASSETS_DIR/"
cp "$TEMP_DIR/sqlite-wasm-${SQLITE_VERSION}/jswasm/sqlite3.wasm" "$ASSETS_DIR/"
cp "$TEMP_DIR/sqlite-wasm-${SQLITE_VERSION}/jswasm/sqlite3-opfs-async-proxy.js" "$ASSETS_DIR/"

# Clean up
echo "🧹 Cleaning up..."

echo "✨ SQLite WASM assets successfully fetched and installed!"
echo "   Files in $ASSETS_DIR:"
ls -la "$ASSETS_DIR"
