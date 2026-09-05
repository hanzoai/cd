#!/usr/bin/env bash

out="| Hanzo CD version | Kubernetes versions |\n"
out+="|-----------------|---------------------|\n"

current_version=$(git rev-parse --abbrev-ref HEAD | sed 's/release-//')
major_version_num=$(echo "$current_version" | sed -E 's/\.[0-9]+//')
minor_version_num=$(echo "$current_version" | sed -E 's/[0-9]+\.//')

for _ in {1..3}; do
  version="${major_version_num}.${minor_version_num}"
  git checkout "release-$version" > /dev/null || exit 1

  line=$(yq '.jobs["test-e2e"].strategy.matrix |
    # k3s-version was an array prior to 2.12. This checks for the old format first and then falls back to the new format.
    (.["k3s-version"] // (.k3s | map(.version))) |
    .[]' .github/workflows/ci-build.yaml | \
    jq --arg version "$version" --raw-input --slurp --raw-output \
    'split("\n")[:-1] | map(sub("\\.[0-9]+$"; "")) | join(", ") | "| \($version) | \(.) |"')
  out+="$line\n"


  # If we're at minor version 0, there's no further version back in this series. Instead, move to the latest version in
  # the previous major release series.
  if [ "$minor_version_num" -eq 0 ]; then
    major_version_num=$((major_version_num - 1))
    # Get the latest minor version in the previous series.
    minor_version_num=$(git tag -l "v$major_version_num.*" | sort -V | tail -n 1 | sed -E 's/\.[0-9]+$//' | sed -E 's/^v[0-9]+\.//')
  else
    minor_version_num=$((minor_version_num - 1))
  fi
done

git checkout "release-$current_version"

echo -en "$out" > docs/operator-manual/tested-kubernetes-versions.md
