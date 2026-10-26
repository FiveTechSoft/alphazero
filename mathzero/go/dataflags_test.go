package main

import (
	"os"
	"testing"
)

func restoreDataGlobals(t *testing.T) {
	t.Helper()
	prevDir, prevLib, prevUso := DirFix, LemmaLib, LemmaUso
	prevNoSave, prevDataDir := NoSave, DataDir
	t.Cleanup(func() {
		DirFix, LemmaLib, LemmaUso = prevDir, prevLib, prevUso
		NoSave, DataDir = prevNoSave, prevDataDir
	})
}

func listNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names
}

func TestParseDataFlags(t *testing.T) {
	restoreDataGlobals(t)
	NoSave, DataDir = false, ""

	dir := t.TempDir()
	rest, ok := parseDataFlags([]string{"1", "-nosave", "-datadir", dir, "-json"})
	if !ok {
		t.Fatal("parseDataFlags rechazo flags validos")
	}
	if !NoSave || DataDir != dir {
		t.Fatalf("flags mal parseados: NoSave=%v DataDir=%q", NoSave, DataDir)
	}
	if len(rest) != 2 || rest[0] != "1" || rest[1] != "-json" {
		t.Fatalf("posicionales mal conservados: %v", rest)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("-datadir no se creo: %v", err)
	}

	NoSave, DataDir = false, ""
	dir2 := t.TempDir() + "/sub"
	rest, ok = parseDataFlags([]string{"-datadir=" + dir2, "3"})
	if !ok || DataDir != dir2 || NoSave || len(rest) != 1 || rest[0] != "3" {
		t.Fatalf("-datadir=DIR mal parseado: ok=%v NoSave=%v DataDir=%q rest=%v", ok, NoSave, DataDir, rest)
	}
	if info, err := os.Stat(dir2); err != nil || !info.IsDir() {
		t.Fatalf("subdirectorio de -datadir no se creo: %v", err)
	}

	if _, ok = parseDataFlags([]string{"-datadir"}); ok {
		t.Fatal("parseDataFlags acepto -datadir sin valor")
	}
}

// -nosave debe recorrer el pipeline (evaluar y entrenar) sin tocar el arbol.
func TestCurveCycleNoSaveNoEscriben(t *testing.T) {
	restoreDataGlobals(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)

	CurveMain([]string{"1", "1", "5", "-nosave"})
	CycleMain([]string{"1", "1", "5", "-nosave"})

	if names := listNames(t, dir); len(names) != 0 {
		t.Fatalf("-nosave escribio ficheros: %v", names)
	}
}

// -datadir debe redirigir las escrituras fuera del directorio de trabajo.
func TestDatadirRedirigeEscrituras(t *testing.T) {
	restoreDataGlobals(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)

	out := t.TempDir()
	CycleMain([]string{"1", "1", "5", "-datadir", out})

	if names := listNames(t, dir); len(names) != 0 {
		t.Fatalf("-datadir dejo ficheros en el cwd: %v", names)
	}
	names := listNames(t, out)
	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	for _, want := range []string{"model_math.json", "lemas.json"} {
		if !found[want] {
			t.Fatalf("-datadir no contiene %s; tiene %v", want, names)
		}
	}
}
