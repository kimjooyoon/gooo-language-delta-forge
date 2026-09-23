package forge

import "testing"

func TestValidateDenominatorRequiresVersion(t *testing.T) {
	denominator := loadFixtureProgram(t).Denominator
	denominator.Version = "v0"
	if err := ValidateDenominator(denominator); err == nil {
		t.Fatal("denominator with unsupported version was accepted")
	}
}
