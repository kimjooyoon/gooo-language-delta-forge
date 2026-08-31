# Language Delta Forge v1 contract

## Authority and boundary

The forge consumes two read-only inputs: an immutable Gooo release bundle and
one failure observation with zero or more related `UNKNOWN` or `REFUTED`
receipts. It writes only to a caller-owned output directory. The output is a
candidate bundle, not a source patch, merge request, adoption transaction, or
protected-core mutation.

The `.gooo` file declares the forge's fixed denominator and proof vocabulary.
It is not a source of candidate language meaning. Candidate meaning is copied
from the structured direct-cause evidence in the failure receipt and checked
against the release graph. If that evidence is absent, the result is a lower-
resolution typed `UNKNOWN`.

## Resolution precedence

`REFUTED` precedes `UNKNOWN`, which precedes a closed candidate. A related
`REFUTED` receipt is a known counterexample and blocks candidate adoption even
when a direct cause is present. An `UNKNOWN` receipt preserves all six fields:

`stage`, `step`, `reason`, `unknown_class`, `next_operation`, `blocked_by`.

## Exact candidate contents

Every emitted bundle contains:

- `source_digest` from the immutable release and `.gooo` bytes;
- `baseline_graph_digest` from the immutable release;
- `delta_digest` over the canonical semantic graph delta;
- exact target concept, predicate, and field IDs;
- causal frontier edges supplied by the receipt;
- `added_cells`, `retired_cells`, and `split_cells` with exact counts;
- a rollback delta;
- expected positive and negative conformance cases;
- a test manifest with no score or improvement rate;
- an independent-consumer receipt bound to the same source, baseline, and
  delta digests; and
- an `improvement` claim.

`improvement` is `UNKNOWN` with a typed six-field claim unless an exact
before/after pair is present. No score, rate, ranking, or inferred improvement
number is emitted anywhere in the schema.

## Denominator

The fixed denominator has 18 cells: each of FOUNDATION, COHERENCE, and
REGRESSION occurs six times; each of DRIVER, OUTCOME, and GUARDRAIL occurs six
times; and every proof/indicator pair occurs exactly twice. Repository
conformance supplies nine cases: three `CLOSED`, three `UNKNOWN`, and three
`REFUTED`, with each proof choice and each indicator class occurring three
times. This keeps the Munchausen proof choices separate from the outcome and
guardrail indicators.

## Reproducibility

The CI artifact records inventory (excluding root README), output counts,
memory, compile/build/test/conformance durations, test execution count, and
reuse count. Local tests remain intentionally unexecuted.
