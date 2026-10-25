//go:build !js && linux

// d1hog: worker SINTETICO solo para pruebas de supervision de d1serve.
// Reserva N MiB (tocando cada pagina) y duerme S segundos. No se expone por
// HTTP; lo invocan los tests como worker de sustitucion.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "d1hog" {
		fs := flag.NewFlagSet("d1hog", flag.ContinueOnError)
		mb := fs.Int("mb", 128, "MiB a reservar")
		sleep := fs.Int("sleep", 30, "segundos a dormir tras reservar")
		if err := fs.Parse(os.Args[2:]); err != nil {
			os.Exit(2)
		}
		buf := make([][]byte, 0, *mb)
		for i := 0; i < *mb; i++ {
			b := make([]byte, 1<<20)
			for j := range b {
				b[j] = byte(j)
			}
			buf = append(buf, b)
		}
		fmt.Printf("hog ready mb=%d sleep=%d\n", *mb, *sleep)
		time.Sleep(time.Duration(*sleep) * time.Second)
		_ = buf
		os.Exit(0)
	}
}
