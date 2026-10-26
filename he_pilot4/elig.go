package main

// Pilot 4 additions (kept in a separate file; main.go only gets the arm cases and the fitelig mode hook).

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
)

// eligible = NOT deprioritised, with the same definition freqrule and excl_random use:
// deprioritised iff the edit is a remove (type 2) or its expansion is already contained in a library action (feature 5 == 1).
func eligibleEdit(e Edit, f []float64) bool { return !(e.Type == 2 || f[5] == 1) }

// chooseTiered: tier (eligible first) has strict priority over the ridge score; noise breaks exact score ties; index is the last tie-break.
// The first pickM-1 choices are the top of that order. The last choice is a random slot: the first index in idx (the arm's random
// permutation of the pool, drawn before selection) that is not chosen and is eligible; if no eligible candidate remains, the first
// index in idx that is not chosen. fallback is true iff fewer than pickM candidates are eligible (some deprioritised candidate is chosen).
func chooseTiered(elig []bool, score, noise []float64, idx []int, k int) (chosen []int, fallback bool) {
	n := len(elig)
	order := make([]int, n)
	ne := 0
	for i := range order {
		order[i] = i
		if elig[i] {
			ne++
		}
	}
	sort.SliceStable(order, func(a, b int) bool {
		i, j := order[a], order[b]
		if elig[i] != elig[j] {
			return elig[i]
		}
		if score[i] != score[j] {
			return score[i] > score[j]
		}
		if noise[i] != noise[j] {
			return noise[i] > noise[j]
		}
		return i < j
	})
	chosen = append(chosen, order[:k-1]...)
	in := func(x int) bool {
		for _, c := range chosen {
			if c == x {
				return true
			}
		}
		return false
	}
	slot := -1
	for _, ix := range idx {
		if !in(ix) && elig[ix] {
			slot = ix
			break
		}
	}
	if slot < 0 {
		for _, ix := range idx {
			if !in(ix) {
				slot = ix
				break
			}
		}
	}
	chosen = append(chosen, slot)
	return chosen, ne < k
}

// pickLearnedElig draws exactly len(pool) values from ar (one Float64 per pool item, in pool order) after the caller's ar.Perm.
func pickLearnedElig(pool []Edit, F [][]float64, w []float64, ar *rand.Rand, idx []int, k int) ([]int, bool) {
	elig := make([]bool, len(pool))
	score := make([]float64, len(pool))
	noise := make([]float64, len(pool))
	for i, e := range pool {
		elig[i] = eligibleEdit(e, F[i])
		score[i] = predW(w, F[i])
		noise[i] = ar.Float64()
	}
	return chooseTiered(elig, score, noise, idx, k)
}

// fitelig mode: ridge (lambda=1, same Ridge.fit as pilot 2) on the eligible rows of train.json; plus the label-permuted control on the same rows.
func runFitElig(trainFile, wfile string) {
	b, err := os.ReadFile(trainFile)
	if err != nil {
		fmt.Println("cannot read", trainFile)
		os.Exit(1)
	}
	var all []Sample
	if err := json.Unmarshal(b, &all); err != nil {
		fmt.Println("bad json", err)
		os.Exit(1)
	}
	var S []Sample
	for _, s := range all {
		if len(s.X) != nf {
			fmt.Println("bad feature length")
			os.Exit(1)
		}
		if s.X[2] == 0 && s.X[5] == 0 { // not remove (type one-hot index 2) and not redundant (feature 5)
			S = append(S, s)
		}
	}
	w := fitW(S, false, nil)
	ws := fitW(S, true, rand.New(rand.NewSource(424243)))
	f, _ := os.Create(wfile)
	json.NewEncoder(f).Encode(map[string]interface{}{"learned_elig": w, "learned_elig_shuf": ws, "n_rows_total": len(all), "n_rows_eligible": len(S), "shuffle_seed": 424243})
	f.Close()
	fmt.Println("rows total", len(all), "eligible", len(S))
	fmt.Println("learned_elig", w)
	fmt.Println("learned_elig_shuf", ws)
}
