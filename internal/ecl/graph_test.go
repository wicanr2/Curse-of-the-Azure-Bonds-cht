package ecl

import "testing"

func TestTraceGraphFollowsGotoTarget(t *testing.T) {
	// Prefix, GOTO 0x8006, padding, PRINT 3.
	block := []byte{0, 0, 0x01, 0x01, 0x06, 0x80, 0, 0, 0x11, 0x00, 0x03}
	graph, err := TraceGraph(block, nil, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Edges) != 1 || graph.Edges[0].To != 6 {
		t.Fatalf("edges=%#v", graph.Edges)
	}
	if len(graph.Instructions) != 2 || graph.Instructions[1].Command.Name != "PRINT" {
		t.Fatalf("instructions=%#v", graph.Instructions)
	}
}

func TestCodeTargetRejectsDataPointer(t *testing.T) {
	operand := Operand{Word: 0x6B01, WordSet: true}
	if _, ok := CodeTarget(operand, 100); ok {
		t.Fatal("data pointer was treated as code target")
	}
}
