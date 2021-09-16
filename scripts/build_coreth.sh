#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Directory above this script
DIJETS_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )"; cd .. && pwd )

# Load the versions
source "$DIJETS_PATH"/scripts/versions.sh

# Load the constants
source "$DIJETS_PATH"/scripts/constants.sh

# check if there's args defining different coreth source and build paths


# Build Coreth
echo "Building Coreth @ ${coreth_version} ..."
cd "$DIJETS_PATH/coreth"
go build -ldflags "-X github.com/lasthyphen/coreth/plugin/evm.Version=$coreth_version $static_ld_flags" -o "$evm_path" "plugin/"*.go
cd "$DIJETS_PATH"

# Building coreth + using go get can mess with the go.mod file.
go mod tidy
