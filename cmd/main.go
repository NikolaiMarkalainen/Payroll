package main

import (
	"os"

	"github.com/NikolaiMarkalainen/payroll/internal/ui"
)

func main() {
	ui.RunWith(ui.Options{Demo: hasFlag(os.Args[1:], "-demo", "--demo")})
}

func hasFlag(args []string, names ...string) bool {
	set := map[string]struct{}{}
	for _, n := range names {
		set[n] = struct{}{}
	}
	for _, a := range args {
		if _, ok := set[a]; ok {
			return true
		}
	}
	return false
}
