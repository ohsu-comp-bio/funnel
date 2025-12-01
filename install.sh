#!/bin/bash

echo "Installing Funnel..."

# Define the base repository URL
REPO="ohsu-comp-bio/funnel"

# Function to get the latest release URL if no version is provided
get_latest_release_url() {
    echo "https://api.github.com/repos/$REPO/releases/latest"
}

# Function to get the release URL for a specific tag
get_tag_release_url() {
    echo "https://api.github.com/repos/$REPO/releases/tags/$1"
}

# Parse version tag argument
VERSION_TAG=$1

# Determine the release URL based on whether a version tag was provided
RELEASE_URL=""
if [ -z "$VERSION_TAG" ]; then
    echo "No version specified. Fetching the latest release..."
    RELEASE_URL=$(get_latest_release_url)
    VERSION_TAG=$(curl -s $RELEASE_URL | grep '"tag_name":' | cut -d '"' -f 4)
else
    echo "Fetching release for version $VERSION_TAG..."
    RELEASE_URL=$(get_tag_release_url $VERSION_TAG)
fi

# Determine OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" == "x86_64" ]; then
  ARCH="amd64"
elif [[ "$ARCH" == "aarch64" || "$ARCH" == "arm64" ]]; then
  ARCH="arm64"
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

# Define the tar file based on OS and Architecture
TAR_FILE="funnel-${OS}-${ARCH}"
CHECKSUM_CANDIDATES=("funnel-${VERSION_TAG}-checksums.txt" "funnel_${VERSION_TAG}_checksums.txt")
CHECKSUM_FILE=""

# Fetch the release assets URLs
ASSETS=$(curl -s $RELEASE_URL | grep "browser_download_url" | cut -d '"' -f 4)

# Download the tar.gz file and checksums.txt for the detected OS and Arch
echo "Downloading Funnel $VERSION_TAG for $OS $ARCH..."
for asset in $ASSETS; do
    asset_name=$(basename "$asset")
    
    # Binary (tar.gz)
    if [[ "$asset" == *"${TAR_FILE}"* && "$asset" == *.tar.gz ]]; then
        TAR_URL="$asset"
        TAR_NAME="$asset_name"
        echo " ➜ $TAR_NAME"
        curl -L --progress-bar -o "$TAR_NAME" "$TAR_URL"

    # Checksums
    elif [[ "$asset_name" == *checksums* ]]; then
        CHECKSUM_FILE="$asset_name"
        echo " ➜ $CHECKSUM_FILE"
        curl -L --progress-bar -o "$CHECKSUM_FILE" "$asset"
    fi
done

# Verify checksum
echo "Verifying checksum..."
CHECKSUM_EXPECTED=$(grep $TAR_NAME $CHECKSUM_FILE | awk '{print $1}')
CHECKSUM_ACTUAL=$(shasum -a 256 $TAR_NAME | awk '{print $1}')

if [ "$CHECKSUM_EXPECTED" != "$CHECKSUM_ACTUAL" ]; then
    echo "Checksum verification failed for $TAR_NAME. Exiting..."
    exit 1
fi

# Extract and install the package
echo "Extracting the package..."
tar -xzf $TAR_NAME

# Parse installation destination
DEST=$2

# Determine where to install the Funnel binary
if [ -z "$DEST" ]; then
    DEST=$HOME/.local/bin
fi
echo "Installing Funnel to $DEST..."
mkdir -p $DEST
mv funnel $DEST

# Clean up
rm $TAR_NAME $CHECKSUM_FILE

echo "Installation successful: $DEST/funnel"; echo
$DEST/funnel version
echo; echo "Run 'funnel --help' for more info"
