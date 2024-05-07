#! /usr/bin/bash

set -e

VERSION="$1"
if [[ -z "${VERSION}" ]] ; then
    echo "existing versions:"
    git tag
    exit 0
fi

go mod tidy

if [[ ! -z "$(git status --porcelain=2)" ]] ; then
    echo "must have a clean repo"
    exit 1
fi

echo "testing..."
go test ./...

git tag "${VERSION}"

git push origin "${VERSION}"

GOPROXY=proxy.golang.org go list -m "github.com/krelinga/mkv-util-server@${VERSION}"

echo "Success!"
