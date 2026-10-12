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
)

type Lemma struct {
	From, To       *Expr
	FromKey, ToKey string
	Saving         int // pasos del camino original que comprime
	Uses           int
}

type lemmaJSON struct {
	From   json.RawMessage `json:"from"`
	To     json.RawMessage `json:"to"`
	Saving int             `json:"saving"`
	Uses   int             `json:"uses"`
}

type Library struct {
	Lemmas  []*Lemma
	ByKey   map[string][]int // clave estructural -> lemas que la contienen
	MaxSize int
}

// Biblioteca global; nil = desactivada (para ablaciones).
var LemmaLib *Library

func NewLibrary() *Library {
	return &Library{ByKey: map[string][]int{}, MaxSize: 12}
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
	if ka == kb || saving < 2 {
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

func (L *Library) Save(path string) {
	out := make([]lemmaJSON, 0, len(L.Lemmas))
	for _, l := range L.Lemmas {
		fj, _ := json.Marshal(exprJSON(l.From))
		tj, _ := json.Marshal(exprJSON(l.To))
		out = append(out, lemmaJSON{From: fj, To: tj, Saving: l.Saving, Uses: l.Uses})
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
	f, err := os.Open(path)
	if err != nil {
		return L
	}
	defer f.Close()
	var raw []lemmaJSON
	if json.NewDecoder(f).Decode(&raw) != nil {
		return L
	}
	for _, lj := range raw {
		a, b := parseExpr(lj.From), parseExpr(lj.To)
		if L.Insert(a, b, lj.Saving) {
			L.Lemmas[len(L.Lemmas)-1].Uses = lj.Uses
		}
	}
	return L
}
