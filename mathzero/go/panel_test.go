package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPanelEscribeJSONYNoTocaElCwd(t *testing.T) {
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
	PanelMain([]string{"5", "-json", "-no-ttt", "-datadir", out})

	if names := listNames(t, dir); len(names) != 0 {
		t.Fatalf("panel dejo ficheros en el cwd: %v", names)
	}
	data, err := os.ReadFile(filepath.Join(out, "panel.json"))
	if err != nil {
		t.Fatal("panel no escribio panel.json:", err)
	}
	var res panelResult
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatal("panel.json invalido:", err)
	}
	if len(res.PorNivel) != 7 || res.Problemas != 700 || res.Cap != 5 {
		t.Fatalf("panel inesperado: niveles=%d problemas=%d cap=%d", len(res.PorNivel), res.Problemas, res.Cap)
	}
	if res.TTT != nil {
		t.Fatal("-no-ttt deberia omitir el tres en raya")
	}
}

func TestPanelNoSaveNoEscribe(t *testing.T) {
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
	PanelMain([]string{"5", "-json", "-no-ttt", "-nosave", "-datadir", out})
	if names := listNames(t, out); len(names) != 0 {
		t.Fatalf("-nosave escribio ficheros: %v", names)
	}
}
