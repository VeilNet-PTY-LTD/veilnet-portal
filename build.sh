#!/bin/bash

# VeilNet Portal Build Script
# Builds for all major platforms and architectures

set -e

BUILD_DIR="build"

# Clean build directory
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

echo "Building VeilNet Portal..."

# Build for all platforms
echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -o $BUILD_DIR/veilnet-portal-linux-amd64 .

echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -o $BUILD_DIR/veilnet-portal-linux-arm64 .

echo "Build completed! Files in $BUILD_DIR/"
ls -lh $BUILD_DIR/