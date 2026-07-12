#!/usr/bin/env bash
set -euo pipefail

# User journey: a maintainer can produce every documented Release download
# locally, and each archive is complete and checksum-verifiable.
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work_dir="$(mktemp -d)"
trap 'rm -rf "$work_dir"' EXIT

version="v9.9.9-test"
dist_dir="$work_dir/dist"

"$repo_root/scripts/build-release.sh" "$version" "$dist_dir" "$repo_root"
"$repo_root/scripts/verify-release-assets.sh" "$version" "$dist_dir"
