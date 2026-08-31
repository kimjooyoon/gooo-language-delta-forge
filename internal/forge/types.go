package forge

const (
	InputSchema               = "gooo/language-delta-forge/input/v1"
	ReleaseSchema             = "gooo/language-delta-forge/immutable-release/v1"
	ReceiptSchema             = "gooo/language-delta-forge/observation-receipt/v1"
	ProgramSchema             = "gooo/language-delta-forge/program/v1"
	CandidateBundleSchema     = "gooo/language-delta-forge/candidate-bundle/v1"
	IndependentReceiptSchema  = "gooo/language-delta-forge/independent-consumer-receipt/v1"
	ConformanceManifestSchema = "gooo/language-delta-forge/conformance-manifest/v1"
	ConformanceReportSchema   = "gooo/language-delta-forge/conformance-report/v1"
	DenominatorSchema         = "gooo/language-delta-forge/denominator/v1"

	StateClosed  = "CLOSED"
	StateUnknown = "UNKNOWN"
	StateRefuted = "REFUTED"
	StateFailure = "FAILURE"

	ProofFoundation    = "FOUNDATION"
	ProofCoherence     = "COHERENCE"
	ProofRegression    = "REGRESSION"
	IndicatorDriver    = "DRIVER"
	IndicatorOutcome   = "OUTCOME"
	IndicatorGuardrail = "GUARDRAIL"
)

type InputBundle struct {
	Schema      string               `json:"schema"`
	ReleasePath string               `json:"release_path,omitempty"`
	Release     ImmutableRelease     `json:"release"`
	Failure     ObservationReceipt   `json:"failure"`
	Receipts    []ObservationReceipt `json:"receipts"`
}

type ImmutableRelease struct {
	Schema   string          `json:"schema"`
	Version  string          `json:"version"`
	Identity ReleaseIdentity `json:"identity"`
	Graph    SemanticGraph   `json:"graph"`
}

type ReleaseIdentity struct {
	Repository   string `json:"repository"`
	Tag          string `json:"tag"`
	Commit       string `json:"commit"`
	SourceDigest string `json:"source_digest"`
	GraphDigest  string `json:"graph_digest"`
	Digest       string `json:"digest"`
}

type SemanticGraph struct {
	Schema      string      `json:"schema"`
	Version     string      `json:"version"`
	Concepts    []Concept   `json:"concepts"`
	Predicates  []Predicate `json:"predicates"`
	Edges       []GraphEdge `json:"edges"`
	Cells       []GraphCell `json:"cells"`
	GraphDigest string      `json:"graph_digest"`
}

type GraphSnapshot struct {
	Digest     string      `json:"digest"`
	Concepts   []Concept   `json:"concepts"`
	Predicates []Predicate `json:"predicates"`
	Edges      []GraphEdge `json:"edges"`
	Cells      []GraphCell `json:"cells"`
}

type Concept struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Predicate struct {
	ID        string  `json:"id"`
	ConceptID string  `json:"concept_id"`
	Name      string  `json:"name"`
	Fields    []Field `json:"fields"`
}

type Field struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type GraphEdge struct {
	ID          string `json:"id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Kind        string `json:"kind"`
	PredicateID string `json:"predicate_id"`
}

type GraphCell struct {
	ID          string `json:"id"`
	ConceptID   string `json:"concept_id"`
	PredicateID string `json:"predicate_id"`
	FieldID     string `json:"field_id"`
	Relation    string `json:"relation"`
	Constraint  string `json:"constraint"`
}

type ObservationReceipt struct {
	Schema         string               `json:"schema"`
	ID             string               `json:"id"`
	State          string               `json:"state"`
	RelatedTo      string               `json:"related_to,omitempty"`
	SourceDigest   string               `json:"source_digest"`
	GraphDigest    string               `json:"graph_digest"`
	Stage          string               `json:"stage,omitempty"`
	Step           string               `json:"step,omitempty"`
	Reason         string               `json:"reason,omitempty"`
	UnknownClass   string               `json:"unknown_class,omitempty"`
	NextOperation  string               `json:"next_operation,omitempty"`
	BlockedBy      []string             `json:"blocked_by,omitempty"`
	Target         *Target              `json:"target,omitempty"`
	DirectCause    *DirectCauseEvidence `json:"direct_cause,omitempty"`
	Counterexample string               `json:"counterexample,omitempty"`
	Digest         string               `json:"digest,omitempty"`
}

type Target struct {
	ConceptID   string `json:"concept_id"`
	PredicateID string `json:"predicate_id"`
	FieldID     string `json:"field_id"`
}

type DirectCauseEvidence struct {
	Target               Target                     `json:"target"`
	CausalFrontier       []FrontierEdge             `json:"causal_frontier"`
	BeforeGraphDigest    string                     `json:"before_graph_digest"`
	AfterGraph           GraphSnapshot              `json:"after_graph"`
	AffectedPredicateIDs []string                   `json:"affected_predicate_ids"`
	AddedCells           []GraphCell                `json:"added_cells"`
	RetiredCells         []GraphCell                `json:"retired_cells"`
	SplitCells           []SplitCell                `json:"split_cells"`
	Rollback             RollbackDelta              `json:"rollback"`
	PositiveCases        []ConformanceCase          `json:"positive_cases"`
	NegativeCases        []ConformanceCase          `json:"negative_cases"`
	IndependentConsumer  IndependentConsumerReceipt `json:"independent_consumer"`
}

type FrontierEdge struct {
	EdgeID      string `json:"edge_id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Kind        string `json:"kind"`
	PredicateID string `json:"predicate_id"`
}

type SplitCell struct {
	SplitID       string   `json:"split_id"`
	RetiredCellID string   `json:"retired_cell_id"`
	AddedCellIDs  []string `json:"added_cell_ids"`
}

type RollbackDelta struct {
	RemoveAddedCellIDs  []string    `json:"remove_added_cell_ids"`
	RestoreRetiredCells []GraphCell `json:"restore_retired_cells"`
	Unsplit             []SplitCell `json:"unsplit"`
	ExactPair           bool        `json:"exact_pair"`
}

type ConformanceCase struct {
	ID                     string `json:"id"`
	Polarity               string `json:"polarity"`
	BeforeGraphDigest      string `json:"before_graph_digest"`
	AfterGraphDigest       string `json:"after_graph_digest"`
	Target                 Target `json:"target"`
	ExpectedPredicateState string `json:"expected_predicate_state"`
	FixtureDigest          string `json:"fixture_digest"`
}

type IndependentConsumerReceipt struct {
	Schema              string   `json:"schema"`
	ReceiptID           string   `json:"receipt_id"`
	ProducerID          string   `json:"producer_id"`
	ConsumerID          string   `json:"consumer_id"`
	State               string   `json:"state"`
	SourceDigest        string   `json:"source_digest"`
	BaselineGraphDigest string   `json:"baseline_graph_digest"`
	DeltaDigest         string   `json:"delta_digest"`
	ReceiptDigest       string   `json:"receipt_digest,omitempty"`
	Reason              string   `json:"reason,omitempty"`
	BlockedBy           []string `json:"blocked_by,omitempty"`
}

type Program struct {
	Schema        string        `json:"schema"`
	Version       string        `json:"version"`
	Denominator   Denominator   `json:"denominator"`
	Authority     Authority     `json:"authority"`
	Precedence    []string      `json:"precedence"`
	UnknownFields []string      `json:"unknown_fields"`
	Cells         []ProgramCell `json:"cells"`
	SourceDigest  string        `json:"source_digest"`
}

type Authority struct {
	RepositoryWrites      int `json:"repository_writes"`
	LocalTestExecutions   int `json:"local_test_executions"`
	ProtectedCoreAdoption int `json:"protected_core_adoption"`
	MergeAuthority        int `json:"merge_authority"`
}

type Denominator struct {
	Schema                 string   `json:"schema"`
	ID                     string   `json:"id"`
	CellCount              int      `json:"cell_count"`
	Fixed                  bool     `json:"fixed"`
	ProofChoices           []string `json:"proof_choices"`
	IndicatorClasses       []string `json:"indicator_classes"`
	CellsPerProofChoice    int      `json:"cells_per_proof_choice"`
	CellsPerIndicatorClass int      `json:"cells_per_indicator_class"`
	UnknownFields          []string `json:"unknown_fields"`
	Precedence             []string `json:"precedence"`
}

type ProgramCell struct {
	Ordinal        int    `json:"ordinal"`
	ID             string `json:"id"`
	ProofChoice    string `json:"proof_choice"`
	IndicatorClass string `json:"indicator_class"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
}

type CandidateBundle struct {
	Schema                     string                     `json:"schema"`
	Version                    string                     `json:"version"`
	Decision                   string                     `json:"decision"`
	SourceDigest               string                     `json:"source_digest"`
	BaselineGraphDigest        string                     `json:"baseline_graph_digest"`
	DeltaDigest                string                     `json:"delta_digest"`
	Target                     TargetResolution           `json:"target"`
	CausalFrontier             []FrontierEdge             `json:"causal_frontier"`
	SemanticGraphDelta         SemanticGraphDelta         `json:"semantic_graph_delta"`
	RollbackDelta              RollbackDelta              `json:"rollback_delta"`
	AffectedPredicateIDs       []string                   `json:"affected_predicate_ids"`
	Counts                     DeltaCounts                `json:"counts"`
	TestManifest               TestManifest               `json:"test_manifest"`
	IndependentConsumerReceipt IndependentConsumerReceipt `json:"independent_consumer_receipt"`
	SourceReceipts             []ReceiptReference         `json:"source_receipts"`
	Claim                      Claim                      `json:"claim"`
	Improvement                Claim                      `json:"improvement"`
	Authority                  AuthorityReceipt           `json:"authority"`
	Adoption                   AdoptionBoundary           `json:"adoption"`
	CandidateDigest            string                     `json:"candidate_digest"`
}

type TargetResolution struct {
	ResolutionLevel string `json:"resolution_level"`
	ConceptID       string `json:"concept_id"`
	PredicateID     string `json:"predicate_id"`
	FieldID         string `json:"field_id"`
}

type SemanticGraphDelta struct {
	Before       GraphSnapshot `json:"before"`
	After        GraphSnapshot `json:"after"`
	AddedCells   []GraphCell   `json:"added_cells"`
	RetiredCells []GraphCell   `json:"retired_cells"`
	SplitCells   []SplitCell   `json:"split_cells"`
	ExactPair    bool          `json:"exact_pair"`
}

type DeltaCounts struct {
	AddedCells   int `json:"added_cells"`
	RetiredCells int `json:"retired_cells"`
	SplitCells   int `json:"split_cells"`
}

type TestManifest struct {
	Schema               string            `json:"schema"`
	PositiveCases        []ConformanceCase `json:"positive_cases"`
	NegativeCases        []ConformanceCase `json:"negative_cases"`
	ExactBeforeAfterPair bool              `json:"exact_before_after_pair"`
	TestExecutionCount   int               `json:"test_execution_count"`
	ReuseCount           int               `json:"reuse_count"`
}

type ReceiptReference struct {
	ID           string `json:"id"`
	State        string `json:"state"`
	Digest       string `json:"digest"`
	SourceDigest string `json:"source_digest"`
	GraphDigest  string `json:"graph_digest"`
}

type Claim struct {
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type AuthorityReceipt struct {
	RepositoryWrites      int  `json:"repository_writes"`
	LocalTestExecutions   int  `json:"local_test_executions"`
	ProtectedCoreAdoption int  `json:"protected_core_adoption"`
	MergeAuthority        int  `json:"merge_authority"`
	ReadOnlyInput         bool `json:"read_only_input"`
}

type AdoptionBoundary struct {
	CandidateOnly         bool `json:"candidate_only"`
	ProtectedCoreMutation bool `json:"protected_core_mutation"`
	AutomaticMerge        bool `json:"automatic_merge"`
	SeparateAuthorityStep bool `json:"separate_authority_step"`
}

type ConformanceManifest struct {
	Schema string                `json:"schema"`
	Cases  []ConformanceCaseSpec `json:"cases"`
}

type ConformanceCaseSpec struct {
	ID               string `json:"id"`
	InputPath        string `json:"input_path"`
	ExpectedDecision string `json:"expected_decision"`
	ProofChoice      string `json:"proof_choice"`
	IndicatorClass   string `json:"indicator_class"`
}

type ConformanceReport struct {
	Schema      string                  `json:"schema"`
	Decision    string                  `json:"decision"`
	Denominator int                     `json:"denominator"`
	Cases       []ConformanceCaseResult `json:"cases"`
	Summary     ConformanceSummary      `json:"summary"`
	Metrics     ExecutionMetrics        `json:"metrics"`
}

type ConformanceCaseResult struct {
	ID               string `json:"id"`
	ExpectedDecision string `json:"expected_decision"`
	ObservedDecision string `json:"observed_decision"`
	ProofChoice      string `json:"proof_choice"`
	IndicatorClass   string `json:"indicator_class"`
	CandidateDigest  string `json:"candidate_digest"`
	Pass             bool   `json:"pass"`
}

type ConformanceSummary struct {
	Total    int `json:"total"`
	Closed   int `json:"closed"`
	Unknown  int `json:"unknown"`
	Refuted  int `json:"refuted"`
	Failures int `json:"failures"`
}

type ExecutionMetrics struct {
	Files               int   `json:"files"`
	Directories         int   `json:"directories"`
	PhysicalLines       int   `json:"physical_lines"`
	OutputFiles         int   `json:"output_files"`
	PeakRSSKiB          int64 `json:"peak_rss_kib"`
	CompileWallMS       int64 `json:"compile_wall_ms"`
	BuildWallMS         int64 `json:"build_wall_ms"`
	TestWallMS          int64 `json:"test_wall_ms"`
	ConformanceWallMS   int64 `json:"conformance_wall_ms"`
	TestExecutionCount  int   `json:"test_execution_count"`
	ReuseCount          int   `json:"reuse_count"`
	LocalTestExecutions int   `json:"local_test_executions"`
}

type InventoryReport struct {
	Schema             string `json:"schema"`
	Files              int    `json:"files"`
	Directories        int    `json:"directories"`
	GoFiles            int    `json:"go_files"`
	GoooFiles          int    `json:"gooo_files"`
	PhysicalLines      int    `json:"physical_lines"`
	GoLines            int    `json:"go_lines"`
	GoooLines          int    `json:"gooo_lines"`
	RootReadmeExcluded bool   `json:"root_readme_excluded"`
}

func (c Claim) HasUnknownTuple() bool {
	return c.State == StateUnknown && c.Stage != "" && c.Step != "" && c.Reason != "" &&
		c.UnknownClass != "" && c.NextOperation != "" && len(c.BlockedBy) > 0
}
