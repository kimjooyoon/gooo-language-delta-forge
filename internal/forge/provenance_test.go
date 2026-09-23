package forge

import "testing"

func TestProvenanceChainRejectsStageTampering(t *testing.T) {
	candidate := CandidateBundle{
		Decision:             StateClosed,
		SourceDigest:         "source-digest",
		BaselineGraphDigest:  "baseline-digest",
		CandidateDigest:      "candidate-digest",
		IndependentConsumerReceipt: IndependentConsumerReceipt{ReceiptDigest: "receipt-digest"},
	}
	chain, err := BuildProvenanceChain(candidate, "spiffe://github-actions/gooo-language-delta-forge")
	if err != nil {
		t.Fatal(err)
	}
	chain.Stages[2].ArtifactDigest = "tampered"
	if err := ValidateProvenanceChain(chain); err == nil {
		t.Fatal("expected tampered provenance stage to be rejected")
	}
}

func TestProvenanceChainRequiresPathBearingSPIFFEBuilder(t *testing.T) {
	candidate := CandidateBundle{Decision: StateUnknown, SourceDigest: "source", BaselineGraphDigest: "baseline", CandidateDigest: "candidate"}
	if _, err := BuildProvenanceChain(candidate, "spiffe://github-actions"); err == nil {
		t.Fatal("expected builder identity without a path to be rejected")
	}
}

func TestProvenanceChainRejectsMixedBuilderIdentity(t *testing.T) {
	candidate := CandidateBundle{Decision: StateClosed, SourceDigest: "source", BaselineGraphDigest: "baseline", CandidateDigest: "candidate"}
	chain, err := BuildProvenanceChain(candidate, "spiffe://github-actions/gooo-language-delta-forge")
	if err != nil {
		t.Fatal(err)
	}
	chain.Stages[3].BuilderID = "spiffe://other-builder/gooo-language-delta-forge"
	if err := ValidateProvenanceChain(chain); err == nil {
		t.Fatal("expected mixed builder identities to be rejected")
	}
}

func TestProvenanceChainRejectsStateDriftBetweenObservationAndDecision(t *testing.T) {
	candidate := CandidateBundle{Decision: StateClosed, SourceDigest: "source", BaselineGraphDigest: "baseline", CandidateDigest: "candidate"}
	chain, err := BuildProvenanceChain(candidate, "spiffe://github-actions/gooo-language-delta-forge")
	if err != nil {
		t.Fatal(err)
	}
	chain.Stages[3].State = string(StateRefuted)
	if err := ValidateProvenanceChain(chain); err == nil {
		t.Fatal("expected observation and decision state drift to be rejected")
	}
}
