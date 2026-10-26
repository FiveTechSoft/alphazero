package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

type fakeNet struct {
	priors []float64
	value  float64
}

func (f fakeNet) Predict(any, []any) ([]float64, float64) {
	return f.priors, f.value
}

func TestExpandNormalizaPriors(t *testing.T) {
	priors := make([]float64, 9)
	for i := range priors {
		priors[i] = 1
	}
	priors[0] = 3

	n := &Node{}
	v := expand(n, any(TTTState{Player: 1}), TicTacToe{}, fakeNet{priors: priors, value: math.NaN()})
	if v != 0 {
		t.Errorf("valor no finito no se neutralizo: %v", v)
	}
	sum := 0.0
	for _, p := range n.P {
		sum += p
	}
	if math.Abs(sum-1) > 1e-12 {
		t.Errorf("priors no normalizadas: suma=%v", sum)
	}
	if math.Abs(n.P[0]-3.0/11.0) > 1e-12 {
		t.Errorf("prior relativa mal conservada: %v", n.P[0])
	}
}

func TestExpandPriorsInvalidasFallbackUniforme(t *testing.T) {
	casos := []struct {
		nombre string
		priors []float64
	}{
		{"longitud", []float64{1, 2}},
		{"negativa", func() []float64 {
			p := make([]float64, 9)
			p[2] = -1
			return p
		}()},
		{"nan", func() []float64 {
			p := make([]float64, 9)
			p[3] = math.NaN()
			return p
		}()},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			n := &Node{}
			expand(n, any(TTTState{Player: 1}), TicTacToe{}, fakeNet{priors: c.priors})
			for i, p := range n.P {
				if math.Abs(p-1.0/9.0) > 1e-12 {
					t.Errorf("prior %d = %v, esperado uniforme", i, p)
				}
			}
		})
	}
}

func TestTrainCEEntradasInvalidasNoPanican(t *testing.T) {
	n := NewMLP(1)
	if got := n.TrainCE(nil, 0, 0.003); got != 0 {
		t.Errorf("sin acciones devolvio %v", got)
	}
	feats := [][]float64{make([]float64, NFeat), make([]float64, NFeat)}
	if got := n.TrainCE(feats, -1, 0.003); got != 0 {
		t.Errorf("target negativo devolvio %v", got)
	}
	if got := n.TrainCE(feats, len(feats), 0.003); got != 0 {
		t.Errorf("target fuera de rango devolvio %v", got)
	}
	if got := n.TrainCE([][]float64{{1, 2}}, 0, 0.003); got != 0 {
		t.Errorf("features cortas devolvieron %v", got)
	}
	loss := n.TrainCE(feats, 1, 0.003)
	if math.IsNaN(loss) || math.IsInf(loss, 0) {
		t.Errorf("loss no finita en entrada valida: %v", loss)
	}
}

func TestGuidedSolveYPath(t *testing.T) {
	prevDir, prevLib := DirFix, LemmaLib
	DirFix, LemmaLib = true, nil
	defer func() { DirFix, LemmaLib = prevDir, prevLib }()

	start := EMul(V("a"), EAdd(V("b"), V("c")))
	target := EAdd(EMul(V("a"), V("b")), EMul(V("a"), V("c")))

	ok, nodes, samples := GuidedSolve(nil, start, target, 200)
	if !ok {
		t.Fatal("GuidedSolve no encontro la identidad")
	}
	if nodes == 0 || len(samples) == 0 {
		t.Errorf("GuidedSolve devolvio nodos=%d muestras=%d", nodes, len(samples))
	}

	ok, nodes, path := SolvePath(nil, start, target, 200)
	if !ok {
		t.Fatal("SolvePath no encontro la identidad")
	}
	if nodes == 0 || len(path) == 0 {
		t.Errorf("SolvePath devolvio nodos=%d pasos=%d", nodes, len(path))
	}
	if got := Key(path[len(path)-1].To); got != Key(target) {
		t.Errorf("SolvePath termina en %s, esperado %s", got, Key(target))
	}
	for _, paso := range path {
		if !canonEq(Canon(paso.From), Canon(paso.To)) {
			t.Fatalf("paso %s rompe la forma canonica", paso.Rule)
		}
	}
}

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "modelo.json")
	if err := os.WriteFile(path, []byte("viejo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(path, []byte("nuevo")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "nuevo" {
		t.Errorf("contenido = %q, esperado nuevo", data)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("quedaron %d ficheros en el directorio, esperado 1", len(entries))
	}
}

func TestLibrarySaveRoundTrip(t *testing.T) {
	lib := NewLibrary()
	a, b := EMul(V("a"), C(1)), V("a")
	if !lib.Insert(a, b, 2) {
		t.Fatal("Insert rechazo un lema valido")
	}
	lib.Lemmas[0].Uses = 3
	path := filepath.Join(t.TempDir(), "lemas.json")
	lib.Save(path)
	got := LoadLibrary(path)
	if len(got.Lemmas) != 1 {
		t.Fatalf("cargados %d lemas, esperado 1", len(got.Lemmas))
	}
	if got.Lemmas[0].Uses != 3 {
		t.Errorf("Uses = %d, esperado 3", got.Lemmas[0].Uses)
	}
}

func TestLoadTTTNetValidaFormas(t *testing.T) {
	data, err := os.ReadFile("ttt_model.json")
	if err != nil {
		t.Skip("ttt_model.json no presente")
	}
	if n := LoadTTTNet(data); n == nil {
		t.Fatal("LoadTTTNet rechazo el modelo valido")
	}
	if n := LoadTTTNet([]byte(`{"w1":[],"b1":[],"wp":[],"bp":[],"wv":[]}`)); n != nil {
		t.Error("LoadTTTNet acepto un modelo sin capas")
	}
}

func TestLoadMLPRechazaFormasInvalidas(t *testing.T) {
	dir := t.TempDir()

	invalido := filepath.Join(dir, "invalido.json")
	if err := os.WriteFile(invalido, []byte(`{"w1":[],"b1":[],"w2":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if n := LoadMLP(invalido); n != nil {
		t.Error("LoadMLP acepto una red sin capas")
	}

	n := NewMLP(1)
	n.W1[0] = n.W1[0][:NFeat-1]
	data, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	malformado := filepath.Join(dir, "malformado.json")
	if err := os.WriteFile(malformado, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := LoadMLP(malformado); got != nil {
		t.Error("LoadMLP acepto una fila de W1 con longitud incorrecta")
	}
}
