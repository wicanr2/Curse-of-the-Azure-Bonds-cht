// Package ecl contains small, evidence-backed pieces of the ECL reader.
// It does not yet execute ECL commands.
package ecl

import "fmt"

type Operand struct {
	Code    byte
	Low     byte
	High    byte
	Word    uint16
	WordSet bool
}

// ParseOperands follows the operand framing used by the Gold Box ECL dump:
// code and low are read at offset+1/offset+2; codes 1, 2 and 3 consume one
// additional high byte. The skipped byte at offset and the final increment
// are part of the original VM's instruction cursor convention.
func ParseOperands(payload []byte, offset, count int) ([]Operand, int, error) {
	if offset < 0 || count < 0 || offset > len(payload) {
		return nil, offset, fmt.Errorf("invalid operand range")
	}
	operands := make([]Operand, 0, count)
	pos := offset
	for i := 0; i < count; i++ {
		if pos+2 >= len(payload) {
			return nil, pos, fmt.Errorf("operand %d is truncated", i)
		}
		operand := Operand{Code: payload[pos+1], Low: payload[pos+2]}
		pos += 2
		if operand.Code == 1 || operand.Code == 2 || operand.Code == 3 {
			pos++
			if pos >= len(payload) {
				return nil, pos, fmt.Errorf("operand %d high byte is truncated", i)
			}
			operand.High = payload[pos]
			operand.Word = uint16(operand.High)<<8 | uint16(operand.Low)
			operand.WordSet = true
		}
		operands = append(operands, operand)
	}
	return operands, pos + 1, nil
}
