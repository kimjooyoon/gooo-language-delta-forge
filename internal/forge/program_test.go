package forge

import "testing"

func TestParseProgramRejectsDuplicateDeclarationKeys(t *testing.T) {
	_, err := parseProgram([]byte("gooo language_delta_forge v1\ndenominator id=first cell_count=18 id=second\n"))
	if err == nil {
		t.Fatal("expected duplicate declaration key to be rejected")
	}
}
