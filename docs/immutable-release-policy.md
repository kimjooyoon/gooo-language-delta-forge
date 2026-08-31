# Immutable release policy

The repository-level policy is enabled through the GitHub REST API:

- `PUT https://api.github.com/repos/kimjooyoon/gooo-language-delta-forge/immutable-releases`
- `GET https://api.github.com/repos/kimjooyoon/gooo-language-delta-forge/immutable-releases`
- API version: `2026-03-10`
- observed setting: `enabled=true`, `enforced_by_owner=false`

The existing `v0.1.0` release predates this setting and remains the historical
`NON_DURABLE_RELEASE`. The release workflow creates a draft, uploads all
assets, publishes it, and then fails closed unless the GitHub REST release
record reports `immutable=true`, the tag is annotated and points to the exact
expected commit, and every expected asset name, size, and digest matches.

`v0.1.1` is the first release eligible for the durable-release designation.
