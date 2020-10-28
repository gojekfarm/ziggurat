#!/usr/bin/env bash

VERSION="$1"

git push origin master
git tag "$VERSION"
git push origin master --tags
