#!/bin/sh
set -e

echo "Installing dependencies with yarn..."
yarn install --immutable

echo "Starting development server..."
exec yarn dev
