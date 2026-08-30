// Package ecl is the Curse of the Azure Bonds title adapter for the reusable
// Gold Box ECL decoder. It owns title memory, text, event and UI effects; byte
// framing and control-flow decoding live in golden-box-remake-engine/ecl.
package ecl

import engineecl "github.com/wicanr2/golden-box-remake-engine/ecl"

type Operand = engineecl.Operand
type Command = engineecl.Command
type Instruction = engineecl.Instruction
type Edge = engineecl.Edge
type Graph = engineecl.Graph

// CodeAddressBase is CoAB title data, not a Gold Box engine constant.
const CodeAddressBase = 0x8000

var KnownCommands = engineecl.KnownCommands
var VariableLengthCommands = engineecl.VariableLengthCommands

var ParseOperands = engineecl.ParseOperands
var EntryPoints = engineecl.EntryPoints
var Trace = engineecl.Trace
var TraceAt = engineecl.TraceAt
var ScanKnownInstructions = engineecl.ScanKnownInstructions
var FindSaveDestinationCandidates = engineecl.FindSaveDestinationCandidates
var RecordEnd = engineecl.RecordEnd
var MenuEnd = engineecl.MenuEnd

func BranchTargets(block []byte, offset int) ([]int, int, error) {
	return engineecl.BranchTargetsAtBase(block, offset, CodeAddressBase)
}

func CodeTarget(operand Operand, payloadLength int) (int, bool) {
	return engineecl.CodeTargetAtBase(operand, CodeAddressBase, payloadLength)
}

func TraceGraph(block []byte, starts []int, limit int) (Graph, error) {
	return engineecl.TraceGraphAtBase(block, starts, CodeAddressBase, limit)
}

// decodeInstruction preserves the title adapter's internal seam while the
// generic decoder remains owned and tested by the engine module.
func decodeInstruction(payload []byte, offset int) (Instruction, error) {
	return engineecl.DecodeInstruction(payload, offset)
}
