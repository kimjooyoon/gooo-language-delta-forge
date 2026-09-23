package forge

import "testing"

func TestValidateGraphRejectsDanglingEdgeEndpoint(t *testing.T) {
	graph := SemanticGraph{
		Schema:  "gooo/semantic-graph/v1",
		Version: "v1",
		Concepts: []Concept{{ID: "concept.order", Name: "Order"}},
		Predicates: []Predicate{{
			ID:        "predicate.order_status",
			ConceptID: "concept.order",
			Name:      "has_status",
			Fields:    []Field{{ID: "field.order_status", Name: "status", Type: "string"}},
		}},
		Edges: []GraphEdge{{
			ID:          "edge.dangling",
			From:        "concept.order",
			To:          "predicate.missing",
			Kind:        "binds",
			PredicateID: "predicate.order_status",
		}},
	}

	if err := validateGraph(graph); err == nil {
		t.Fatal("validateGraph accepted an edge with a dangling endpoint")
	}
}
