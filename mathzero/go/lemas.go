// Biblioteca persistente de lemas: identidades YA DEMOSTRADAS por el
// motor (camino start->target verificado con el juez exacto) que se
// guardan en disco y se reutilizan como reglas extra en problemas
// nuevos. Es la memoria que crece sola del core: cada problema
// resuelto anade reescrituras nuevas. Criterio de verdad: canonEq
// (forma canonica de polinomio) en el momento de insertar.
package main

import (
	"encoding/json"
	"os"
	"sort"
)

type Lemma struct {
	From, To             *Expr
	FromKey, ToKey       string
	Saving               int // pasos del camino original que comprime
	Uses                 int
	MatchingAttempts     int
	WinningApplications  int
	FailedGeneratedNodes int
	TrainingProblems     int
}

type lemmaJSON struct {
	From                 json.RawMessage `json:"from"`
	To                   json.RawMessage `json:"to"`
	Saving               int             `json:"saving"`
	SavedStepsBase       int             `json:"pasos_ahorrados"`
	Uses                 int             `json:"uses"`
	MatchingAttempts     int             `json:"intentos_matching"`
	WinningApplications  int             `json:"aplicaciones_en_camino_ganador"`
	FailedGeneratedNodes int             `json:"nodos_generados_fallidos"`
	TrainingProblems     int             `json:"problemas_expuestos"`
}

type Library struct {
	Lemmas             []*Lemma
	ByKey              map[string][]int // clave estructural -> lemas que la contienen
	MaxSize            int
	ActiveCap          int
	UtilityLambda      float64
	UtilityMinProblems int
}

// Biblioteca global; nil = desactivada (para ablaciones).
var LemmaLib *Library

// Contador de uso de lemas en caminos ganadores (nil = sin medicion).
// Lo rellena GuidedSolve solo cuando la biblioteca esta activa.
var LemmaUso map[int]int

func NewLibrary() *Library {
	return &Library{ByKey: map[string][]int{}, MaxSize: 12, ActiveCap: LemmaActiveCap, UtilityLambda: LemmaUtilityLambda, UtilityMinProblems: LemmaUtilityMinProblems}
}

func exprJSON(e *Expr) map[string]any {
	switch e.Op {
	case "const":
		return map[string]any{"op": "const", "n": e.N}
	case "var":
		return map[string]any{"op": "var", "s": e.S}
	case "neg":
		return map[string]any{"op": "neg", "a": exprJSON(e.A)}
	}
	return map[string]any{"op": e.Op, "a": exprJSON(e.A), "b": exprJSON(e.B)}
}

// Insert verifica la identidad con el juez exacto y evita duplicados.
func (L *Library) Insert(a, b *Expr, saving int) bool {
	ka, kb := Key(a), Key(b)
	if ka == kb || saving < 2 || (L.ActiveCap > 0 && len(L.Lemmas) >= L.ActiveCap) {
		return false
	}
	if size(a) > L.MaxSize || size(b) > L.MaxSize {
		return false
	}
	if !canonEq(Canon(a), Canon(b)) {
		return false // jamas deberia pasar: las reglas preservan canon
	}
	for _, i := range L.ByKey[ka] {
		l := L.Lemmas[i]
		if l.ToKey == kb || l.FromKey == kb {
			return false
		}
	}
	idx := len(L.Lemmas)
	L.Lemmas = append(L.Lemmas, &Lemma{From: a, To: b, FromKey: ka, ToKey: kb, Saving: saving})
	L.ByKey[ka] = append(L.ByKey[ka], idx)
	L.ByKey[kb] = append(L.ByKey[kb], idx)
	return true
}

// Pruned devuelve una biblioteca con solo los lemas de uso medido
// (uso>=1), ordenados por uso, como maximo maxK: poda la branching.
func (L *Library) Pruned(uso map[int]int, maxK int) *Library {
	type us struct{ i, u int }
	ord := make([]us, 0, len(L.Lemmas))
	for i := range L.Lemmas {
		if uso[i] >= 1 {
			ord = append(ord, us{i, uso[i]})
		}
	}
	sort.Slice(ord, func(a, b int) bool {
		if ord[a].u != ord[b].u {
			return ord[a].u > ord[b].u
		}
		return ord[a].i < ord[b].i
	})
	if L.ActiveCap > 0 && maxK > L.ActiveCap {
		maxK = L.ActiveCap
	}
	P := NewLibrary()
	P.MaxSize = L.MaxSize
	P.ActiveCap, P.UtilityLambda, P.UtilityMinProblems = L.ActiveCap, L.UtilityLambda, L.UtilityMinProblems
	for j, o := range ord {
		if j >= maxK {
			break
		}
		l := L.Lemmas[o.i]
		idx := len(P.Lemmas)
		P.Lemmas = append(P.Lemmas, l)
		P.ByKey[l.FromKey] = append(P.ByKey[l.FromKey], idx)
		P.ByKey[l.ToKey] = append(P.ByKey[l.ToKey], idx)
	}
	return P
}

func (L *Library) Save(path string) {
	out := make([]lemmaJSON, 0, len(L.Lemmas))
	for _, l := range L.Lemmas {
		fj, _ := json.Marshal(exprJSON(l.From))
		tj, _ := json.Marshal(exprJSON(l.To))
		out = append(out, lemmaJSON{From: fj, To: tj, Saving: l.Saving, SavedStepsBase: l.Saving, Uses: l.Uses, MatchingAttempts: l.MatchingAttempts, WinningApplications: l.WinningApplications, FailedGeneratedNodes: l.FailedGeneratedNodes, TrainingProblems: l.TrainingProblems})
	}
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(out)
}

func LoadLibrary(path string) *Library {
	L := NewLibrary()
	// Load all candidates before ranking; never keep merely the first96.
	L.ActiveCap = 0
	data := readDataFile(path)
	if data == nil {
		L.ActiveCap = LemmaActiveCap
		return L
	}
	var raw []lemmaJSON
	if json.Unmarshal(data, &raw) != nil {
		L.ActiveCap = LemmaActiveCap
		return L
	}
	for _, lj := range raw {
		if lj.Saving == 0 {
			lj.Saving = lj.SavedStepsBase
		}
		a, b := parseExpr(lj.From), parseExpr(lj.To)
		if L.Insert(a, b, lj.Saving) {
			l := L.Lemmas[len(L.Lemmas)-1]
			l.Uses = lj.Uses
			l.MatchingAttempts, l.WinningApplications = lj.MatchingAttempts, lj.WinningApplications
			l.FailedGeneratedNodes, l.TrainingProblems = lj.FailedGeneratedNodes, lj.TrainingProblems
		}
	}
	L.ActiveCap = LemmaActiveCap
	L.PruneUtility()
	return L
}
