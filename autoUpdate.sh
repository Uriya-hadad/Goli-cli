#!/usr/bin/env bash

# create lock
LOCKFILE="/tmp/autoUpdate.lock"

# Check if the lock file exists
if [ -e "$LOCKFILE" ]; then
    echo "Script is already running."
    exit 1
fi

# Create the lock file
echo "creating lock"
touch "$LOCKFILE"

# Ensure lock file is removed when script exits
trap 'echo "removing lock";rm -f "$LOCKFILE"; exit' EXIT

# Configuration
BASE_URL="https://common.repositories.cloud.sap/artifactory/portal/go/plugins/goli/"
USERNAME="i564168"
PASSWORD="****************"
VERSION_FILE="version.txt"
LATEST_VERSION_FILE="latest_version.txt"
OUTPUT_FILE="Goli-latest-version.zip"
CLI_DIR="$1"


# Function to get data from Artifactory
getDataFromArtifactory() {
    local fileName="$1"
    curl  --silent -u "$USERNAME:$PASSWORD" -s "${BASE_URL}${fileName}"
}

# Read the current version
if [ ! -f "$VERSION_FILE" ]; then
    echo "version.txt not found, defaulting to 0.0.0."
    currentVersion="0.0.0"
else
    currentVersion=$(cat "$VERSION_FILE" | tr -d '\n')
fi

# Get the latest version from Artifactory
version=$(getDataFromArtifactory "$LATEST_VERSION_FILE" | tr -d '\n')

# Compare versions
if [ "$version" == "$currentVersion" ]; then
    echo "Already up to date."
    exit 0
fi

# Determine OS and Architecture
OS=$(uname -s)
ARCH=$(uname -m)
if [ "$OS" == "Darwin" ]; then
    if [ "$ARCH" == "arm64" ]; then
        fileToDownload="Goli-${version}-macOS-arm64.zip"
    else
        echo "Unsupported architecture for macOS."
        exit 1
    fi
elif [ "$OS" == "Linux" ]; then
    if [ "$ARCH" == "x86_64" ]; then
        fileToDownload="Goli-${version}-linux-amd64.zip"
    elif [ "$ARCH" == "arm64" ]; then
        fileToDownload="Goli-${version}-linux-arm64.zip"
    else
        echo "Unsupported architecture for Linux."
        exit 1
    fi
elif [[ "$OS" == "MINGW"* || "$OS" == "CYGWIN"* || "$OS" == "MSYS"* ]]; then
    fileToDownload="Goli-${version}-windows-amd64.zip"
else
    echo "Unsupported operating system."
    exit 1
fi

# Download the latest version
echo "Downloading $fileToDownload..."
curl --silent -u "$USERNAME:$PASSWORD" -o "$OUTPUT_FILE" "${BASE_URL}${fileToDownload}"

if [ $? -eq 0 ]; then
    echo "Downloaded and saved as $OUTPUT_FILE"
else
    echo "Failed to download the file."
    exit 1
fi

unzip -o "$OUTPUT_FILE" -d "./"
cp -f -r goliCli/* .
rm -rf goliCli
rm "$OUTPUT_FILE"
