#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Testing specific variables
dijets_testing_repo="djtplatform/dijets-testing"
paaro_byzantine_repo="djtplatform/dijets-byzantine"

# Define default versions to use
dijets_testing_image="djtplatform/dijets-testing:master"
paaro_byzantine_image="djtplatform/dijets-byzantine:master"

# Fetch the images
# If Docker Credentials are not available fail
if [[ -z ${DOCKER_USERNAME} ]]; then
    echo "Skipping Tests because Docker Credentials were not present."
    exit 1
fi

# Dijets root directory
DIJETS_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )"; cd ../.. && pwd )

# Load the versions
source "$DIJETS_PATH"/scripts/versions.sh

# Load the constants
source "$DIJETS_PATH"/scripts/constants.sh

# Login to docker
echo "$DOCKER_PASS" | docker login --username "$DOCKER_USERNAME" --password-stdin

# Checks available docker tags exist
function docker_tag_exists() {
    TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "'${DOCKER_USERNAME}'", "password": "'${DOCKER_PASS}'"}' https://hub.docker.com/v2/users/login/ | jq -r .token)
    curl --silent -H "Authorization: JWT ${TOKEN}" -f --head -lL https://hub.docker.com/v2/repositories/$1/tags/$2/ > /dev/null
}

testBatch="${1:-}"
shift 1

echo "Running Test Batch: ${testBatch}"

# Defines the dijets-testing tag to use
# Either uses the same tag as the current branch or uses the default
if docker_tag_exists $dijets_testing_repo $current_branch; then
    echo "$dijets_testing_repo:$current_branch exists; using this image to run e2e tests"
    dijets_testing_image="$dijets_testing_repo:$current_branch"
else
    echo "$dijets_testing_repo $current_branch does NOT exist; using the default image to run e2e tests"
fi

# Defines the paaro-byzantine tag to use
# Either uses the same tag as the current branch or uses the default
if docker_tag_exists $paaro_byzantine_repo $current_branch; then
    echo "$paaro_byzantine_repo:$current_branch exists; using this image to run e2e tests"
    paaro_byzantine_image="$paaro_byzantine_repo:$current_branch"
else
    echo "$paaro_byzantine_repo $current_branch does NOT exist; using the default image to run e2e tests"
fi

echo "Using $dijets_testing_image for e2e tests"
echo "Using $paaro_byzantine_image for e2e tests"

# pulling the dijets-testing image
docker pull $dijets_testing_image
docker pull $paaro_byzantine_image

# Setting the build ID
git_commit_id=$( git rev-list -1 HEAD )

# Build current paaro
source "$DIJETS_PATH"/scripts/build_image.sh

# Target built version to use in dijets-testing
dijets_image="$paaro_dockerhub_repo:$current_branch"

echo "Execution Summary:"
echo ""
echo "Running Dijets Image: ${dijets_image}"
echo "Running Dijets Image Tag: $current_branch"
echo "Running Dijets Testing Image: ${dijets_testing_image}"
echo "Running Dijets Byzantine Image: ${paaro_byzantine_image}"
echo "Git Commit ID : ${git_commit_id}"
echo ""


# >>>>>>>> dijets-testing custom parameters <<<<<<<<<<<<<
custom_params_json="{
    \"isKurtosisCoreDevMode\": false,
    \"paaroImage\":\"${dijets_image}\",
    \"paaroByzantineImage\":\"${paaro_byzantine_image}\",
    \"testBatch\":\"${testBatch}\"
}"
# >>>>>>>> dijets-testing custom parameters <<<<<<<<<<<<<

bash "$DIJETS_PATH/.kurtosis/kurtosis.sh" \
    --custom-params "${custom_params_json}" \
    ${1+"${@}"} \
    "${dijets_testing_image}" 
