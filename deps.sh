#!/usr/bin/env sh

set -eu

# Get root of repository.
repo_root=$(git rev-parse --show-toplevel)

export GOBIN="${repo_root}/bin"

cd "${repo_root}/_tools"

# Temporarily disable "unbound variable" mode because no argument is a valid use case.
set +u
args="${*}"
set -u

if [ -z "${args}" ]; then
  echo "No tools specified, installing everything..."

  # Retrieve all tools inside "import" block in tools.go
  tools=$(go list -e -f "{{range .Imports}}{{.}} {{end}}" ./tools.go)

  # Ensure build dependencies are downloaded and ready for build phase.
  go mod tidy
else
  # Use tools passed as arguments.
  tools=${args}
fi

for tool in $tools; do
  printf "Installing \e[94m%s\e[0m...\n" "${tool}"
  go install "${tool}"
done
