#!/bin/bash
# -----------------------------------------------------------------------------
# GravityAnchor macOS DMG Packaging Script
# 
# This script bundles the compiled GravityAnchor.app into a standard,
# high-quality, drag-and-drop DMG installer for macOS distribution.
# Zero external dependencies required.
# -----------------------------------------------------------------------------

set -e

APP_NAME="GravityAnchor"
APP_PATH="build/bin/${APP_NAME}.app"
DMG_NAME="GravityAnchor_macOS_universal.dmg"
DMG_PATH="build/bin/${DMG_NAME}"
STAGE_DIR="build/bin/dmg_stage"

echo "===================================================="
echo "⚓ Starting GravityAnchor macOS DMG Packaging..."
echo "===================================================="

# Check if application bundle exists
if [ ! -d "$APP_PATH" ]; then
    echo "❌ Error: $APP_PATH does not exist."
    echo "Please build the application first by running:"
    echo "  wails build -platform darwin/universal"
    exit 1
fi

# Clean up previous build remnants
echo "🧹 Cleaning up old staging files..."
rm -f "$DMG_PATH"
rm -rf "$STAGE_DIR"
mkdir -p "$STAGE_DIR"

# Copy the app bundle
echo "📦 Copying $APP_NAME.app to staging directory..."
cp -R "$APP_PATH" "$STAGE_DIR/"

# Create standard symlink to /Applications inside the DMG
echo "🔗 Creating symlink to /Applications..."
ln -s /Applications "$STAGE_DIR/Applications"

# Compile staging folder to highly-compressed UDZO DMG
echo "💿 Generating DMG file using hdiutil..."
hdiutil create \
    -volname "${APP_NAME}" \
    -srcfolder "$STAGE_DIR" \
    -ov \
    -format UDZO \
    "$DMG_PATH"

# Clean up staging directories
echo "🧹 Post-build clean up..."
rm -rf "$STAGE_DIR"

echo "===================================================="
echo "🎉 SUCCESS: macOS DMG installer compiled!"
echo "📍 Location: $DMG_PATH"
echo "===================================================="
