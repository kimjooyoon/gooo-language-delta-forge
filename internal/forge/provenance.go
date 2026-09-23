package forge

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	ProvenanceChainSchema  = "gooo/language-delta-forge/provenance-chain/v1"
	ProvenanceChainVersion = "v1"
)

// ProvenanceStage is a deterministic link in the candidate's evidence chain.
// BuilderID is an identity label, not a credential or an authorization grant.
type ProvenanceStage struct {
	Name          string `json:"name"`
	State         string `json:"state"`
	BuilderID     string `json:"builder_id"`
	InputDigest   string `json:"input_digest,omitempty"`
	ArtifactDigest string `json:"artifact_digest"`
	ParentDigest  string `json:"parent_digest,omitempty"`
	Digest        string `json:"digest"`
}

type ProvenanceChain struct {
	Schema       string             `json:"schema"`
	Version      string             `json:"version"`
	SubjectDigest string             `json:"subject_digest"`
	RootDigest   string             `json:"root_digest"`
	TipDigest    string             `json:"tip_digest"`
	Stages       []ProvenanceStage  `json:"stages"`
	Digest       string             `json:"digest"`
}

// BuildProvenanceChain binds a candidate to its source, transformation,
// observation, and decision without granting mutation or merge authority.
func BuildProvenanceChain(candidate CandidateBundle, builderID string) (ProvenanceChain, error) {
	if candidate.SourceDigest == "" || candidate.BaselineGraphDigest == "" || candidate.CandidateDigest == "" {
		return ProvenanceChain{}, fmt.Errorf("candidate is missing a required digest")
	}
	if err := validateBuilderID(builderID); err != nil {
		return ProvenanceChain{}, err
	}

	deltaDigest := candidate.DeltaDigest
	if deltaDigest == "" {
		deltaDigest = mustDigestValue(struct {
			SourceDigest         string `json:"source_digest"`
			BaselineGraphDigest  string `json:"baseline_graph_digest"`
		}{candidate.SourceDigest, candidate.BaselineGraphDigest})
	}
	observationDigest := mustDigestValue(struct {
		Decision      string `json:"decision"`
		ReceiptDigest string `json:"receipt_digest"`
	}{candidate.Decision, candidate.IndependentConsumerReceipt.ReceiptDigest})

	stages := []ProvenanceStage{
		{Name: "origin", State: "BOUND", BuilderID: builderID, ArtifactDigest: candidate.SourceDigest},
		{Name: "transformation", State: "BOUND", BuilderID: builderID, InputDigest: candidate.SourceDigest, ArtifactDigest: deltaDigest},
		{Name: "observation", State: candidate.Decision, BuilderID: builderID, InputDigest: deltaDigest, ArtifactDigest: observationDigest},
		{Name: "decision", State: candidate.Decision, BuilderID: builderID, InputDigest: observationDigest, ArtifactDigest: candidate.CandidateDigest},
	}
	for index := range stages {
		if index > 0 {
			stages[index].ParentDigest = stages[index-1].Digest
		}
		stages[index].Digest = mustDigestValue(stages[index])
	}

	chain := ProvenanceChain{
		Schema:        ProvenanceChainSchema,
		Version:       ProvenanceChainVersion,
		SubjectDigest: candidate.CandidateDigest,
		RootDigest:    stages[0].Digest,
		TipDigest:     stages[len(stages)-1].Digest,
		Stages:        stages,
	}
	chain.Digest = mustDigestValue(chain)
	if err := ValidateProvenanceChain(chain); err != nil {
		return ProvenanceChain{}, err
	}
	return chain, nil
}

func ValidateProvenanceChain(chain ProvenanceChain) error {
	if chain.Schema != ProvenanceChainSchema || chain.Version != ProvenanceChainVersion {
		return fmt.Errorf("unsupported provenance chain schema")
	}
	if chain.SubjectDigest == "" || chain.RootDigest == "" || chain.TipDigest == "" || chain.Digest == "" {
		return fmt.Errorf("provenance chain is missing a required digest")
	}
	if len(chain.Stages) != 4 {
		return fmt.Errorf("provenance chain requires four stages")
	}
	wantNames := []string{"origin", "transformation", "observation", "decision"}
	for index := range chain.Stages {
		stage := chain.Stages[index]
		if stage.Name != wantNames[index] || stage.State == "" || stage.BuilderID == "" || stage.ArtifactDigest == "" || stage.Digest == "" {
			return fmt.Errorf("invalid provenance stage %d", index)
		}
		if err := validateBuilderID(stage.BuilderID); err != nil {
			return err
		}
		if index == 0 {
			if stage.ParentDigest != "" || stage.InputDigest != "" {
				return fmt.Errorf("origin stage must not have a parent or input")
			}
		} else {
			if stage.ParentDigest != chain.Stages[index-1].Digest || stage.InputDigest == "" {
				return fmt.Errorf("provenance stage %d is not linked to its predecessor", index)
			}
		}
		stageDigest := stage.Digest
		stage.Digest = ""
		if mustDigestValue(stage) != stageDigest {
			return fmt.Errorf("provenance stage %d digest mismatch", index)
		}
	}
	if chain.RootDigest != chain.Stages[0].Digest || chain.TipDigest != chain.Stages[len(chain.Stages)-1].Digest {
		return fmt.Errorf("provenance root or tip mismatch")
	}
	withoutDigest := chain
	withoutDigest.Digest = ""
	if mustDigestValue(withoutDigest) != chain.Digest {
		return fmt.Errorf("provenance chain digest mismatch")
	}
	return nil
}

func validateBuilderID(builderID string) error {
	parsed, err := url.Parse(builderID)
	if err != nil || parsed.Scheme != "spiffe" || parsed.Host == "" || strings.Trim(parsed.Path, "/") == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("builder_id must be a path-bearing spiffe URI")
	}
	return nil
}
