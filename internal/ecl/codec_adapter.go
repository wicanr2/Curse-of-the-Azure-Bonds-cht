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

const CodeAddressBase = engineecl.CodeAddressBase

var KnownCommands = engineecl.KnownCommands
var VariableLengthCommands = engineecl.VariableLengthCommands

var ParseOperands = engineecl.ParseOperands
var EntryPoints = engineecl.EntryPoints
var Trace = engineecl.Trace
var TraceAt = engineecl.TraceAt
var ScanKnownInstructions = engineecl.ScanKnownInstructions
var FindSaveDestinationCandidates = engineecl.FindSaveDestinationCandidates
var RecordEnd = engineecl.RecordEnd
var BranchTargets = engineecl.BranchTargets
var MenuEnd = engineecl.MenuEnd
var CodeTarget = engineecl.CodeTarget
var TraceGraph = engineecl.TraceGraph

// decodeInstruction preserves the title adapter's internal seam while the
// generic decoder remains owned and tested by the engine module.
func decodeInstruction(payload []byte, offset int) (Instruction, error) {
	return engineecl.DecodeInstruction(payload, offset)
}
