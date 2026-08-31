package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-language-delta-forge/internal/forge"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "generate":
		err = generate(os.Args[2:])
	case "conformance":
		err = conformance(os.Args[2:])
	case "inventory":
		err = inventory(os.Args[2:])
	case "help", "--help", "-h":
		usage()
		return
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate(args []string) error {
	flags := flag.NewFlagSet("generate", flag.ContinueOnError)
	programPath := flags.String("program", "", "path to the .gooo forge program")
	denominatorPath := flags.String("denominator", "", "path to the fixed denominator")
	inputPath := flags.String("input", "", "path to an immutable release and receipt input bundle")
	outputDir := flags.String("output", "", "empty caller-owned output directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *programPath == "" || *denominatorPath == "" || *inputPath == "" || *outputDir == "" {
		return fmt.Errorf("generate requires --program, --denominator, --input, and --output")
	}
	program, err := forge.LoadProgram(*programPath, *denominatorPath)
	if err != nil {
		return err
	}
	input, err := forge.LoadInput(*inputPath)
	if err != nil {
		return err
	}
	candidate, err := forge.Generate(program, input)
	if err != nil {
		return err
	}
	if err := forge.PrepareOutput(*outputDir); err != nil {
		return err
	}
	if err := forge.WriteJSONFile(filepath.Join(*outputDir, "candidate-bundle.json"), candidate); err != nil {
		return err
	}
	return printJSON(map[string]any{"decision": candidate.Decision, "candidate_digest": candidate.CandidateDigest})
}

func conformance(args []string) error {
	flags := flag.NewFlagSet("conformance", flag.ContinueOnError)
	programPath := flags.String("program", "", "path to the .gooo forge program")
	denominatorPath := flags.String("denominator", "", "path to the fixed denominator")
	manifestPath := flags.String("manifest", "", "path to the conformance manifest")
	outputDir := flags.String("output", "", "empty caller-owned output directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *programPath == "" || *denominatorPath == "" || *manifestPath == "" || *outputDir == "" {
		return fmt.Errorf("conformance requires --program, --denominator, --manifest, and --output")
	}
	program, err := forge.LoadProgram(*programPath, *denominatorPath)
	if err != nil {
		return err
	}
	report, err := forge.RunConformance(program, *manifestPath, *outputDir)
	if err != nil {
		return err
	}
	if err := printJSON(report); err != nil {
		return err
	}
	if report.Decision != forge.StateClosed {
		return fmt.Errorf("conformance decision is %s", report.Decision)
	}
	return nil
}

func inventory(args []string) error {
	flags := flag.NewFlagSet("inventory", flag.ContinueOnError)
	root := flags.String("root", ".", "repository root")
	output := flags.String("output", "", "optional output JSON path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	report, err := forge.Inventory(*root)
	if err != nil {
		return err
	}
	if *output != "" {
		if err := forge.WriteJSONFile(*output, report); err != nil {
			return err
		}
	}
	return printJSON(report)
}

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func usage() {
	fmt.Fprintln(os.Stderr, "gooo-language-delta-forge generate|conformance|inventory")
}
