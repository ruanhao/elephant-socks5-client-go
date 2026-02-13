#!/usr/bin/env bash
# -*- coding: utf-8 -*-
#
# Description:


set -e
GIT_COMMIT=$(git rev-parse --short HEAD)
PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64" "windows/arm64")
OUTPUT_NAME="elephant"

for PLATFORM in "${PLATFORMS[@]}"; do
    IFS="/" read -r GOOS GOARCH <<< "$PLATFORM"
    OUTPUT_FILE="${OUTPUT_NAME}_${GOOS}_${GOARCH}"
    if [ "$GOOS" == "windows" ]; then
        OUTPUT_FILE+=".exe"
    fi
    echo "Building for $GOOS/$GOARCH..."
    env GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="-s -w -X main.gitCommit=${GIT_COMMIT}" -o "$OUTPUT_FILE" .
    echo "Created artifact: $OUTPUT_FILE"
done

