#!/bin/sh
set -e

echo "Installing dependencies..."
yarn install --immutable

echo "Building frontend assets..."
yarn run build

echo "Build complete."
