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
