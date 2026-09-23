package forge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadInputRejectsUnknownAndTrailingJSON(t *testing.T) {
	cases := map[string]string{
		"unknown":  `{"unexpected":true}`,
		"trailing": `{} {"extra":true}`,
	}
	for name, payload := range cases {
		path := filepath.Join(t.TempDir(), name+".json")
		if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadInput(path); err == nil {
			t.Fatalf("%s input was accepted", name)
		}
	}
}
