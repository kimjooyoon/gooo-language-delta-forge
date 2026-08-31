# Gooo Language Delta Forge

`gooo-language-delta-forge` is an independent, read-only consumer of
immutable Gooo releases and observation receipts. Given one observed failure,
it emits a machine-checkable candidate bundle for a human-authorized language
change:

- exact concept, predicate, and field target;
- causal frontier;
- before/after semantic graph delta and rollback delta;
- expected positive and negative conformance cases;
- exact added, retired, and split cell counts; and
- source, graph, delta, test-manifest, and independent-consumer receipts.

The forge never edits a Gooo core checkout, opens or merges a change, or
adopts a candidate. Protected-core adoption is a separate authority step.
Candidate meaning must be present in the structured failure evidence. The Go
consumer validates and binds that evidence; it does not invent a target,
predicate, field, frontier, or graph mutation. Missing direct cause lowers the
result to a typed `UNKNOWN`. A related `REFUTED` receipt takes precedence.

## Quick start

The repository targets Go 1.27.0. The local development policy intentionally
does not run tests; GitHub Actions is the verification authority.

```text
go run ./cmd/gooo-language-delta-forge generate \
  --program examples/language-delta-forge-v1/main.gooo \
  --input fixtures/input-normal.json \
  --output /tmp/gooo-language-delta-candidate
```

The `generate` command only writes to the caller-owned output directory. Use
`conformance` to evaluate the nine repository fixtures in CI.

## Contracts

- [v1 contract](docs/contract-v1.md)
- [bootstrap and release facts](docs/bootstrap.md)
- [v0.1.0 release record](docs/release-v0.1.0.md)
