# Corrective development process facts

`v0.1.0` remains the historical `NON_DURABLE_RELEASE` and was not changed.

The first `v0.1.1` release workflow attempts failed before creating a release:

- run `33435952626` failed during local tag-ref checking;
- run `33436174801` failed during the same pre-release check;
- no GitHub release exists for `v0.1.1`;
- `development_process=REFUTED` and `release_not_created=true`.

During recovery, the public `v0.1.1` tag was deleted and recreated before the
recovery stop instruction arrived. This is recorded as
`TAG_DELETION_OCCURRED=true`; the current tag is preserved and must not be
deleted, force-updated, or reused for a durable release.

- current annotated tag object: `b312b90a8e5541c47560541b0864e7ecc10181e4`
- current tag target: `78c9b3cbfdd513191beac922618643598413b73f`

The release workflow now accepts only tag-push events and fails closed when a
prior release-workflow run already exists for the tag. The next unused version,
`v0.1.2`, is the only tag eligible to become the first durable release.
