package forge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func LoadConformanceManifest(path string) (ConformanceManifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ConformanceManifest{}, fmt.Errorf("read conformance manifest: %w", err)
	}
	var manifest ConformanceManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return ConformanceManifest{}, fmt.Errorf("decode conformance manifest: %w", err)
	}
	if manifest.Schema != ConformanceManifestSchema || len(manifest.Cases) == 0 {
		return ConformanceManifest{}, fmt.Errorf("conformance manifest is empty or malformed")
	}
	seen := map[string]bool{}
	for _, testCase := range manifest.Cases {
		if testCase.ID == "" || seen[testCase.ID] || testCase.InputPath == "" || (testCase.ExpectedDecision != StateClosed && testCase.ExpectedDecision != StateUnknown && testCase.ExpectedDecision != StateRefuted) || !contains([]string{ProofFoundation, ProofCoherence, ProofRegression}, testCase.ProofChoice) || !contains([]string{IndicatorDriver, IndicatorOutcome, IndicatorGuardrail}, testCase.IndicatorClass) {
			return ConformanceManifest{}, fmt.Errorf("malformed conformance case %q", testCase.ID)
		}
		seen[testCase.ID] = true
	}
	return manifest, nil
}

func RunConformance(program Program, manifestPath, outputDir string) (ConformanceReport, error) {
	manifest, err := LoadConformanceManifest(manifestPath)
	if err != nil {
		return ConformanceReport{}, err
	}
	if err := prepareEmptyOutput(outputDir); err != nil {
		return ConformanceReport{}, err
	}
	baseDir := filepath.Dir(manifestPath)
	report := ConformanceReport{Schema: ConformanceReportSchema, Decision: StateClosed, Denominator: program.Denominator.CellCount, Cases: []ConformanceCaseResult{}}
	proofCounts := map[string]int{}
	indicatorCounts := map[string]int{}
	for _, spec := range manifest.Cases {
		input, loadErr := LoadInput(filepath.Join(baseDir, spec.InputPath))
		result := ConformanceCaseResult{ID: spec.ID, ExpectedDecision: spec.ExpectedDecision, ProofChoice: spec.ProofChoice, IndicatorClass: spec.IndicatorClass}
		proofCounts[spec.ProofChoice]++
		indicatorCounts[spec.IndicatorClass]++
		if loadErr == nil {
			candidate, generateErr := Generate(program, input)
			if generateErr == nil {
				result.ObservedDecision = candidate.Decision
				result.CandidateDigest = candidate.CandidateDigest
				result.Pass = candidate.Decision == spec.ExpectedDecision
				if err := writeJSON(filepath.Join(outputDir, "candidates", spec.ID+".json"), candidate); err != nil {
					return ConformanceReport{}, err
				}
			} else {
				result.ObservedDecision = StateRefuted
				result.Pass = false
			}
		} else {
			result.ObservedDecision = StateRefuted
			result.Pass = false
		}
		report.Cases = append(report.Cases, result)
		if !result.Pass {
			report.Summary.Failures++
		}
		switch result.ObservedDecision {
		case StateClosed:
			report.Summary.Closed++
		case StateUnknown:
			report.Summary.Unknown++
		case StateRefuted:
			report.Summary.Refuted++
		}
	}
	report.Summary.Total = len(report.Cases)
	if report.Summary.Closed < 3 || report.Summary.Unknown < 3 || report.Summary.Refuted < 3 {
		report.Summary.Failures++
	}
	for _, proof := range []string{ProofFoundation, ProofCoherence, ProofRegression} {
		if proofCounts[proof] != 3 {
			report.Summary.Failures++
		}
	}
	for _, indicator := range []string{IndicatorDriver, IndicatorOutcome, IndicatorGuardrail} {
		if indicatorCounts[indicator] != 3 {
			report.Summary.Failures++
		}
	}
	if report.Summary.Failures > 0 {
		report.Decision = StateRefuted
	}
	if err := writeJSON(filepath.Join(outputDir, "conformance-report.json"), report); err != nil {
		return ConformanceReport{}, err
	}
	return report, nil
}

func prepareEmptyOutput(path string) error {
	if path == "" {
		return fmt.Errorf("caller-owned output path is required")
	}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0o755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("output path is not a directory")
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("output directory must be empty: %s", path)
	}
	return nil
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}

// WriteJSONFile is the small output boundary used by the CLI and CI. It only
// writes caller-owned paths; it never receives or mutates an input path.
func WriteJSONFile(path string, value any) error { return writeJSON(path, value) }

func PrepareOutput(path string) error { return prepareEmptyOutput(path) }
