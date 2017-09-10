#!/bin/bash

cd $(dirname $0)

set -eu

readonly VERSION=$1; shift

# Ensure all tests pass
make build test

# Update all vendor packages
govendor fetch +vendor

# Validate that tests still pass
make build test

# Write new version
cat ${VERSION} > VERSION

# Commit everything
git add .
git commit -m "Bump to version ${VERSION}"
git tag -a ${VERSION} -m "Tag version ${VERSION}"
