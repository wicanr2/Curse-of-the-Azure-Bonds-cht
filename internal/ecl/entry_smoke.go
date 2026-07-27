package ecl

// EntrySmokeReport is the bounded result of running one of the five
// vm_init_ecl command-set entries. Err is kept per entry so one unsupported
// routine does not hide the safe results of the other entries.
type EntrySmokeReport struct {
	Index   int
	Address uint16
	Start   int
	Result  RunResult
	Err     error
}

// SmokeInitializationEntries runs each decoded initialization entry with the
// same bounded input sequence. It is an analysis helper, not a replacement
// for the full DOS event loop: unknown commands remain visible in Err and no
// external PROGRAM side effect is invented.
func SmokeInitializationEntries(block []byte, count, maxSteps int, selections []uint16) ([]EntrySmokeReport, error) {
	points, _, err := EntryPoints(block, count)
	if err != nil {
		return nil, err
	}
	reports := make([]EntrySmokeReport, 0, len(points))
	for index, address := range points {
		start := int(address) - CodeAddressBase
		result, runErr := RunSubsetInteractive(block, start, maxSteps, selections)
		reports = append(reports, EntrySmokeReport{
			Index: index, Address: address, Start: start, Result: result, Err: runErr,
		})
	}
	return reports, nil
}
