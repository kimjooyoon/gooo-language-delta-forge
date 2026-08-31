package forge

import (
	"path/filepath"
	"runtime"
	"testing"
)

func fixtureRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func loadFixtureProgram(t *testing.T) Program {
	t.Helper()
	root := fixtureRoot(t)
	program, err := LoadProgram(filepath.Join(root, "examples/language-delta-forge-v1/main.gooo"), filepath.Join(root, "contracts/denominator-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func TestGenerateClosedExactBundle(t *testing.T) {
	root := fixtureRoot(t)
	input, err := LoadInput(filepath.Join(root, "fixtures/input-closed.json"))
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := Generate(loadFixtureProgram(t), input)
	if err != nil {
		t.Fatal(err)
	}
	if candidate.Decision != StateClosed || candidate.Target.ResolutionLevel != "FIELD" {
		t.Fatalf("unexpected candidate: %+v", candidate)
	}
	if candidate.Counts.AddedCells != 2 || candidate.Counts.RetiredCells != 1 || candidate.Counts.SplitCells != 1 {
		t.Fatalf("unexpected exact counts: %+v", candidate.Counts)
	}
	if candidate.Improvement.State != StateClosed {
		t.Fatalf("exact pair did not close improvement: %+v", candidate.Improvement)
	}
}

func TestGenerateUnknownPreservesTuple(t *testing.T) {
	root := fixtureRoot(t)
	input, err := LoadInput(filepath.Join(root, "fixtures/input-unknown.json"))
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := Generate(loadFixtureProgram(t), input)
	if err != nil {
		t.Fatal(err)
	}
	if candidate.Decision != StateUnknown || !candidate.Claim.HasUnknownTuple() {
		t.Fatalf("unknown tuple was not preserved: %+v", candidate.Claim)
	}
	if candidate.Target.ResolutionLevel != "CONCEPT" || candidate.Target.PredicateID != "" || candidate.Target.FieldID != "" {
		t.Fatalf("unknown did not lower target resolution: %+v", candidate.Target)
	}
	if candidate.Improvement.State != StateUnknown || !candidate.Improvement.HasUnknownTuple() {
		t.Fatalf("missing exact pair was not unknown: %+v", candidate.Improvement)
	}
}

func TestRefutedReceiptTakesPrecedence(t *testing.T) {
	root := fixtureRoot(t)
	input, err := LoadInput(filepath.Join(root, "fixtures/input-refuted.json"))
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := Generate(loadFixtureProgram(t), input)
	if err != nil {
		t.Fatal(err)
	}
	if candidate.Decision != StateRefuted || candidate.Claim.State != StateRefuted {
		t.Fatalf("refutation did not take precedence: %+v", candidate)
	}
}
