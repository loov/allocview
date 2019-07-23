package main

import "fmt"

func SizeToString(bytes int64) string {
	abs := bytes
	if abs < 0 {
		abs = -abs
	}

	switch {
	case abs < (1<<10)*2/3:
		return fmt.Sprintf("%dB", bytes)
	case abs < (1<<20)*2/3:
		return fmt.Sprintf("%0.2fKB", float64(bytes)/float64(1<<10))
	case abs < (1<<30)*2/3:
		return fmt.Sprintf("%0.2fMB", float64(bytes)/float64(1<<20))
	case abs < (1<<40)*2/3:
		return fmt.Sprintf("%0.2fGB", float64(bytes)/float64(1<<30))
	case abs < (1<<50)*2/3:
		return fmt.Sprintf("%0.2fTB", float64(bytes)/float64(1<<40))
	default:
		return fmt.Sprintf("%0.2fPB", float64(bytes)/float64(1<<50))
	}
}
