#!/bin/bash

cd $(dirname $0)

set -eu

readonly VERSION=$1; shift
readonly TAG=v${VERSION}

if (git tag | grep --quiet --line-regexp "${TAG}"); then
	echo "Version ${VERSION} already exists"
	exit 1
fi

# Ensure all tests pass
make build test

# Update all vendor packages
govendor fetch +vendor

# Validate that tests still pass
make build test

# Write new version
echo ${VERSION} > VERSION

# Commit everything
git add .
git commit -m "Bump to version ${VERSION}"
git push
# Define tag for the release
git tag -a "${TAG}" -m "Tag version ${VERSION}"
git push --tags
