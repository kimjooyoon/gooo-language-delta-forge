package forge

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func DigestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DigestValue(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return DigestBytes(raw), nil
}

func mustDigestValue(value any) string {
	digest, err := DigestValue(value)
	if err != nil {
		panic(err)
	}
	return digest
}

func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || value[:len("sha256:")] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}

func validateDigest(name, value string) error {
	if !validDigest(value) {
		return fmt.Errorf("%s must be a sha256 digest", name)
	}
	return nil
}

func snapshotFromGraph(graph SemanticGraph) GraphSnapshot {
	return GraphSnapshot{
		Digest:     graph.GraphDigest,
		Concepts:   cloneConcepts(graph.Concepts),
		Predicates: clonePredicates(graph.Predicates),
		Edges:      append([]GraphEdge(nil), graph.Edges...),
		Cells:      append([]GraphCell(nil), graph.Cells...),
	}
}

func digestSnapshot(snapshot GraphSnapshot) string {
	unsigned := snapshot
	unsigned.Digest = ""
	return mustDigestValue(unsigned)
}

func digestIdentity(identity ReleaseIdentity) string {
	unsigned := identity
	unsigned.Digest = ""
	return mustDigestValue(unsigned)
}

func digestReceipt(receipt ObservationReceipt) string {
	unsigned := receipt
	unsigned.Digest = ""
	return mustDigestValue(unsigned)
}

func digestIndependent(receipt IndependentConsumerReceipt) string {
	unsigned := receipt
	unsigned.ReceiptDigest = ""
	return mustDigestValue(unsigned)
}

func digestCandidate(bundle CandidateBundle) string {
	unsigned := bundle
	unsigned.CandidateDigest = ""
	return mustDigestValue(unsigned)
}

func cloneConcepts(values []Concept) []Concept {
	return append([]Concept(nil), values...)
}

func clonePredicates(values []Predicate) []Predicate {
	result := make([]Predicate, len(values))
	for index, predicate := range values {
		result[index] = predicate
		result[index].Fields = append([]Field(nil), predicate.Fields...)
	}
	return result
}

func cloneFrontier(values []FrontierEdge) []FrontierEdge {
	return append([]FrontierEdge(nil), values...)
}

func cloneCells(values []GraphCell) []GraphCell {
	return append([]GraphCell(nil), values...)
}

func cloneSplits(values []SplitCell) []SplitCell {
	result := make([]SplitCell, len(values))
	for index, value := range values {
		result[index] = value
		result[index].AddedCellIDs = append([]string(nil), value.AddedCellIDs...)
	}
	return result
}

func cloneCases(values []ConformanceCase) []ConformanceCase {
	return append([]ConformanceCase(nil), values...)
}
