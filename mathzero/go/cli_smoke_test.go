package main

import (
	"os"
	"testing"
)

// Los comandos de solo evaluacion, los no-op y los argumentos invalidos
// no deben crear ni modificar ficheros en el directorio de trabajo.
func TestComandosInvalidosONoOpNoEscriben(t *testing.T) {
	prevDir, prevLib, prevUso := DirFix, LemmaLib, LemmaUso
	defer func() { DirFix, LemmaLib, LemmaUso = prevDir, prevLib, prevUso }()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)

	CurveMain([]string{"x"})
	CurveMain([]string{"0"})
	CycleMain([]string{"1", "x"})
	CycleMain([]string{"0"})
	ABMain([]string{"e2", "x"})
	EvalFail([]string{"levels.json", "model_math.json", "100", "100", "8", "2"})
	EvalBlindMain([]string{"0"})
	EvalBlindMain([]string{"1", "modelo-inexistente.json"})
	EvalBlindMain([]string{"1"})

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("se escribieron ficheros: %v", names)
	}
}
