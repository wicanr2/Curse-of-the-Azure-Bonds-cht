package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/dax"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

func main() {
	image := flag.String("image", "curseoftheazurebonds.zip", "original DOS image ZIP")
	member := flag.String("member", "ECL1.DAX", "DAX member to inspect")
	trace := flag.Bool("trace", false, "trace known ECL cursor commands")
	traceStart := flag.Int("trace-start", -1, "decoded payload offset for -trace (default 0)")
	stringsOnly := flag.Bool("strings", false, "print ECL packed-text candidates")
	graph := flag.Bool("graph", false, "trace statically reachable ECL branches")
	entryPoints := flag.Bool("entrypoints", false, "print five ECL initialization entry points")
	runSubset := flag.Bool("run-subset", false, "run the bounded ECL command subset from the initial entry")
	flag.Parse()
	data, err := zipMember(*image, *member)
	if err != nil {
		log.Fatal(err)
	}
	blocks, err := dax.Parse(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: %d blocks\n", *member, len(blocks))
	for _, block := range blocks {
		fmt.Printf("block %d: %d decoded bytes\n", block.Entry.ID, len(block.Data))
		if *trace {
			var instructions []ecl.Instruction
			var err error
			if *traceStart >= 0 {
				instructions, err = ecl.TraceAt(block.Data, *traceStart, 40)
			} else {
				instructions, err = ecl.Trace(block.Data, 40)
			}
			for _, instruction := range instructions {
				fmt.Printf("  +0x%04X %-16s operands=%d\n", instruction.Offset, instruction.Command.Name, len(instruction.Operands))
				for _, operand := range instruction.Operands {
					if len(operand.Packed) > 0 {
						fmt.Printf("    packed=%q\n", ecl.DecodePackedText(operand.Packed))
					}
				}
			}
			if err != nil {
				fmt.Printf("  trace stopped safely: %v\n", err)
			}
		}
		if *stringsOnly {
			for _, text := range ecl.FindPackedTextCandidates(block.Data) {
				fmt.Printf("  text=%q\n", text)
			}
		}
		if *graph {
			starts := []int(nil)
			if points, _, entryErr := ecl.EntryPoints(block.Data, 5); entryErr == nil && len(points) == 5 {
				starts = []int{int(points[4]) - ecl.CodeAddressBase}
				fmt.Printf("  graph start=+0x%04X (initial entry 0x%04X)\n", starts[0], points[4])
			} else if entryErr != nil {
				fmt.Printf("  entry point unavailable; graph fallback start=+0x0000: %v\n", entryErr)
			}
			result, err := ecl.TraceGraph(block.Data, starts, 2000)
			fmt.Printf("  graph instructions=%d edges=%d\n", len(result.Instructions), len(result.Edges))
			for _, edge := range result.Edges {
				fmt.Printf("  edge +0x%04X -> +0x%04X (%s)\n", edge.From, edge.To, edge.Kind)
			}
			if err != nil {
				fmt.Printf("  graph stopped safely: %v\n", err)
			}
		}
		if *entryPoints {
			points, next, err := ecl.EntryPoints(block.Data, 5)
			if err != nil {
				fmt.Printf("  entry points stopped safely: %v\n", err)
				continue
			}
			fmt.Printf("  entry points cursor=+0x%04X", next)
			for _, point := range points {
				fmt.Printf(" 0x%04X", point)
			}
			fmt.Println()
		}
		if *runSubset {
			start := 0
			if points, _, entryErr := ecl.EntryPoints(block.Data, 5); entryErr == nil && len(points) == 5 {
				start = int(points[4]) - ecl.CodeAddressBase
			}
			result, runErr := ecl.RunSubset(block.Data, start, 500)
			fmt.Printf("  subset steps=%d stop=+0x%04X texts=%d\n", result.Steps, result.PC, len(result.Text))
			for _, message := range result.Text {
				fmt.Printf("    text=%q\n", message)
			}
			if runErr != nil {
				fmt.Printf("  subset stopped safely: %v\n", runErr)
			}
		}
	}
}

func zipMember(path, member string) ([]byte, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	for _, file := range archive.File {
		if file.Name != member {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return io.ReadAll(reader)
	}
	return nil, fmt.Errorf("member %q not found", member)
}
