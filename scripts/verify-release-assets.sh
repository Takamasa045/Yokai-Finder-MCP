#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 <version-tag> <release-directory>" >&2
  exit 2
}

version="${1:-}"
release_dir="${2:-}"

[[ "$version" =~ ^v[0-9][0-9A-Za-z._-]*$ && -d "$release_dir" ]] || usage

targets=(
  "darwin amd64 tar.gz"
  "darwin arm64 tar.gz"
  "linux amd64 tar.gz"
  "linux arm64 tar.gz"
  "windows amd64 zip"
  "windows arm64 zip"
)
expected_files=(checksums.txt)
expected_archives=()

for target in "${targets[@]}"; do
  read -r goos goarch archive_type <<<"$target"
  package_name="yokai-finder-mcp_${version}_${goos}_${goarch}"
  archive="$package_name.$archive_type"
  expected_files+=("$archive")
  expected_archives+=("$archive")

  [[ -f "$release_dir/$archive" ]] || {
    echo "Missing release asset: $archive" >&2
    exit 1
  }

  if [[ "$archive_type" == "zip" ]]; then
    entries="$(unzip -Z1 "$release_dir/$archive")"
    binary_name="yokai-finder-mcp.exe"
  else
    entries="$(tar -tzf "$release_dir/$archive")"
    binary_name="yokai-finder-mcp"
  fi

  grep -Fxq "$package_name/$binary_name" <<<"$entries" || {
    echo "Missing executable in $archive" >&2
    exit 1
  }
  grep -Fxq "$package_name/README.md" <<<"$entries" || {
    echo "Missing README in $archive" >&2
    exit 1
  }
  grep -Fxq "$package_name/LICENSE" <<<"$entries" || {
    echo "Missing LICENSE in $archive" >&2
    exit 1
  }
done

actual_files=()
while IFS= read -r file; do
  actual_files+=("$(basename "$file")")
done < <(find "$release_dir" -maxdepth 1 -type f -print | sort)

expected_list="$(printf '%s\n' "${expected_files[@]}" | sort)"
actual_list="$(printf '%s\n' "${actual_files[@]}" | sort)"
[[ "$actual_list" == "$expected_list" ]] || {
  echo "Release assets do not match the expected manifest." >&2
  diff -u <(printf '%s\n' "$expected_list") <(printf '%s\n' "$actual_list") || true
  exit 1
}

expected_archive_list="$(printf '%s\n' "${expected_archives[@]}" | sort)"
checksum_archive_list="$(awk '{print $2}' "$release_dir/checksums.txt" | sed 's/^\*//' | sort)"
[[ "$checksum_archive_list" == "$expected_archive_list" ]] || {
  echo "Checksum manifest does not cover exactly the release archives." >&2
  diff -u <(printf '%s\n' "$expected_archive_list") <(printf '%s\n' "$checksum_archive_list") || true
  exit 1
}

(
  cd "$release_dir"
  shasum -a 256 -c checksums.txt
)
