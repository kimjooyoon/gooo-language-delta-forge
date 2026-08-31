package forge

import (
	"fmt"
	"sort"
)

func Generate(program Program, input InputBundle) (CandidateBundle, error) {
	if err := ValidateDenominator(program.Denominator); err != nil {
		return CandidateBundle{}, err
	}
	if err := ValidateInput(input); err != nil {
		return CandidateBundle{}, err
	}
	if err := validateProgramDenominator(program, program.Denominator); err != nil {
		return CandidateBundle{}, err
	}

	bundle := CandidateBundle{
		Schema:               CandidateBundleSchema,
		Version:              "v1",
		Decision:             StateUnknown,
		SourceDigest:         input.Release.Identity.SourceDigest,
		BaselineGraphDigest:  input.Release.Identity.GraphDigest,
		SemanticGraphDelta:   emptyDelta(input.Release.Graph),
		RollbackDelta:        RollbackDelta{RemoveAddedCellIDs: []string{}, RestoreRetiredCells: []GraphCell{}, Unsplit: []SplitCell{}, ExactPair: false},
		AffectedPredicateIDs: []string{},
		TestManifest:         emptyTestManifest(),
		SourceReceipts:       sourceReceiptReferences(input),
		Authority: AuthorityReceipt{
			RepositoryWrites: 0, LocalTestExecutions: 0, ProtectedCoreAdoption: 0,
			MergeAuthority: 0, ReadOnlyInput: true,
		},
		Adoption: AdoptionBoundary{CandidateOnly: true, ProtectedCoreMutation: false, AutomaticMerge: false, SeparateAuthorityStep: true},
	}

	direct := input.Failure.DirectCause
	directValid := false
	if direct != nil {
		if err := validateDirectCause(*direct, input.Release.Graph); err == nil {
			directValid = true
			bindDirectEvidence(&bundle, *direct, input.Release.Graph)
		} else {
			bundle.Decision = StateRefuted
			bundle.Claim = Claim{State: StateRefuted, Stage: "CAUSE", Step: "validate_direct_cause", Reason: "DIRECT_CAUSE_EVIDENCE_INVALID", NextOperation: "RECORD_REFUTED_EVIDENCE", BlockedBy: []string{err.Error()}}
		}
	}

	refuted := relatedReceipt(input, StateRefuted)
	unknown := relatedReceipt(input, StateUnknown)
	switch {
	case refuted != nil:
		bundle.Decision = StateRefuted
		bundle.Claim = refutedClaim(*refuted)
	case bundle.Decision == StateRefuted:
		// A malformed direct-cause assertion is itself a known counterexample.
	case unknown != nil:
		bundle.Decision = StateUnknown
		bundle.Claim = unknownClaim(*unknown)
	case directValid:
		bundle.Decision = StateClosed
		bundle.Claim = Claim{State: StateClosed, Stage: "CANDIDATE", Step: "bind_direct_cause", Reason: "DIRECT_CAUSE_BOUND_FROM_RECEIPT", NextOperation: "HUMAN_REVIEW_CANDIDATE", BlockedBy: []string{}}
	default:
		bundle.Decision = StateUnknown
		bundle.Claim = Claim{State: StateUnknown, Stage: "CAUSE", Step: "resolve_direct_cause", Reason: "DIRECT_CAUSE_NOT_OBSERVED", UnknownClass: "MISSING_DIRECT_CAUSE", NextOperation: "OBTAIN_STRUCTURED_DIRECT_CAUSE", BlockedBy: []string{"failure:" + input.Failure.ID}}
	}

	if !directValid {
		bundle.Target = lowerResolutionTarget(input.Release.Graph, input.Failure.Target)
		bundle.IndependentConsumerReceipt = unknownIndependentReceipt(bundle, input.Failure.ID, bundle.Claim)
	}
	if directValid {
		bundle.IndependentConsumerReceipt = bindIndependentReceipt(bundle.IndependentConsumerReceipt, bundle.SourceDigest, bundle.BaselineGraphDigest, bundle.DeltaDigest, bundle.Decision)
	}
	bundle.Improvement = improvementClaim(bundle.SemanticGraphDelta.ExactPair)
	bundle.CandidateDigest = digestCandidate(bundle)
	if err := ValidateCandidateBundle(bundle, program, input.Release.Graph); err != nil {
		return CandidateBundle{}, err
	}
	return bundle, nil
}

func emptyDelta(graph SemanticGraph) SemanticGraphDelta {
	return SemanticGraphDelta{Before: snapshotFromGraph(graph), After: GraphSnapshot{Digest: "", Concepts: []Concept{}, Predicates: []Predicate{}, Edges: []GraphEdge{}, Cells: []GraphCell{}}, AddedCells: []GraphCell{}, RetiredCells: []GraphCell{}, SplitCells: []SplitCell{}, ExactPair: false}
}

func emptyTestManifest() TestManifest {
	return TestManifest{Schema: "gooo/language-delta-forge/test-manifest/v1", PositiveCases: []ConformanceCase{}, NegativeCases: []ConformanceCase{}, ExactBeforeAfterPair: false, TestExecutionCount: 0, ReuseCount: 0}
}

func bindDirectEvidence(bundle *CandidateBundle, direct DirectCauseEvidence, graph SemanticGraph) {
	delta := SemanticGraphDelta{
		Before:       snapshotFromGraph(graph),
		After:        direct.AfterGraph,
		AddedCells:   cloneCells(direct.AddedCells),
		RetiredCells: cloneCells(direct.RetiredCells),
		SplitCells:   cloneSplits(direct.SplitCells),
		ExactPair:    true,
	}
	bundle.Target = TargetResolution{ResolutionLevel: "FIELD", ConceptID: direct.Target.ConceptID, PredicateID: direct.Target.PredicateID, FieldID: direct.Target.FieldID}
	bundle.CausalFrontier = cloneFrontier(direct.CausalFrontier)
	bundle.SemanticGraphDelta = delta
	bundle.DeltaDigest = mustDigestValue(delta)
	bundle.RollbackDelta = cloneRollback(direct.Rollback)
	bundle.AffectedPredicateIDs = sortedStrings(direct.AffectedPredicateIDs)
	bundle.Counts = DeltaCounts{AddedCells: len(direct.AddedCells), RetiredCells: len(direct.RetiredCells), SplitCells: len(direct.SplitCells)}
	bundle.TestManifest = TestManifest{Schema: "gooo/language-delta-forge/test-manifest/v1", PositiveCases: cloneCases(direct.PositiveCases), NegativeCases: cloneCases(direct.NegativeCases), ExactBeforeAfterPair: true, TestExecutionCount: 0, ReuseCount: 0}
	bundle.IndependentConsumerReceipt = direct.IndependentConsumer
}

func cloneRollback(value RollbackDelta) RollbackDelta {
	return RollbackDelta{RemoveAddedCellIDs: append([]string(nil), value.RemoveAddedCellIDs...), RestoreRetiredCells: cloneCells(value.RestoreRetiredCells), Unsplit: cloneSplits(value.Unsplit), ExactPair: value.ExactPair}
}

func lowerResolutionTarget(graph SemanticGraph, target *Target) TargetResolution {
	if target != nil {
		for _, concept := range graph.Concepts {
			if concept.ID == target.ConceptID {
				return TargetResolution{ResolutionLevel: "CONCEPT", ConceptID: target.ConceptID}
			}
		}
	}
	return TargetResolution{ResolutionLevel: "NONE", ConceptID: "", PredicateID: "", FieldID: ""}
}

func relatedReceipt(input InputBundle, state string) *ObservationReceipt {
	for index := range input.Receipts {
		if input.Receipts[index].State == state {
			return &input.Receipts[index]
		}
	}
	return nil
}

func sourceReceiptReferences(input InputBundle) []ReceiptReference {
	result := []ReceiptReference{{ID: input.Failure.ID, State: input.Failure.State, Digest: digestReceipt(input.Failure), SourceDigest: input.Failure.SourceDigest, GraphDigest: input.Failure.GraphDigest}}
	for _, receipt := range input.Receipts {
		digest := receipt.Digest
		if digest == "" {
			digest = digestReceipt(receipt)
		}
		result = append(result, ReceiptReference{ID: receipt.ID, State: receipt.State, Digest: digest, SourceDigest: receipt.SourceDigest, GraphDigest: receipt.GraphDigest})
	}
	return result
}

func unknownClaim(receipt ObservationReceipt) Claim {
	return Claim{State: StateUnknown, Stage: receipt.Stage, Step: receipt.Step, Reason: receipt.Reason, UnknownClass: receipt.UnknownClass, NextOperation: receipt.NextOperation, BlockedBy: append([]string(nil), receipt.BlockedBy...)}
}

func refutedClaim(receipt ObservationReceipt) Claim {
	reason := receipt.Reason
	if reason == "" {
		reason = "KNOWN_COUNTEREXAMPLE_RECEIPT"
	}
	return Claim{State: StateRefuted, Stage: receipt.Stage, Step: receipt.Step, Reason: reason, UnknownClass: "", NextOperation: "RECORD_REFUTED_CANDIDATE", BlockedBy: []string{receipt.ID}}
}

func unknownIndependentReceipt(bundle CandidateBundle, failureID string, claim Claim) IndependentConsumerReceipt {
	return IndependentConsumerReceipt{Schema: IndependentReceiptSchema, ReceiptID: "independent:" + failureID, ProducerID: "observation-receipt", ConsumerID: "language-delta-forge-independent-consumer", State: StateUnknown, SourceDigest: bundle.SourceDigest, BaselineGraphDigest: bundle.BaselineGraphDigest, DeltaDigest: "", Reason: claim.Reason, BlockedBy: append([]string(nil), claim.BlockedBy...)}
}

func bindIndependentReceipt(receipt IndependentConsumerReceipt, sourceDigest, graphDigest, deltaDigest, decision string) IndependentConsumerReceipt {
	if receipt.Schema == "" {
		receipt.Schema = IndependentReceiptSchema
	}
	if receipt.ProducerID == "" {
		receipt.ProducerID = "observation-receipt"
	}
	if receipt.ConsumerID == "" {
		receipt.ConsumerID = "language-delta-forge-independent-consumer"
	}
	if receipt.ReceiptID == "" {
		receipt.ReceiptID = "independent:candidate"
	}
	receipt.SourceDigest = sourceDigest
	receipt.BaselineGraphDigest = graphDigest
	receipt.DeltaDigest = deltaDigest
	if receipt.State == "" {
		receipt.State = decision
	}
	receipt.ReceiptDigest = digestIndependent(receipt)
	return receipt
}

func improvementClaim(exactPair bool) Claim {
	if exactPair {
		return Claim{State: StateClosed, Stage: "IMPROVEMENT", Step: "bind_exact_before_after_pair", Reason: "EXACT_BEFORE_AFTER_PAIR_BOUND_WITHOUT_SCORE", NextOperation: "HUMAN_REVIEW_EXACT_PAIR", BlockedBy: []string{}}
	}
	return Claim{State: StateUnknown, Stage: "IMPROVEMENT", Step: "require_exact_before_after_pair", Reason: "EXACT_BEFORE_AFTER_PAIR_MISSING", UnknownClass: "MISSING_EXACT_PAIR", NextOperation: "OBTAIN_EXACT_BEFORE_AFTER_PAIR", BlockedBy: []string{"before_graph_digest", "after_graph_digest"}}
}

func validateDirectCause(direct DirectCauseEvidence, graph SemanticGraph) error {
	if !targetExists(graph, direct.Target) {
		return fmt.Errorf("direct target does not exist in baseline graph")
	}
	if direct.BeforeGraphDigest != graph.GraphDigest {
		return fmt.Errorf("direct before graph digest does not match baseline")
	}
	if direct.AfterGraph.Digest == "" || !validDigest(direct.AfterGraph.Digest) || digestSnapshot(direct.AfterGraph) != direct.AfterGraph.Digest {
		return fmt.Errorf("direct after graph digest is not exact")
	}
	if err := validateSnapshot(direct.AfterGraph); err != nil {
		return err
	}
	if len(direct.CausalFrontier) == 0 {
		return fmt.Errorf("causal frontier is empty")
	}
	for _, frontier := range direct.CausalFrontier {
		if !frontierMatches(graph.Edges, frontier) {
			return fmt.Errorf("causal frontier edge %q is not in baseline graph", frontier.EdgeID)
		}
	}
	if len(direct.AffectedPredicateIDs) == 0 {
		return fmt.Errorf("affected predicate IDs are empty")
	}
	seenPredicates := map[string]bool{}
	for _, id := range direct.AffectedPredicateIDs {
		if id == "" || seenPredicates[id] || !predicateExists(graph, id) {
			return fmt.Errorf("invalid affected predicate ID %q", id)
		}
		seenPredicates[id] = true
	}
	if err := validateCellDelta(direct, graph); err != nil {
		return err
	}
	if err := validateRollback(direct); err != nil {
		return err
	}
	if err := validateCases(direct.PositiveCases, "POSITIVE", direct.Target, graph.GraphDigest, direct.AfterGraph.Digest); err != nil {
		return err
	}
	if err := validateCases(direct.NegativeCases, "NEGATIVE", direct.Target, graph.GraphDigest, direct.AfterGraph.Digest); err != nil {
		return err
	}
	if direct.IndependentConsumer.Schema != IndependentReceiptSchema || direct.IndependentConsumer.ProducerID == "" || direct.IndependentConsumer.ConsumerID == "" || direct.IndependentConsumer.ProducerID == direct.IndependentConsumer.ConsumerID {
		return fmt.Errorf("independent consumer receipt is not independent")
	}
	if direct.IndependentConsumer.SourceDigest != "" && !validDigest(direct.IndependentConsumer.SourceDigest) {
		return fmt.Errorf("independent receipt source digest is malformed")
	}
	if direct.IndependentConsumer.BaselineGraphDigest != "" && direct.IndependentConsumer.BaselineGraphDigest != graph.GraphDigest {
		return fmt.Errorf("independent receipt baseline digest mismatch")
	}
	return nil
}

func validateCellDelta(direct DirectCauseEvidence, graph SemanticGraph) error {
	beforeIDs := map[string]bool{}
	for _, cell := range graph.Cells {
		beforeIDs[cell.ID] = true
	}
	after := SemanticGraph{Schema: "gooo/semantic-graph/v1", Version: "v1", Concepts: direct.AfterGraph.Concepts, Predicates: direct.AfterGraph.Predicates, Edges: direct.AfterGraph.Edges, Cells: direct.AfterGraph.Cells, GraphDigest: direct.AfterGraph.Digest}
	if err := validateGraph(after); err != nil {
		return fmt.Errorf("after graph: %w", err)
	}
	addedIDs := map[string]bool{}
	for _, cell := range direct.AddedCells {
		if addedIDs[cell.ID] || beforeIDs[cell.ID] || !containsCell(after.Cells, cell) {
			return fmt.Errorf("added cell %q is not an after-only exact cell", cell.ID)
		}
		addedIDs[cell.ID] = true
	}
	retiredIDs := map[string]bool{}
	for _, cell := range direct.RetiredCells {
		if retiredIDs[cell.ID] || !beforeIDs[cell.ID] || containsCell(after.Cells, cell) {
			return fmt.Errorf("retired cell %q is not a before-only exact cell", cell.ID)
		}
		retiredIDs[cell.ID] = true
	}
	seenSplits := map[string]bool{}
	for _, split := range direct.SplitCells {
		if split.SplitID == "" || seenSplits[split.SplitID] || split.RetiredCellID == "" || !retiredIDs[split.RetiredCellID] || len(split.AddedCellIDs) < 2 {
			return fmt.Errorf("invalid split cell %q", split.SplitID)
		}
		seenSplits[split.SplitID] = true
		for _, addedID := range split.AddedCellIDs {
			if !addedIDs[addedID] {
				return fmt.Errorf("split %q references non-added cell %q", split.SplitID, addedID)
			}
		}
	}
	return nil
}

func validateRollback(direct DirectCauseEvidence) error {
	if !direct.Rollback.ExactPair {
		return fmt.Errorf("rollback is not exact")
	}
	addedIDs := make([]string, 0, len(direct.AddedCells))
	for _, cell := range direct.AddedCells {
		addedIDs = append(addedIDs, cell.ID)
	}
	if !sameSorted(addedIDs, direct.Rollback.RemoveAddedCellIDs) {
		return fmt.Errorf("rollback added-cell inverse mismatch")
	}
	if !sameCells(direct.RetiredCells, direct.Rollback.RestoreRetiredCells) {
		return fmt.Errorf("rollback retired-cell inverse mismatch")
	}
	if !sameSplits(direct.SplitCells, direct.Rollback.Unsplit) {
		return fmt.Errorf("rollback split inverse mismatch")
	}
	return nil
}

func validateCases(cases []ConformanceCase, polarity string, target Target, beforeDigest, afterDigest string) error {
	if len(cases) == 0 {
		return fmt.Errorf("%s conformance cases are empty", polarity)
	}
	seen := map[string]bool{}
	for _, testCase := range cases {
		if testCase.ID == "" || seen[testCase.ID] || testCase.Polarity != polarity || testCase.Target != target || testCase.BeforeGraphDigest != beforeDigest || testCase.AfterGraphDigest != afterDigest || !validDigest(testCase.FixtureDigest) || testCase.ExpectedPredicateState == "" {
			return fmt.Errorf("invalid %s conformance case %q", polarity, testCase.ID)
		}
		seen[testCase.ID] = true
	}
	return nil
}

func validateSnapshot(snapshot GraphSnapshot) error {
	graph := SemanticGraph{Schema: "gooo/semantic-graph/v1", Version: "v1", Concepts: snapshot.Concepts, Predicates: snapshot.Predicates, Edges: snapshot.Edges, Cells: snapshot.Cells, GraphDigest: snapshot.Digest}
	return validateGraph(graph)
}

func frontierMatches(edges []GraphEdge, target FrontierEdge) bool {
	for _, edge := range edges {
		if edge.ID == target.EdgeID && edge.From == target.From && edge.To == target.To && edge.Kind == target.Kind && edge.PredicateID == target.PredicateID {
			return true
		}
	}
	return false
}

func predicateExists(graph SemanticGraph, id string) bool {
	_, ok := predicateByID(graph, id)
	return ok
}

func containsCell(cells []GraphCell, target GraphCell) bool {
	for _, cell := range cells {
		if cell == target {
			return true
		}
	}
	return false
}

func sameSorted(left, right []string) bool {
	return equalStrings(sortedStrings(left), sortedStrings(right))
}

func sameCells(left, right []GraphCell) bool {
	if len(left) != len(right) {
		return false
	}
	leftDigest := make([]string, len(left))
	rightDigest := make([]string, len(right))
	for index, value := range left {
		leftDigest[index] = mustDigestValue(value)
	}
	for index, value := range right {
		rightDigest[index] = mustDigestValue(value)
	}
	return sameSorted(leftDigest, rightDigest)
}

func sameSplits(left, right []SplitCell) bool {
	if len(left) != len(right) {
		return false
	}
	leftDigest := make([]string, len(left))
	rightDigest := make([]string, len(right))
	for index, value := range left {
		leftDigest[index] = mustDigestValue(value)
	}
	for index, value := range right {
		rightDigest[index] = mustDigestValue(value)
	}
	return sameSorted(leftDigest, rightDigest)
}

func ValidateCandidateBundle(bundle CandidateBundle, program Program, graph SemanticGraph) error {
	if bundle.Schema != CandidateBundleSchema || bundle.Version != "v1" || (bundle.Decision != StateClosed && bundle.Decision != StateUnknown && bundle.Decision != StateRefuted) {
		return fmt.Errorf("candidate bundle schema or decision is invalid")
	}
	if !validDigest(bundle.SourceDigest) || !validDigest(bundle.BaselineGraphDigest) || bundle.BaselineGraphDigest != graph.GraphDigest {
		return fmt.Errorf("candidate source or baseline digest is invalid")
	}
	if bundle.Authority.RepositoryWrites != 0 || bundle.Authority.LocalTestExecutions != 0 || bundle.Authority.ProtectedCoreAdoption != 0 || bundle.Authority.MergeAuthority != 0 || !bundle.Authority.ReadOnlyInput {
		return fmt.Errorf("candidate authority is not read-only")
	}
	if !bundle.Adoption.CandidateOnly || bundle.Adoption.ProtectedCoreMutation || bundle.Adoption.AutomaticMerge || !bundle.Adoption.SeparateAuthorityStep {
		return fmt.Errorf("candidate adoption boundary is invalid")
	}
	if bundle.CandidateDigest == "" || bundle.CandidateDigest != digestCandidate(bundle) {
		return fmt.Errorf("candidate digest mismatch")
	}
	if bundle.SemanticGraphDelta.ExactPair {
		if err := validateBoundDelta(bundle, graph); err != nil {
			return err
		}
	} else if bundle.Decision == StateClosed {
		return fmt.Errorf("closed candidate has no exact graph pair")
	} else {
		if err := validateUnboundIndependentReceipt(bundle); err != nil {
			return err
		}
	}
	if bundle.Decision == StateUnknown && !bundle.Claim.HasUnknownTuple() {
		return fmt.Errorf("UNKNOWN candidate lacks six-field claim")
	}
	if bundle.Decision == StateRefuted && bundle.Claim.State != StateRefuted {
		return fmt.Errorf("REFUTED candidate lacks refuted claim")
	}
	if bundle.Decision == StateClosed && bundle.Claim.State != StateClosed {
		return fmt.Errorf("CLOSED candidate claim mismatch")
	}
	if bundle.Decision == StateUnknown && (bundle.Target.ResolutionLevel == "FIELD" || bundle.Target.PredicateID != "" || bundle.Target.FieldID != "") {
		return fmt.Errorf("UNKNOWN candidate was not lowered to concept resolution")
	}
	if bundle.Improvement.State == StateUnknown && !bundle.Improvement.HasUnknownTuple() {
		return fmt.Errorf("UNKNOWN improvement lacks six-field claim")
	}
	if bundle.Improvement.State == StateUnknown && bundle.SemanticGraphDelta.ExactPair {
		return fmt.Errorf("exact pair cannot produce UNKNOWN improvement")
	}
	if bundle.TestManifest.TestExecutionCount != 0 || bundle.TestManifest.ReuseCount != 0 {
		return fmt.Errorf("candidate test manifest reports execution or reuse")
	}
	if len(program.Cells) != 18 {
		return fmt.Errorf("program denominator is not fixed")
	}
	return nil
}

func validateUnboundIndependentReceipt(bundle CandidateBundle) error {
	receipt := bundle.IndependentConsumerReceipt
	if receipt.Schema != IndependentReceiptSchema || receipt.ProducerID == "" || receipt.ConsumerID == "" || receipt.ProducerID == receipt.ConsumerID || receipt.SourceDigest != bundle.SourceDigest || receipt.BaselineGraphDigest != bundle.BaselineGraphDigest || receipt.DeltaDigest != "" || receipt.State != StateUnknown {
		return fmt.Errorf("unbound independent consumer receipt is invalid")
	}
	if !validDigest(receipt.SourceDigest) || !validDigest(receipt.BaselineGraphDigest) || len(receipt.BlockedBy) == 0 || receipt.Reason == "" {
		return fmt.Errorf("unbound independent consumer receipt lacks blocker evidence")
	}
	return nil
}

func validateBoundDelta(bundle CandidateBundle, graph SemanticGraph) error {
	delta := bundle.SemanticGraphDelta
	if delta.Before.Digest != graph.GraphDigest || digestSnapshot(delta.Before) != graph.GraphDigest || !validDigest(delta.After.Digest) || digestSnapshot(delta.After) != delta.After.Digest {
		return fmt.Errorf("bound graph delta is not exact")
	}
	if bundle.DeltaDigest != mustDigestValue(delta) {
		return fmt.Errorf("delta digest mismatch")
	}
	if bundle.Counts.AddedCells != len(delta.AddedCells) || bundle.Counts.RetiredCells != len(delta.RetiredCells) || bundle.Counts.SplitCells != len(delta.SplitCells) {
		return fmt.Errorf("delta counts are not exact")
	}
	if err := validateBoundCellDelta(delta, graph); err != nil {
		return err
	}
	if bundle.Target.ResolutionLevel != "FIELD" || !targetExists(graph, Target{ConceptID: bundle.Target.ConceptID, PredicateID: bundle.Target.PredicateID, FieldID: bundle.Target.FieldID}) {
		return fmt.Errorf("exact target is not bound to baseline graph")
	}
	for _, frontier := range bundle.CausalFrontier {
		if !frontierMatches(graph.Edges, frontier) {
			return fmt.Errorf("candidate causal frontier edge %q is not in baseline graph", frontier.EdgeID)
		}
	}
	if len(bundle.CausalFrontier) == 0 {
		return fmt.Errorf("candidate causal frontier is empty")
	}
	if err := validateCases(deltaCases(bundle), "POSITIVE", Target{ConceptID: bundle.Target.ConceptID, PredicateID: bundle.Target.PredicateID, FieldID: bundle.Target.FieldID}, graph.GraphDigest, delta.After.Digest); err != nil {
		return err
	}
	if err := validateCases(bundle.TestManifest.NegativeCases, "NEGATIVE", Target{ConceptID: bundle.Target.ConceptID, PredicateID: bundle.Target.PredicateID, FieldID: bundle.Target.FieldID}, graph.GraphDigest, delta.After.Digest); err != nil {
		return err
	}
	if !bundle.TestManifest.ExactBeforeAfterPair || bundle.TestManifest.Schema != "gooo/language-delta-forge/test-manifest/v1" {
		return fmt.Errorf("test manifest is not exact")
	}
	addedIDs := make([]string, 0, len(delta.AddedCells))
	for _, cell := range delta.AddedCells {
		addedIDs = append(addedIDs, cell.ID)
	}
	if !sameSorted(addedIDs, bundle.RollbackDelta.RemoveAddedCellIDs) || !sameCells(delta.RetiredCells, bundle.RollbackDelta.RestoreRetiredCells) || !sameSplits(delta.SplitCells, bundle.RollbackDelta.Unsplit) || !bundle.RollbackDelta.ExactPair {
		return fmt.Errorf("candidate rollback delta is not the exact inverse")
	}
	if bundle.IndependentConsumerReceipt.SourceDigest != bundle.SourceDigest || bundle.IndependentConsumerReceipt.BaselineGraphDigest != bundle.BaselineGraphDigest || bundle.IndependentConsumerReceipt.DeltaDigest != bundle.DeltaDigest || bundle.IndependentConsumerReceipt.ReceiptDigest != digestIndependent(bundle.IndependentConsumerReceipt) {
		return fmt.Errorf("independent consumer receipt is not bound")
	}
	if bundle.IndependentConsumerReceipt.Schema != IndependentReceiptSchema || bundle.IndependentConsumerReceipt.ProducerID == bundle.IndependentConsumerReceipt.ConsumerID || bundle.IndependentConsumerReceipt.ProducerID == "" || bundle.IndependentConsumerReceipt.ConsumerID == "" {
		return fmt.Errorf("independent consumer receipt is not independent")
	}
	return nil
}

func validateBoundCellDelta(delta SemanticGraphDelta, graph SemanticGraph) error {
	beforeIDs := map[string]bool{}
	for _, cell := range graph.Cells {
		beforeIDs[cell.ID] = true
	}
	if err := validateSnapshot(delta.After); err != nil {
		return fmt.Errorf("after graph: %w", err)
	}
	addedIDs := map[string]bool{}
	for _, cell := range delta.AddedCells {
		if addedIDs[cell.ID] || beforeIDs[cell.ID] || !containsCell(delta.After.Cells, cell) {
			return fmt.Errorf("invalid added cell %q", cell.ID)
		}
		addedIDs[cell.ID] = true
	}
	retiredIDs := map[string]bool{}
	for _, cell := range delta.RetiredCells {
		if retiredIDs[cell.ID] || !beforeIDs[cell.ID] || containsCell(delta.After.Cells, cell) {
			return fmt.Errorf("invalid retired cell %q", cell.ID)
		}
		retiredIDs[cell.ID] = true
	}
	for _, split := range delta.SplitCells {
		if split.SplitID == "" || !retiredIDs[split.RetiredCellID] || len(split.AddedCellIDs) < 2 {
			return fmt.Errorf("invalid split cell %q", split.SplitID)
		}
		for _, addedID := range split.AddedCellIDs {
			if !addedIDs[addedID] {
				return fmt.Errorf("split %q references non-added cell", split.SplitID)
			}
		}
	}
	return nil
}

func deltaCases(bundle CandidateBundle) []ConformanceCase { return bundle.TestManifest.PositiveCases }

func sortCases(cases []ConformanceCase) []ConformanceCase {
	result := cloneCases(cases)
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
