package forge

import "testing"

func TestParseProgramRejectsDuplicateDeclarationKeys(t *testing.T) {
	_, err := parseProgram([]byte("gooo language_delta_forge v1\ndenominator id=first cell_count=18 id=second\n"))
	if err == nil {
		t.Fatal("expected duplicate declaration key to be rejected")
	}
}

func TestParseProgramRejectsMissingHeader(t *testing.T) {
	_, err := parseProgram([]byte("denominator id=delta cell_count=18\ncell ordinal=1 id=cell proof=FOUNDATION indicator=DRIVER stage=parse step=load\n"))
	if err == nil {
		t.Fatal("expected missing program header to be rejected")
	}
}

func TestParseProgramRejectsUnknownDeclaration(t *testing.T) {
	_, err := parseProgram([]byte("gooo language_delta_forge v1\nunknown value=ignored\n"))
	if err == nil {
		t.Fatal("expected unknown program declaration to be rejected")
	}
}

func TestParseProgramRejectsDuplicateHeaders(t *testing.T) {
	_, err := parseProgram([]byte("gooo language_delta_forge v1\ngooo language_delta_forge v1\n"))
	if err == nil {
		t.Fatal("expected duplicate program header to be rejected")
	}
}
