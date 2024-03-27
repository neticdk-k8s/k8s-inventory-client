#!/usr/bin/env bash

set -eo pipefail

git fetch --all --tags

latest_tag=$(git tag --sort='v:refname' | tail -1)
latest_version=$(echo "${latest_tag}" | tr -d 'v')

if [ -n "${DEBUG}" ]; then
	printf "%-14s : %10s\n" "Latest tag" "${latest_tag}"
	printf "%-14s : %10s\n" "Latest version" "${latest_version}"
fi

IFS="." read -r -a semver <<<"${latest_version}"

# Assume patch version bump if no argument given
if [ -z "$1" ] || [ "$1" = "patch" ]; then
	((semver[2]++))
elif [ "$1" = "minor" ]; then
	((semver[1]++))
	semver[2]=0
fi

next_version="${semver[0]}.${semver[1]}.${semver[2]}"

if [ -n "${DEBUG}" ]; then
	printf "%-14s : %10s\n" "Next version" "${next_version}"
fi

if [[ ! ${next_version} =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	echo >&2 "${next_version} is not valid"
	exit 1
fi

echo "${next_version}"
