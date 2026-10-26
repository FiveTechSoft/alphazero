package main

import (
	"fmt"
	"os"
	"strconv"
)

// parseIntArg valida un argumento numerico de CLI antes de usarlo.
func parseIntArg(name, raw string, min int) (int, bool) {
	v, err := strconv.Atoi(raw)
	if err != nil || v < min {
		fmt.Fprintf(os.Stderr, "argumento invalido: %s=%q (entero >= %d)\n", name, raw, min)
		return 0, false
	}
	return v, true
}

func parseInt64Arg(name, raw string) (int64, bool) {
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "argumento invalido: %s=%q (entero de 64 bits)\n", name, raw)
		return 0, false
	}
	return v, true
}

// parseFloatArg valida un argumento flotante de CLI antes de usarlo.
func parseFloatArg(name, raw string) (float64, bool) {
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 || v >= 1 {
		fmt.Fprintf(os.Stderr, "argumento invalido: %s=%q (flotante en [0,1))\n", name, raw)
		return 0, false
	}
	return v, true
}
