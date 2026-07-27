package ecl

import "testing"

func TestRunSubsetDelayContinuesWithoutChangingMemory(t *testing.T) {
	block := []byte{0, 0, 0x3A, 0x3A, 0x00}
	result, err := RunSubset(block, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if result.DelayCount != 2 || result.Steps != 3 || result.PC != 3 {
		t.Fatalf("result=%+v", result)
	}
}
