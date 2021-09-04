#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Dijetsgo root folder
DIJETS_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )"; cd .. && pwd )
# Load the versions
source "$DIJETS_PATH"/scripts/versions.sh
# Load the constants
source "$DIJETS_PATH"/scripts/constants.sh

# Download dependencies
echo "Downloading dependencies..."
go mod download

# Build paaro
"$DIJETS_PATH"/scripts/build_dijets.sh

# Build coreth
"$DIJETS_PATH"/scripts/build_coreth.sh

# Exit build successfully if the binaries are created
if [[ -f "$paaro_path" && -f "$evm_path" ]]; then
        echo "Build Successful"
        exit 0
else
        echo "Build failure" >&2
        exit 1
fi
