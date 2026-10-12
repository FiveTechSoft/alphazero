// Conjunto CIEGO de evaluacion: problemas generados con semilla aparte,
// deduplicados entre si y contra levels.json, y NUNCA usados para
// entrenar (el entrenamiento filtra cualquier coincidencia por clave).
// "Verificado" en el ROADMAP solo cuenta lo medido sobre este conjunto.
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
)

func pairKey(a, b *Expr) string { return Key(a) + "=>" + Key(b) }

// GenBlind genera perLevel problemas por nivel k2..k8 con semilla propia.
func GenBlind(path string, perLevel int, seed int64) {
	r := rand.New(rand.NewSource(seed))
	// claves ya usadas en el conjunto fijo habitual (no queremos solape)
	used := map[string]bool{}
	for k := 2; k <= 8; k++ {
		for _, p := range LoadProblems("levels.json", fmt.Sprint(k)) {
			used[pairKey(p.Cur, p.Target)] = true
		}
	}
	out := map[string][]map[string]json.RawMessage{}
	for k := 2; k <= 8; k++ {
		ks := fmt.Sprint(k)
		for len(out[ks]) < perLevel {
			a, b := genTrain(r, k)
			pk := pairKey(a, b)
			if used[pk] {
				continue
			}
			used[pk] = true
			fj, _ := json.Marshal(exprJSON(a))
			tj, _ := json.Marshal(exprJSON(b))
			out[ks] = append(out[ks], map[string]json.RawMessage{"a": fj, "b": tj})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(out)
	fmt.Printf("[genblind] %d problemas/nivel x 7 niveles en %s (semilla %d, sin solape con levels.json)\n",
		perLevel, path, seed)
}

// LoadBlindKeys: claves del conjunto ciego para excluirlas del entrenamiento.
func LoadBlindKeys(path string) map[string]bool {
	keys := map[string]bool{}
	for k := 2; k <= 8; k++ {
		for _, p := range LoadProblems(path, fmt.Sprint(k)) {
			keys[pairKey(p.Cur, p.Target)] = true
		}
	}
	return keys
}
