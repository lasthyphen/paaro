#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

echo "Building docker image based off of most recent local commits of paaro and coreth"

DIJETS_REMOTE="git@github.com:lasthyphen/paaro.git"
CORETH_REMOTE="git@github.com:lasthyphen/coreth.git"
DOCKERHUB_REPO="djtplatform/paaro"

DOCKER="${DOCKER:-docker}"
SCRIPT_DIRPATH=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
ROOT_DIRPATH="$(dirname "${SCRIPT_DIRPATH}")"

LASTHYPHEN_RELATIVE_PATH="src/github.com/lasthyphen"
EXISTING_GOPATH="$GOPATH"

export GOPATH="$SCRIPT_DIRPATH/.build_image_gopath"
WORKPREFIX="$GOPATH/src/github.com/lasthyphen"

# Clone the remotes and checkout the desired branch/commits
DIJETS_CLONE="$WORKPREFIX/paaro"
CORETH_CLONE="$WORKPREFIX/coreth"

# Replace the WORKPREFIX directory
rm -rf "$WORKPREFIX"
mkdir -p "$WORKPREFIX"


DIJETS_COMMIT_HASH="$(git -C "$EXISTING_GOPATH/$LASTHYPHEN_RELATIVE_PATH/paaro" rev-parse --short HEAD)"
CORETH_COMMIT_HASH="$(git -C "$EXISTING_GOPATH/$LASTHYPHEN_RELATIVE_PATH/coreth" rev-parse --short HEAD)"

git config --global credential.helper cache

git clone "$DIJETS_REMOTE" "$DIJETS_CLONE"
git -C "$DIJETS_CLONE" checkout "$DIJETS_COMMIT_HASH"

git clone "$CORETH_REMOTE" "$CORETH_CLONE"
git -C "$CORETH_CLONE" checkout "$CORETH_COMMIT_HASH"

CONCATENATED_HASHES="$DIJETS_COMMIT_HASH-$CORETH_COMMIT_HASH"

"$DOCKER" build -t "$DOCKERHUB_REPO:$CONCATENATED_HASHES" "$WORKPREFIX" -f "$SCRIPT_DIRPATH/local.Dockerfile"
