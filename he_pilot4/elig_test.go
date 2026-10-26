package main

// Unit tests on ARTIFICIAL data only: no search, zero decision expansions.

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

// independent reference: repeated selection of the best remaining candidate by explicit comparison.
func refChoose(elig []bool, score, noise []float64, idx []int, k int) ([]int, bool) {
	n := len(elig)
	used := make([]bool, n)
	better := func(i, j int) bool { // is i strictly better than j
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
	}
	var out []int
	for len(out) < k-1 {
		best := -1
		for i := 0; i < n; i++ {
			if used[i] {
				continue
			}
			if best < 0 || better(i, best) {
				best = i
			}
		}
		used[best] = true
		out = append(out, best)
	}
	slot := -1
	for pass := 0; pass < 2 && slot < 0; pass++ {
		for p := 0; p < n; p++ {
			ix := idx[p]
			if used[ix] {
				continue
			}
			if pass == 0 && !elig[ix] {
				continue
			}
			slot = ix
			break
		}
	}
	ne := 0
	for _, e := range elig {
		if e {
			ne++
		}
	}
	return append(out, slot), ne < k
}

func TestChooseTieredMatchesReference(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for c := 0; c < 3000; c++ {
		n := 24
		pe := []float64{0, 0.05, 0.1, 0.2, 0.5, 0.9, 1}[c%7]
		elig := make([]bool, n)
		score := make([]float64, n)
		noise := make([]float64, n)
		for i := range elig {
			elig[i] = r.Float64() < pe
			score[i] = math.Round(r.NormFloat64()*3) / 2 // many exact ties on purpose
			noise[i] = r.Float64()
			if c%11 == 0 {
				noise[i] = 0.5 // exact noise ties too: index decides
			}
		}
		idx := r.Perm(n)
		g, gf := chooseTiered(elig, score, noise, idx, pickM)
		w, wf := refChoose(elig, score, noise, idx, pickM)
		if gf != wf || len(g) != len(w) {
			t.Fatalf("case %d: fallback/len mismatch %v %v", c, g, w)
		}
		for i := range g {
			if g[i] != w[i] {
				t.Fatalf("case %d: %v != %v", c, g, w)
			}
		}
		seen := map[int]bool{}
		ne := 0
		for _, e := range elig {
			if e {
				ne++
			}
		}
		for _, x := range g {
			if seen[x] {
				t.Fatalf("case %d: duplicate choice %v", c, g)
			}
			seen[x] = true
		}
		if gf != (ne < pickM) {
			t.Fatalf("case %d: fallback flag wrong", c)
		}
		// tier priority: if at least pickM eligible, every choice is eligible
		if ne >= pickM {
			for _, x := range g {
				if !elig[x] {
					t.Fatalf("case %d: deprioritised chosen although %d eligible", c, ne)
				}
			}
		}
		// no deprioritised candidate before eligible ones are exhausted, whatever the score scale
		sc2 := make([]float64, n)
		for i := range sc2 {
			sc2[i] = score[i]
			if elig[i] {
				sc2[i] -= 1e9
			} else {
				sc2[i] += 1e9
			}
		}
		g2, _ := chooseTiered(elig, sc2, noise, idx, pickM)
		if ne >= pickM {
			for _, x := range g2 {
				if !elig[x] {
					t.Fatalf("case %d: score scale overrode tier", c)
				}
			}
		}
	}
}

func TestRandomSlotComesFromIdxOrder(t *testing.T) {
	elig := []bool{true, true, true, true, false}
	score := []float64{4, 3, 2, 1, 9}
	noise := make([]float64, 5)
	idx := []int{4, 3, 2, 1, 0}
	g, fb := chooseTiered(elig, score, noise, idx, 3)
	if fb || g[0] != 0 || g[1] != 1 || g[2] != 3 {
		t.Fatalf("got %v fb=%v (expected top two by score 0,1 and first eligible in idx order not chosen = 3)", g, fb)
	}
	// fewer than 3 eligible: fallback, deprioritised fill by idx order for the slot
	elig = []bool{true, true, false, false, false}
	g, fb = chooseTiered(elig, score, noise, idx, 3)
	if !fb || g[0] != 0 || g[1] != 1 || g[2] != 4 {
		t.Fatalf("got %v fb=%v", g, fb)
	}
	// one eligible: ridge picks the eligible one, then the best-scoring deprioritised one
	elig = []bool{false, true, false, false, false}
	g, fb = chooseTiered(elig, score, noise, idx, 3)
	if !fb || g[0] != 1 || g[1] != 4 || g[2] != 3 {
		t.Fatalf("got %v fb=%v", g, fb)
	}
}

func TestPickLearnedEligRNGConsumption(t *testing.T) {
	for seed := int64(1); seed <= 50; seed++ {
		n := poolK
		pool := make([]Edit, n)
		F := make([][]float64, n)
		r := rand.New(rand.NewSource(seed + 1000))
		for i := range pool {
			pool[i].Type = r.Intn(4)
			F[i] = make([]float64, nf)
			F[i][pool[i].Type] = 1
			F[i][4] = r.Float64()
			if r.Float64() < 0.3 && pool[i].Type != 2 {
				F[i][5] = 1
			}
			F[i][7] = r.Float64()
		}
		w := make([]float64, nf+1)
		for i := range w {
			w[i] = r.NormFloat64()
		}
		ar := rand.New(rand.NewSource(seed))
		idx := ar.Perm(n)
		twin := rand.New(rand.NewSource(seed))
		_ = twin.Perm(n)
		noise := make([]float64, n)
		for i := range noise {
			noise[i] = twin.Float64()
		}
		got, gf := pickLearnedElig(pool, F, w, ar, idx, pickM)
		elig := make([]bool, n)
		score := make([]float64, n)
		for i := range pool {
			elig[i] = !(pool[i].Type == 2 || F[i][5] == 1)
			score[i] = predW(w, F[i])
		}
		want, wf := refChoose(elig, score, noise, idx, pickM)
		if gf != wf {
			t.Fatalf("seed %d fallback mismatch", seed)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("seed %d: %v != %v", seed, got, want)
			}
		}
		if ar.Float64() != twin.Float64() {
			t.Fatalf("seed %d: consumed a number of draws different from exactly one Float64 per pool item", seed)
		}
	}
}

func TestEligibleEdit(t *testing.T) {
	f := make([]float64, nf)
	if !eligibleEdit(Edit{Type: 0}, f) || !eligibleEdit(Edit{Type: 1}, f) || !eligibleEdit(Edit{Type: 3}, f) {
		t.Fatal("non-remove, non-redundant must be eligible")
	}
	if eligibleEdit(Edit{Type: 2}, f) {
		t.Fatal("remove must not be eligible")
	}
	f[5] = 1
	if eligibleEdit(Edit{Type: 0}, f) {
		t.Fatal("redundant must not be eligible")
	}
}

// independent ridge solution: Cholesky on (X'X + I) w = X'y with a bias column, lambda = 1 on all coefficients as in Ridge.fit.
func cholRidge(X [][]float64, Y []float64) []float64 {
	n := nf + 1
	A := make([][]float64, n)
	b := make([]float64, n)
	for i := range A {
		A[i] = make([]float64, n)
		A[i][i] = 1
	}
	for k, x := range X {
		xx := append([]float64{1}, x...)
		for i := 0; i < n; i++ {
			b[i] += xx[i] * Y[k]
			for j := 0; j < n; j++ {
				A[i][j] += xx[i] * xx[j]
			}
		}
	}
	L := make([][]float64, n)
	for i := range L {
		L[i] = make([]float64, n)
	}
	for i := 0; i < n; i++ {
		for j := 0; j <= i; j++ {
			s := A[i][j]
			for k := 0; k < j; k++ {
				s -= L[i][k] * L[j][k]
			}
			if i == j {
				L[i][j] = math.Sqrt(s)
			} else {
				L[i][j] = s / L[j][j]
			}
		}
	}
	z := make([]float64, n)
	for i := 0; i < n; i++ {
		s := b[i]
		for k := 0; k < i; k++ {
			s -= L[i][k] * z[k]
		}
		z[i] = s / L[i][i]
	}
	w := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		s := z[i]
		for k := i + 1; k < n; k++ {
			s -= L[k][i] * w[k]
		}
		w[i] = s / L[i][i]
	}
	return w
}

func TestRidgeMatchesCholesky(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	var S []Sample
	for i := 0; i < 300; i++ {
		x := make([]float64, nf)
		for j := range x {
			x[j] = r.NormFloat64()
		}
		S = append(S, Sample{x, r.NormFloat64()})
	}
	w := fitW(S, false, nil)
	X := make([][]float64, len(S))
	Y := make([]float64, len(S))
	for i, s := range S {
		X[i], Y[i] = s.X, s.Y
	}
	c := cholRidge(X, Y)
	for i := range w {
		if math.Abs(w[i]-c[i]) > 1e-8 {
			t.Fatalf("coef %d: %g vs %g", i, w[i], c[i])
		}
	}
}

func TestFitEligFilterAndShuffle(t *testing.T) {
	dir := t.TempDir()
	r := rand.New(rand.NewSource(9))
	var S []Sample
	nElig := 0
	for i := 0; i < 400; i++ {
		x := make([]float64, nf)
		ty := r.Intn(4)
		x[ty] = 1
		for j := 6; j < nf; j++ {
			x[j] = r.Float64()
		}
		x[4] = r.Float64()
		if r.Float64() < 0.25 && ty != 2 {
			x[5] = 1
		}
		if ty != 2 && x[5] == 0 {
			nElig++
		}
		S = append(S, Sample{x, float64(r.Intn(9) - 4)})
	}
	tf := filepath.Join(dir, "train.json")
	wf := filepath.Join(dir, "w.json")
	b, _ := json.Marshal(S)
	os.WriteFile(tf, b, 0o644)
	runFitElig(tf, wf)
	var m map[string]interface{}
	bb, _ := os.ReadFile(wf)
	json.Unmarshal(bb, &m)
	if int(m["n_rows_eligible"].(float64)) != nElig || int(m["n_rows_total"].(float64)) != 400 {
		t.Fatalf("filter count %v vs %d", m["n_rows_eligible"], nElig)
	}
	var X [][]float64
	var Y []float64
	for _, s := range S {
		if s.X[2] == 0 && s.X[5] == 0 {
			X = append(X, s.X)
			Y = append(Y, s.Y)
		}
	}
	c := cholRidge(X, Y)
	got := m["learned_elig"].([]interface{})
	for i := range c {
		if math.Abs(got[i].(float64)-c[i]) > 1e-8 {
			t.Fatalf("learned_elig coef %d differs from independent solution", i)
		}
	}
	sh := m["learned_elig_shuf"].([]interface{})
	same := true
	for i := range c {
		if math.Abs(sh[i].(float64)-c[i]) > 1e-9 {
			same = false
		}
	}
	if same {
		t.Fatal("shuffled-label weights equal the real ones")
	}
}
