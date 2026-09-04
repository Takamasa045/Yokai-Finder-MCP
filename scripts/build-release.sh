#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 <version-tag> [output-dir] [source-directory]" >&2
  exit 2
}

version="${1:-}"
output_dir="${2:-dist}"
source_dir="${3:-}"

[[ "$version" =~ ^v[0-9][0-9A-Za-z._-]*$ ]] || usage

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source_dir="${source_dir:-$repo_root}"
[[ -f "$source_dir/go.mod" && -f "$source_dir/README.md" ]] || usage
source_dir="$(cd "$source_dir" && pwd)"
cd "$source_dir"

rm -rf "$output_dir"
mkdir -p "$output_dir"
output_dir="$(cd "$output_dir" && pwd)"

work_dir="$(mktemp -d)"
trap 'rm -rf "$work_dir"' EXIT

targets=(
  "darwin amd64 tar.gz"
  "darwin arm64 tar.gz"
  "linux amd64 tar.gz"
  "linux arm64 tar.gz"
  "windows amd64 zip"
  "windows arm64 zip"
)

for target in "${targets[@]}"; do
  read -r goos goarch archive_type <<<"$target"
  package_name="yokai-finder-mcp_${version}_${goos}_${goarch}"
  package_dir="$work_dir/$package_name"
  binary_name="yokai-finder-mcp"
  [[ "$goos" == "windows" ]] && binary_name+=".exe"

  mkdir -p "$package_dir"
  version_plain="${version#v}"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath \
      -ldflags="-s -w -X github.com/Takamasa045/Yokai-Finder-MCP/internal/version.Version=${version_plain}" \
      -o "$package_dir/$binary_name" ./cmd/server
  cp README.md "$package_dir/README.md"

  if [[ "$archive_type" == "zip" ]]; then
    (
      cd "$work_dir"
      zip -q -r "$output_dir/$package_name.zip" "$package_name"
    )
  else
    tar -C "$work_dir" -czf "$output_dir/$package_name.tar.gz" "$package_name"
  fi
done

(
  cd "$output_dir"
  shasum -a 256 yokai-finder-mcp_*.tar.gz yokai-finder-mcp_*.zip > checksums.txt
)
