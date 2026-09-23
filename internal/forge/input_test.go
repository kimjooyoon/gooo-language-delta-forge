package forge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadInputRejectsTrailingJSONValue(t *testing.T) {
	root := fixtureRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "fixtures/input-closed.json"))
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, []byte("\n{}\n")...)
	path := filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadInput(path); err == nil {
		t.Fatal("LoadInput accepted a trailing JSON value")
	}
}
