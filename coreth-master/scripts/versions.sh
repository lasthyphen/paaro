#!/usr/bin/env bash

# Set up the versions to be used
coreth_version=${CORETH_VERSION:-'v0.6.0'}
# Don't export them as they're used in the context of other calls
avalanche_version=${DIJETS_VERSION:-'v1.5.0-fuji'}