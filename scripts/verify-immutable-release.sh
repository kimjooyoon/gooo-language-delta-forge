#!/usr/bin/env bash
set -euo pipefail

repo=${1:?repository is required}
tag=${2:?tag is required}
expected_commit=${3:?expected commit is required}
manifest_path=${4:?release manifest is required}
assets_dir=${5:?assets directory is required}

release=$(gh api -H 'Accept: application/vnd.github+json' -H 'X-GitHub-Api-Version: 2026-03-10' "repos/$repo/releases/tags/$tag")
test "$(jq -r '.immutable // false' <<<"$release")" = "true"
release_id=$(jq -r '.id' <<<"$release")

ref=$(gh api -H 'Accept: application/vnd.github+json' -H 'X-GitHub-Api-Version: 2026-03-10' "repos/$repo/git/ref/tags/$tag")
test "$(jq -r '.object.type' <<<"$ref")" = "tag"
tag_object_sha=$(jq -r '.object.sha' <<<"$ref")
tag_object=$(gh api -H 'Accept: application/vnd.github+json' -H 'X-GitHub-Api-Version: 2026-03-10' "repos/$repo/git/tags/$tag_object_sha")
test "$(jq -r '.object.type' <<<"$tag_object")" = "commit"
test "$(jq -r '.object.sha' <<<"$tag_object")" = "$expected_commit"

expected_count=$(jq -r '.expected_asset_count' "$manifest_path")
mapfile -t expected_names < <(jq -r '.expected_asset_names[]' "$manifest_path")
test "$expected_count" = "${#expected_names[@]}"
actual_count=$(jq '.assets | length' <<<"$release")
test "$actual_count" = "$expected_count"

expected_names_file=$(mktemp)
actual_names_file=$(mktemp)
for name in "${expected_names[@]}"; do echo "$name"; done | sort > "$expected_names_file"
jq -r '.assets[].name' <<<"$release" | sort > "$actual_names_file"
diff -u "$expected_names_file" "$actual_names_file"

verify_dir=$(mktemp -d)
verified_assets='[]'
for name in "${expected_names[@]}"; do
  local_path="$assets_dir/$name"
  test -s "$local_path"
  expected_size=$(stat -c '%s' "$local_path")
  expected_digest="sha256:$(sha256sum "$local_path" | awk '{print $1}')"
  asset=$(jq -c --arg name "$name" '.assets[] | select(.name == $name)' <<<"$release")
  test -n "$asset" && test "$asset" != "null"
  asset_id=$(jq -r '.id' <<<"$asset")
  test "$(jq -r '.size' <<<"$asset")" = "$expected_size"
  test "$(jq -r '.digest' <<<"$asset")" = "$expected_digest"
  gh api -H 'Accept: application/octet-stream' "repos/$repo/releases/assets/$asset_id" > "$verify_dir/$name"
  downloaded_size=$(stat -c '%s' "$verify_dir/$name")
  downloaded_digest="sha256:$(sha256sum "$verify_dir/$name" | awk '{print $1}')"
  test "$downloaded_size" = "$expected_size"
  test "$downloaded_digest" = "$expected_digest"
  verified_assets=$(jq --argjson asset "$asset" '. + [{id:$asset.id,name:$asset.name,size:($asset.size|tonumber),digest:$asset.digest}]' <<<"$verified_assets")
done

jq -n --arg repo "$repo" --arg tag "$tag" --argjson release_id "$release_id" \
  --arg tag_object_sha "$tag_object_sha" --arg target_sha "$expected_commit" \
  --argjson assets "$verified_assets" \
  '{schema:"gooo/language-delta-forge/immutable-release-verification/v1",repository:$repo,tag:$tag,release_id:$release_id,immutable:true,tag_object_sha:$tag_object_sha,target_sha:$target_sha,assets:$assets}'
