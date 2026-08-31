package forge

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func LoadProgram(programPath, denominatorPath string) (Program, error) {
	raw, err := os.ReadFile(programPath)
	if err != nil {
		return Program{}, fmt.Errorf("read program: %w", err)
	}
	program, err := parseProgram(raw)
	if err != nil {
		return Program{}, err
	}
	program.SourceDigest = DigestBytes(raw)
	denominatorRaw, err := os.ReadFile(denominatorPath)
	if err != nil {
		return Program{}, fmt.Errorf("read denominator: %w", err)
	}
	var denominator Denominator
	if err := json.Unmarshal(denominatorRaw, &denominator); err != nil {
		return Program{}, fmt.Errorf("decode denominator: %w", err)
	}
	if err := ValidateDenominator(denominator); err != nil {
		return Program{}, err
	}
	if err := validateProgramDenominator(program, denominator); err != nil {
		return Program{}, err
	}
	program.Denominator = denominator
	return program, nil
}

func parseProgram(raw []byte) (Program, error) {
	var err error
	program := Program{
		Schema:        ProgramSchema,
		Version:       "v1",
		Authority:     Authority{},
		Precedence:    []string{"REFUTED", "UNKNOWN", "CLOSED"},
		UnknownFields: []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"},
	}
	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		switch fields[0] {
		case "gooo":
			if len(fields) != 3 || fields[1] != "language_delta_forge" || fields[2] != "v1" {
				return Program{}, fmt.Errorf("invalid program declaration")
			}
		case "denominator":
			values := keyValues(fields[1:])
			program.Denominator = Denominator{Schema: DenominatorSchema, ID: values["id"]}
			program.Denominator.CellCount, err = strconv.Atoi(values["cell_count"])
			if err != nil {
				return Program{}, fmt.Errorf("invalid denominator cell_count")
			}
		case "authority":
			values := keyValues(fields[1:])
			program.Authority.RepositoryWrites, err = parseInt(values, "repository_writes")
			if err != nil {
				return Program{}, err
			}
			program.Authority.LocalTestExecutions, err = parseInt(values, "local_test_executions")
			if err != nil {
				return Program{}, err
			}
			program.Authority.ProtectedCoreAdoption, err = parseInt(values, "protected_core_adoption")
			if err != nil {
				return Program{}, err
			}
			program.Authority.MergeAuthority, err = parseInt(values, "merge_authority")
			if err != nil {
				return Program{}, err
			}
		case "precedence":
			if len(fields) != 2 {
				return Program{}, fmt.Errorf("invalid precedence declaration")
			}
			program.Precedence = strings.Split(fields[1], ">")
		case "unknown_fields":
			if len(fields) != 2 {
				return Program{}, fmt.Errorf("invalid unknown_fields declaration")
			}
			program.UnknownFields = strings.Split(fields[1], ",")
		case "cell":
			values := keyValues(fields[1:])
			ordinal, parseErr := strconv.Atoi(values["ordinal"])
			if parseErr != nil {
				return Program{}, fmt.Errorf("invalid cell ordinal: %w", parseErr)
			}
			program.Cells = append(program.Cells, ProgramCell{
				Ordinal: ordinal, ID: values["id"], ProofChoice: values["proof"],
				IndicatorClass: values["indicator"], Stage: values["stage"], Step: values["step"],
			})
		}
	}
	if program.Denominator.ID == "" || program.Denominator.CellCount == 0 || len(program.Cells) == 0 {
		return Program{}, fmt.Errorf("program does not declare a denominator and cells")
	}
	if program.Authority.RepositoryWrites != 0 || program.Authority.LocalTestExecutions != 0 ||
		program.Authority.ProtectedCoreAdoption != 0 || program.Authority.MergeAuthority != 0 {
		return Program{}, fmt.Errorf("program escalates authority")
	}
	return program, nil
}

func keyValues(fields []string) map[string]string {
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) == 2 {
			values[parts[0]] = parts[1]
		}
	}
	return values
}

func parseInt(values map[string]string, key string) (int, error) {
	value, err := strconv.Atoi(values[key])
	if err != nil {
		return 0, fmt.Errorf("invalid authority %s", key)
	}
	return value, nil
}

func ValidateDenominator(denominator Denominator) error {
	if denominator.Schema != DenominatorSchema || denominator.ID == "" || denominator.CellCount != 18 || !denominator.Fixed {
		return fmt.Errorf("denominator is not fixed at 18 cells")
	}
	if !equalStrings(denominator.ProofChoices, []string{ProofFoundation, ProofCoherence, ProofRegression}) ||
		!equalStrings(denominator.IndicatorClasses, []string{IndicatorDriver, IndicatorOutcome, IndicatorGuardrail}) ||
		denominator.CellsPerProofChoice != 6 || denominator.CellsPerIndicatorClass != 6 {
		return fmt.Errorf("denominator proof and indicator vocabulary is not fixed")
	}
	if !equalStrings(denominator.UnknownFields, []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"}) ||
		!equalStrings(denominator.Precedence, []string{StateRefuted, StateUnknown, StateClosed}) {
		return fmt.Errorf("denominator precedence or UNKNOWN tuple is invalid")
	}
	return nil
}

func validateProgramDenominator(program Program, denominator Denominator) error {
	if program.Denominator.ID != denominator.ID || program.Denominator.CellCount != denominator.CellCount {
		return fmt.Errorf("program and denominator identity mismatch")
	}
	if program.Precedence[0] != StateRefuted || program.Precedence[1] != StateUnknown || program.Precedence[2] != StateClosed {
		return fmt.Errorf("program precedence mismatch")
	}
	if len(program.UnknownFields) != 6 || !equalStrings(program.UnknownFields, denominator.UnknownFields) {
		return fmt.Errorf("program UNKNOWN tuple mismatch")
	}
	if len(program.Cells) != denominator.CellCount {
		return fmt.Errorf("program cell count mismatch")
	}
	seenID := map[string]bool{}
	seenOrdinal := map[int]bool{}
	proofCounts := map[string]int{}
	indicatorCounts := map[string]int{}
	pairCounts := map[string]int{}
	for _, cell := range program.Cells {
		if cell.Ordinal < 1 || cell.Ordinal > denominator.CellCount || seenOrdinal[cell.Ordinal] || cell.ID == "" || seenID[cell.ID] {
			return fmt.Errorf("program has duplicate or malformed cell")
		}
		if !contains(denominator.ProofChoices, cell.ProofChoice) || !contains(denominator.IndicatorClasses, cell.IndicatorClass) || cell.Stage == "" || cell.Step == "" {
			return fmt.Errorf("program cell %q has invalid proof or indicator binding", cell.ID)
		}
		seenID[cell.ID] = true
		seenOrdinal[cell.Ordinal] = true
		proofCounts[cell.ProofChoice]++
		indicatorCounts[cell.IndicatorClass]++
		pairCounts[cell.ProofChoice+"/"+cell.IndicatorClass]++
	}
	for _, proof := range denominator.ProofChoices {
		if proofCounts[proof] != denominator.CellsPerProofChoice {
			return fmt.Errorf("proof choice %s is not balanced", proof)
		}
	}
	for _, indicator := range denominator.IndicatorClasses {
		if indicatorCounts[indicator] != denominator.CellsPerIndicatorClass {
			return fmt.Errorf("indicator class %s is not balanced", indicator)
		}
	}
	for _, proof := range denominator.ProofChoices {
		for _, indicator := range denominator.IndicatorClasses {
			if pairCounts[proof+"/"+indicator] != 2 {
				return fmt.Errorf("proof/indicator pair %s/%s is not balanced", proof, indicator)
			}
		}
	}
	return nil
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
