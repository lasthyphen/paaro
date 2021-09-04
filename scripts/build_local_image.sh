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

# WARNING: this will use the most recent commit even if there are un-committed changes present
full_commit_hash="$(git --git-dir="$DIJETS_PATH/.git" rev-parse HEAD)"
commit_hash="${full_commit_hash::8}"

echo "Building Docker Image with tags: $paaro_dockerhub_repo:$commit_hash , $paaro_dockerhub_repo:$current_branch"
docker build -t "$paaro_dockerhub_repo:$commit_hash" \
        -t "$paaro_dockerhub_repo:$current_branch" "$DIJETS_PATH" -f "$DIJETS_PATH/Dockerfile"
