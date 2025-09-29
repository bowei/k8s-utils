#!/bin/bash

# Build script for generating CSS themes from LESS templates
# Usage: ./build-themes.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Building themes from LESS templates..."

# Check if lessc is available
if ! command -v lessc &> /dev/null; then
    echo "Error: lessc (LESS compiler) not found. Install with: npm install -g less"
    exit 1
fi

# Build light theme
echo "Building light theme..."
cp variables-light.less variables.less
lessc app.less app-light.css
echo "✓ Light theme generated: app-light.css"

# Build dark theme
echo "Building dark theme..."
cp variables-dark.less variables.less
lessc app.less app-dark.css
echo "✓ Dark theme generated: app-dark.css"

# Build blue theme (example)
echo "Building blue theme..."
cp variables-blue.less variables.less
lessc app.less app-blue.css
echo "✓ Blue theme generated: app-blue.css"

# Build green theme
echo "Building green theme..."
cp variables-green.less variables.less
lessc app.less app-green.css
echo "✓ Green theme generated: app-green.css"

# Build brown theme
echo "Building brown theme..."
cp variables-brown.less variables.less
lessc app.less app-brown.css
echo "✓ Brown theme generated: app-brown.css"

# Clean up temporary variables file
rm variables.less

echo "All themes built successfully!"
echo ""
echo "To add a new theme:"
echo "1. Create a new variables-THEMENAME.less file"
echo "2. Update this script to build it"
echo "3. Run ./build-themes.sh"