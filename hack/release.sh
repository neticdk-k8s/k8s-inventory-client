#!/usr/bin/env bash

set -eo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

next_version=$("${SCRIPT_DIR}"/next_version.sh "$1")

echo "Next version is: ${next_version}"

if [ -z "${FORCE}" ]; then
	(
		cd "${SCRIPT_DIR}/.."
		git update-index -q --ignore-submodules --refresh
		err=0

		echo "Checking for unstaged changes..."
		if ! git diff-files --quiet --ignore-submodules --; then
			echo >&2 "Cannot tag release: you have unstaged changes."
			git diff-files --name-status -r --ignore-submodules -- >&2
			err=1
		fi

		echo "Checking for uncommitted changes..."
		if ! git diff-index --cached --quiet HEAD --ignore-submodules --; then
			echo >&2 "Cannot tag release: your index contains uncommitted changes."
			git diff-index --cached --name-status -r --ignore-submodules HEAD -- >&2
			err=1
		fi

		if [ $err = 1 ]; then
			echo >&2 "Please commit or stash them."
			exit 1
		fi
	)
fi

echo "Updating helm chart appVersion..."
(
	cd "${SCRIPT_DIR}/.."
	gsed -i -e "s/appVersion:.*/appVersion: \"${next_version}\"/" chart/Chart.yaml
	git commit -m "chore: upgrade chart appVersion to ${next_version}" chart/Chart.yaml
)

echo "Pushing release..."
(
	cd "${SCRIPT_DIR}/.."
	git push
)

echo "Tagging release..."
(
	cd "${SCRIPT_DIR}/.."
	git tag "v${next_version}"
	git push --tags
)

echo "Done"
