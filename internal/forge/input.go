package forge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func LoadInput(path string) (InputBundle, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return InputBundle{}, fmt.Errorf("read input: %w", err)
	}
	var input InputBundle
	if err := json.Unmarshal(raw, &input); err != nil {
		return InputBundle{}, fmt.Errorf("decode input: %w", err)
	}
	if input.Release.Schema == "" && input.ReleasePath != "" {
		releaseRaw, readErr := os.ReadFile(filepath.Join(filepath.Dir(path), input.ReleasePath))
		if readErr != nil {
			return InputBundle{}, fmt.Errorf("read immutable release: %w", readErr)
		}
		if err := json.Unmarshal(releaseRaw, &input.Release); err != nil {
			return InputBundle{}, fmt.Errorf("decode immutable release: %w", err)
		}
	}
	if err := ValidateInput(input); err != nil {
		return InputBundle{}, err
	}
	return input, nil
}

func ValidateInput(input InputBundle) error {
	if input.Schema != InputSchema {
		return fmt.Errorf("input schema mismatch")
	}
	if err := validateRelease(input.Release); err != nil {
		return err
	}
	if err := validateReceipt(input.Failure, input.Release, true); err != nil {
		return err
	}
	if input.Failure.State != StateFailure {
		return fmt.Errorf("input failure receipt must have state FAILURE")
	}
	seen := map[string]bool{input.Failure.ID: true}
	for _, receipt := range input.Receipts {
		if seen[receipt.ID] {
			return fmt.Errorf("duplicate receipt %q", receipt.ID)
		}
		seen[receipt.ID] = true
		if receipt.RelatedTo != input.Failure.ID {
			return fmt.Errorf("receipt %q does not relate to failure %q", receipt.ID, input.Failure.ID)
		}
		if receipt.State != StateUnknown && receipt.State != StateRefuted {
			return fmt.Errorf("receipt %q must be UNKNOWN or REFUTED", receipt.ID)
		}
		if err := validateReceipt(receipt, input.Release, false); err != nil {
			return err
		}
	}
	return nil
}

func validateRelease(release ImmutableRelease) error {
	if release.Schema != ReleaseSchema || release.Version != "v1" {
		return fmt.Errorf("immutable release schema mismatch")
	}
	identity := release.Identity
	if identity.Repository == "" || identity.Tag == "" || identity.Commit == "" {
		return fmt.Errorf("immutable release identity is incomplete")
	}
	if err := validateDigest("release source_digest", identity.SourceDigest); err != nil {
		return err
	}
	if err := validateDigest("release graph_digest", identity.GraphDigest); err != nil {
		return err
	}
	if err := validateDigest("release digest", identity.Digest); err != nil {
		return err
	}
	if digestIdentity(identity) != identity.Digest {
		return fmt.Errorf("immutable release identity digest mismatch")
	}
	graph := release.Graph
	if graph.Schema != "gooo/semantic-graph/v1" || graph.Version != "v1" || graph.GraphDigest != identity.GraphDigest {
		return fmt.Errorf("immutable release graph binding mismatch")
	}
	if digestSnapshot(snapshotFromGraph(graph)) != graph.GraphDigest {
		return fmt.Errorf("immutable release graph digest mismatch")
	}
	if err := validateGraph(graph); err != nil {
		return err
	}
	return nil
}

func validateGraph(graph SemanticGraph) error {
	concepts := map[string]bool{}
	for _, concept := range graph.Concepts {
		if concept.ID == "" || concept.Name == "" || concepts[concept.ID] {
			return fmt.Errorf("malformed or duplicate concept")
		}
		concepts[concept.ID] = true
	}
	predicates := map[string]Predicate{}
	fields := map[string]Field{}
	for _, predicate := range graph.Predicates {
		if predicate.ID == "" || predicate.Name == "" || predicates[predicate.ID].ID != "" || !concepts[predicate.ConceptID] {
			return fmt.Errorf("malformed predicate %q", predicate.ID)
		}
		predicates[predicate.ID] = predicate
		for _, field := range predicate.Fields {
			if field.ID == "" || field.Name == "" || field.Type == "" || fields[field.ID].ID != "" {
				return fmt.Errorf("malformed or duplicate field %q", field.ID)
			}
			fields[field.ID] = field
		}
	}
	edges := map[string]bool{}
	for _, edge := range graph.Edges {
		if edge.ID == "" || edges[edge.ID] || edge.From == "" || edge.To == "" || edge.Kind == "" || predicates[edge.PredicateID].ID == "" {
			return fmt.Errorf("malformed graph edge %q", edge.ID)
		}
		edges[edge.ID] = true
	}
	cells := map[string]bool{}
	for _, cell := range graph.Cells {
		predicate, ok := predicates[cell.PredicateID]
		if cell.ID == "" || cells[cell.ID] || !ok || cell.ConceptID != predicate.ConceptID || !hasField(predicate.Fields, cell.FieldID) || cell.Relation == "" || cell.Constraint == "" {
			return fmt.Errorf("malformed graph cell %q", cell.ID)
		}
		cells[cell.ID] = true
	}
	return nil
}

func validateReceipt(receipt ObservationReceipt, release ImmutableRelease, primary bool) error {
	if receipt.Schema != ReceiptSchema || receipt.ID == "" || receipt.SourceDigest != release.Identity.SourceDigest || receipt.GraphDigest != release.Identity.GraphDigest {
		return fmt.Errorf("receipt %q has invalid release binding", receipt.ID)
	}
	if receipt.Digest != "" && receipt.Digest != digestReceipt(receipt) {
		return fmt.Errorf("receipt %q digest mismatch", receipt.ID)
	}
	if primary && receipt.RelatedTo != "" {
		return fmt.Errorf("primary failure cannot have related_to")
	}
	if receipt.State == StateUnknown {
		claim := Claim{State: StateUnknown, Stage: receipt.Stage, Step: receipt.Step, Reason: receipt.Reason, UnknownClass: receipt.UnknownClass, NextOperation: receipt.NextOperation, BlockedBy: receipt.BlockedBy}
		if !claim.HasUnknownTuple() {
			return fmt.Errorf("receipt %q has incomplete UNKNOWN tuple", receipt.ID)
		}
	}
	if receipt.State == StateRefuted && receipt.Reason == "" && receipt.Counterexample == "" && receipt.DirectCause == nil {
		return fmt.Errorf("receipt %q has no refutation evidence", receipt.ID)
	}
	return nil
}

func hasField(fields []Field, fieldID string) bool {
	for _, field := range fields {
		if field.ID == fieldID {
			return true
		}
	}
	return false
}

func predicateByID(graph SemanticGraph, id string) (Predicate, bool) {
	for _, predicate := range graph.Predicates {
		if predicate.ID == id {
			return predicate, true
		}
	}
	return Predicate{}, false
}

func targetExists(graph SemanticGraph, target Target) bool {
	predicate, ok := predicateByID(graph, target.PredicateID)
	return ok && predicate.ConceptID == target.ConceptID && hasField(predicate.Fields, target.FieldID)
}
