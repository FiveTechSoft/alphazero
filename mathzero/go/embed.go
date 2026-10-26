// Datos embebidos: el ejecutable funciona solo, sin ficheros al lado.
// Si levels.json/blind.json existen en el directorio (o en el padre), se
// usan esos (mismo comportamiento de siempre); si no, los embebidos.
package main

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed blind.json
var embeddedBlind []byte

//go:embed levels.json
var embeddedLevels []byte

//go:embed model_math.json
var embeddedModel []byte

//go:embed lemas.json
var embeddedLemas []byte

//go:embed problems.json
var embeddedProblems []byte

//go:embed ttt_model.json
var embeddedTTT []byte

func readDataFile(path string) []byte {
	p := path
	if DataDir != "" && !filepath.IsAbs(path) {
		p = filepath.Join(DataDir, path)
	}
	if data, err := os.ReadFile(p); err == nil {
		return data
	}
	if DataDir == "" {
		if data, err := os.ReadFile("../" + path); err == nil {
			return data
		}
	}
	// Con -datadir se aceptan los datos embebidos para no depender de copias.
	switch filepath.Base(path) {
	case "blind.json":
		return embeddedBlind
	case "levels.json":
		return embeddedLevels
	case "model_math.json":
		return embeddedModel
	case "lemas.json":
		return embeddedLemas
	case "problems.json":
		return embeddedProblems
	case "ttt_model.json":
		return embeddedTTT
	}
	return nil
}
