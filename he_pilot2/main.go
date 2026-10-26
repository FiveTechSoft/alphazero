// HypothesisEngine pilot: does a selector that learns from edit history pick better
// library edits than random / simple controls at equal total search cost?
// Hand-programmed typed edits over an enumerative program-search solver. No neural claim.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

const M = 10007

var inputs = []int64{1, 2, 3, 5, 8}

// primitives 0..7 base, 8..11 reservoir
var primNames = []string{"inc", "dec", "dbl", "tri", "half", "neg", "sq", "add5", "mul5", "mul7", "add11", "sub3"}

func applyPrim(p int, x int64) int64 {
	switch p {
	case 0:
		return (x + 1) % M
	case 1:
		return (x + M - 1) % M
	case 2:
		return (x * 2) % M
	case 3:
		return (x * 3) % M
	case 4:
		return x / 2
	case 5:
		return (M - x) % M
	case 6:
		return (x * x) % M
	case 7:
		return (x + 5) % M
	case 8:
		return (x * 5) % M
	case 9:
		return (x * 7) % M
	case 10:
		return (x + 11) % M
	default:
		return (x + M - 3) % M
	}
}

type Action []int // primitive sequence
type Lib struct {
	Acts []Action
}

func (l *Lib) clone() *Lib {
	n := &Lib{Acts: make([]Action, len(l.Acts))}
	copy(n.Acts, l.Acts)
	return n
}
func key(a Action) string {
	var sb strings.Builder
	for _, p := range a {
		sb.WriteByte(byte('a' + p))
	}
	return sb.String()
}
func (l *Lib) has(a Action) bool {
	k := key(a)
	for _, b := range l.Acts {
		if key(b) == k {
			return true
		}
	}
	return false
}

type World struct {
	motifs []Action
}

type Task struct {
	prog   Action
	target [5]int64
}

func evalProg(p Action) [5]int64 {
	var out [5]int64
	for i, x := range inputs {
		v := x
		for _, q := range p {
			v = applyPrim(q, v)
		}
		out[i] = v
	}
	return out
}

func newWorld(r *rand.Rand) *World {
	w := &World{}
	for i := 0; i < 8; i++ {
		ln := 2 + r.Intn(2)
		a := make(Action, ln)
		for j := range a {
			if i < 2 && j == 0 {
				a[j] = 8 + r.Intn(4) // reservoir prim
			} else {
				a[j] = r.Intn(8)
			}
		}
		w.motifs = append(w.motifs, a)
	}
	return w
}

func (w *World) task(r *rand.Rand) Task {
	u := r.Float64()
	n := 1
	if u > 0.4 {
		n = 2
	}
	if u > 0.8 {
		n = 3
	}
	var p Action
	if r.Float64() < 0.3 {
		p = append(p, r.Intn(8))
	}
	for i := 0; i < n; i++ {
		p = append(p, w.motifs[r.Intn(len(w.motifs))]...)
	}
	return Task{prog: p, target: evalProg(p)}
}

type actEff struct {
	out [5]int64
}

// solve: BFS over behaviour states; cost = expansions (action applications), cap per task.
func solve(lib *Lib, t Task, cap int, maxDepth int) (bool, int, Action) {
	type node struct {
		st   [5]int64
		prog Action
	}
	start := [5]int64{}
	copy(start[:], inputs)
	if start == t.target {
		return true, 0, nil
	}
	seen := map[[5]int64]struct{}{start: {}}
	frontier := []node{{st: start}}
	cost := 0
	for d := 0; d < maxDepth; d++ {
		var next []node
		for _, nd := range frontier {
			for _, a := range lib.Acts {
				if cost >= cap {
					return false, cost, nil
				}
				cost++
				var ns [5]int64
				for i := 0; i < 5; i++ {
					v := nd.st[i]
					for _, q := range a {
						v = applyPrim(q, v)
					}
					ns[i] = v
				}
				if _, ok := seen[ns]; ok {
					continue
				}
				seen[ns] = struct{}{}
				np := make(Action, len(nd.prog), len(nd.prog)+len(a))
				copy(np, nd.prog)
				np = append(np, a...)
				if ns == t.target {
					return true, cost, np
				}
				next = append(next, node{ns, np})
			}
		}
		frontier = next
		if len(frontier) == 0 {
			break
		}
	}
	return false, cost, nil
}

const taskCap = 1500
const maxDepth = 8

var left = 1 << 60 // remaining decision budget (expansions); evaluation is truncated at 0
var exhausted bool
var ops int64 // non-expansion work proxy: ridge/feature inner-loop steps (not part of budget)

func evalLib(lib *Lib, tasks []Task, corpus *[]Action) (solved, cost int) {
	for _, t := range tasks {
		cp := taskCap
		if left < cp {
			cp = left
		}
		if cp <= 0 {
			exhausted = true
			return
		}
		ok, c, p := solve(lib, t, cp, maxDepth)
		left -= c
		if left <= 0 {
			exhausted = true
		}
		cost += c
		if ok {
			solved++
			if corpus != nil && p != nil {
				*corpus = append(*corpus, p)
			}
		}
	}
	return
}

// ---- edits ----
type Edit struct {
	Type int // 0 macro2, 1 macro3, 2 remove, 3 addprim
	Act  Action
	Idx  int
}

func (l *Lib) apply(e Edit) *Lib {
	n := l.clone()
	switch e.Type {
	case 0, 1, 3:
		n.Acts = append(n.Acts, e.Act)
	case 2:
		n.Acts = append(n.Acts[:e.Idx], n.Acts[e.Idx+1:]...)
	}
	return n
}

func gen(lib *Lib, r *rand.Rand, k int) []Edit {
	var out []Edit
	tries := 0
	for len(out) < k && tries < 1000 {
		tries++
		u := r.Float64()
		switch {
		case u < 0.35 && len(lib.Acts) < 40: // macro2 over lib actions
			a := lib.Acts[r.Intn(len(lib.Acts))]
			b := lib.Acts[r.Intn(len(lib.Acts))]
			m := append(append(Action{}, a...), b...)
			if len(m) > 8 || lib.has(m) {
				continue
			}
			out = append(out, Edit{Type: 0, Act: m})
		case u < 0.7 && len(lib.Acts) < 40: // macro3 over base+reservoir prims
			m := Action{r.Intn(12), r.Intn(12), r.Intn(12)}
			if lib.has(m) {
				continue
			}
			out = append(out, Edit{Type: 1, Act: m})
		case u < 0.85 && len(lib.Acts) > 4:
			out = append(out, Edit{Type: 2, Idx: r.Intn(len(lib.Acts))})
		default:
			p := Action{8 + r.Intn(4)}
			if lib.has(p) || len(lib.Acts) >= 40 {
				continue
			}
			out = append(out, Edit{Type: 3, Act: p})
		}
	}
	return out
}

func countSub(corpus []Action, sub Action) float64 {
	if len(corpus) == 0 {
		return 0
	}
	c := 0
	for _, p := range corpus {
		ops += int64(len(p))
		for i := 0; i+len(sub) <= len(p); i++ {
			ok := true
			for j := range sub {
				if p[i+j] != sub[j] {
					ok = false
					break
				}
			}
			if ok {
				c++
				break
			}
		}
	}
	return float64(c) / float64(len(corpus))
}

const nf = 11

func feats(lib *Lib, e Edit, corpus []Action, rate float64) []float64 {
	f := make([]float64, nf)
	f[e.Type] = 1
	var exp Action
	if e.Type == 2 {
		exp = lib.Acts[e.Idx]
	} else {
		exp = e.Act
	}
	fr := countSub(corpus, exp)
	f[4] = fr
	// redundancy: expansion is substring of an existing action (not for remove)
	red := 0.0
	if e.Type != 2 {
		ks := key(exp)
		for _, a := range lib.Acts {
			if strings.Contains(key(a), ks) {
				red = 1
			}
		}
	}
	f[5] = red
	f[6] = float64(len(exp)) / 8
	f[7] = float64(len(lib.Acts)) / 40
	f[8] = rate
	f[9] = fr * f[7]
	f[10] = fr * rate
	return f
}

// ridge
type Ridge struct {
	X [][]float64
	Y []float64
	w []float64
}

func (r *Ridge) fit() {
	n := nf + 1
	A := make([][]float64, n)
	b := make([]float64, n)
	for i := range A {
		A[i] = make([]float64, n)
		A[i][i] = 1.0
	}
	for k, x := range r.X {
		xx := append([]float64{1}, x...)
		for i := 0; i < n; i++ {
			b[i] += xx[i] * r.Y[k]
			ops += int64(n)
			for j := 0; j < n; j++ {
				A[i][j] += xx[i] * xx[j]
			}
		}
	}
	// gauss
	for i := 0; i < n; i++ {
		p := i
		for j := i + 1; j < n; j++ {
			if math.Abs(A[j][i]) > math.Abs(A[p][i]) {
				p = j
			}
		}
		A[i], A[p] = A[p], A[i]
		b[i], b[p] = b[p], b[i]
		for j := i + 1; j < n; j++ {
			m := A[j][i] / A[i][i]
			for k := i; k < n; k++ {
				A[j][k] -= m * A[i][k]
			}
			b[j] -= m * b[i]
		}
	}
	w := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		s := b[i]
		for j := i + 1; j < n; j++ {
			s -= A[i][j] * w[j]
		}
		w[i] = s / A[i][i]
	}
	r.w = w
}
func (r *Ridge) pred(x []float64) float64 {
	s := r.w[0]
	for i, v := range x {
		s += r.w[i+1] * v
	}
	return s
}

type Result struct {
	Seed       int
	Arm        string
	AuditSolve int
	AuditCost  int
	Rounds     int
	LibSize    int
	Accepts    int
	Evals      int
	Spent      int
	WallMs     int64
	Ops        int64
}

const batchN = 40
const poolK = 24
const pickM = 3

func baseLib() *Lib {
	l := &Lib{}
	for i := 0; i < 8; i++ {
		l.Acts = append(l.Acts, Action{i})
	}
	return l
}

type Sample struct {
	X []float64
	Y float64
}

var transferW []float64 // frozen weights (len nf+1) for transfer arms
var collect *[]Sample   // when non-nil, run in label-collection mode

func predW(w []float64, x []float64) float64 {
	s := w[0]
	for i, v := range x {
		s += w[i+1] * v
	}
	return s
}

func hashLib(seed, round int, lib *Lib) int64 {
	h := uint64(1469598103934665603) ^ uint64(seed)*0x9E3779B97F4A7C15 ^ uint64(round)*0xC2B2AE3D27D4EB4F
	for _, a := range lib.Acts {
		for _, c := range []byte(key(a)) {
			h = (h ^ uint64(c)) * 1099511628211
		}
		h = (h ^ 255) * 1099511628211
	}
	return int64(h >> 1)
}

func runArm(seed int, arm string, budget int, audit []Task) Result {
	wr := rand.New(rand.NewSource(int64(seed)*1000 + 1))
	w := newWorld(wr)
	sr := rand.New(rand.NewSource(int64(seed)*1000 + 7)) // task stream (same across arms; batches depend on round only)
	ar := rand.New(rand.NewSource(int64(seed)*1000 + 99 + int64(len(arm)*7919)))
	lib := baseLib()
	var corpus []Action
	left = budget
	exhausted = false
	t0 := time.Now()
	ops = 0
	rg := &Ridge{}
	typeSum := [4]float64{}
	typeN := [4]float64{}
	res := Result{Seed: seed, Arm: arm}
	shuf := arm == "learned_shuffled"
	for !exhausted {
		batch := make([]Task, batchN)
		for i := range batch {
			batch[i] = w.task(sr)
		}
		// pool is a deterministic function of (seed, round, library): identical libraries get identical pools in every arm
		pr := rand.New(rand.NewSource(hashLib(seed, res.Rounds, lib)))
		pool := gen(lib, pr, poolK)
		if len(pool) == 0 {
			break
		}
		var newCorp []Action
		base, _ := evalLib(lib, batch, &newCorp)
		if exhausted {
			break
		}
		rate := float64(base) / batchN
		cand := append(append([]Action{}, corpus...), newCorp...)
		if len(cand) > 300 {
			cand = cand[len(cand)-300:]
		}
		corpusNext := cand
		F := make([][]float64, len(pool))
		for i, e := range pool {
			F[i] = feats(lib, e, corpusNext, rate)
		}
		idx := ar.Perm(len(pool))
		var chosen []int
		switch arm {
		case "collect":
			chosen = idx
		case "random":
			chosen = idx[:pickM]
		case "ucb":
			sc := make([]float64, len(pool))
			tot := 1.0
			for t := 0; t < 4; t++ {
				tot += typeN[t]
			}
			for i, e := range pool {
				if typeN[e.Type] == 0 {
					sc[i] = 1e3
				} else {
					sc[i] = typeSum[e.Type]/typeN[e.Type] + 2.0*math.Sqrt(math.Log(tot)/typeN[e.Type])
				}
				sc[i] += 1e-6 * ar.Float64()
			}
			chosen = topK(sc, pickM)
		case "freqrule":
			sc := make([]float64, len(pool))
			for i, e := range pool {
				sc[i] = F[i][4]
				if e.Type == 2 || F[i][5] == 1 {
					sc[i] = -1
				}
				sc[i] += 1e-6 * ar.Float64()
			}
			chosen = topK(sc, pickM)
		case "transfer", "transfer_shuf":
			sc := make([]float64, len(pool))
			for i := range pool {
				sc[i] = predW(transferW, F[i]) + 1e-6*ar.Float64()
			}
			chosen = topK(sc, pickM-1)
			for _, ix := range idx {
				dup := false
				for _, c := range chosen {
					if c == ix {
						dup = true
					}
				}
				if !dup {
					chosen = append(chosen, ix)
					break
				}
			}
		case "learned", "learned_shuffled":
			if len(rg.X) < 3*pickM {
				chosen = idx[:pickM]
			} else {
				rg.fit()
				sc := make([]float64, len(pool))
				for i := range pool {
					sc[i] = rg.pred(F[i]) + 1e-6*ar.Float64()
				}
				chosen = topK(sc, pickM-1)
				for _, ix := range idx {
					dup := false
					for _, c := range chosen {
						if c == ix {
							dup = true
						}
					}
					if !dup {
						chosen = append(chosen, ix)
						break
					}
				}
			}
		}
		bestD, bestI := 0, -1
		var bestCorp []Action
		var pos []int
		type ev struct {
			ci, d int
			cc    []Action
		}
		var evs []ev
		for _, ci := range chosen {
			nl := lib.apply(pool[ci])
			var cc []Action
			s, _ := evalLib(nl, batch, &cc)
			if exhausted {
				break
			}
			res.Evals++
			evs = append(evs, ev{ci, s - base, cc})
		}
		if exhausted {
			break // truncated round is discarded: no acceptance, no label (its spend still counts)
		}
		for _, e := range evs {
			rg.X = append(rg.X, F[e.ci])
			rg.Y = append(rg.Y, float64(e.d))
			typeSum[pool[e.ci].Type] += float64(e.d)
			typeN[pool[e.ci].Type]++
			if collect != nil {
				*collect = append(*collect, Sample{F[e.ci], float64(e.d)})
			}
			if e.d > bestD {
				bestD, bestI, bestCorp = e.d, e.ci, e.cc
			}
			if e.d > 0 {
				pos = append(pos, e.ci)
			}
		}
		if shuf && len(rg.Y) > 1 {
			ar.Shuffle(len(rg.Y), func(i, j int) { rg.Y[i], rg.Y[j] = rg.Y[j], rg.Y[i] })
		}
		corpus = corpusNext
		if arm == "collect" && len(pos) > 0 { // diverse trajectory: accept a random improving edit
			bestI = pos[ar.Intn(len(pos))]
			for _, e := range evs {
				if e.ci == bestI {
					bestCorp = e.cc
				}
			}
		}
		if bestI >= 0 {
			lib = lib.apply(pool[bestI])
			res.Accepts++
			corpus = append(corpus, bestCorp...)
			if len(corpus) > 300 {
				corpus = corpus[len(corpus)-300:]
			}
		}
		res.Rounds++
	}
	res.WallMs = time.Since(t0).Milliseconds()
	res.Ops = ops
	res.Spent = budget - left
	left = 1 << 60
	exhausted = false
	s, c := evalLib(lib, audit, nil)
	res.AuditSolve, res.AuditCost, res.LibSize = s, c, len(lib.Acts)
	return res
}

func topK(sc []float64, k int) []int {
	idx := make([]int, len(sc))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(a, b int) bool { return sc[idx[a]] > sc[idx[b]] })
	return idx[:k]
}

func fitW(S []Sample, shuffle bool, r *rand.Rand) []float64 {
	rg := &Ridge{}
	ys := make([]float64, len(S))
	for i, s := range S {
		ys[i] = s.Y
		rg.X = append(rg.X, s.X)
	}
	if shuffle {
		r.Shuffle(len(ys), func(i, j int) { ys[i], ys[j] = ys[j], ys[i] })
	}
	rg.Y = ys
	rg.fit()
	return rg.w
}

func main() {
	mode := flag.String("mode", "test", "collect | test")
	seedLo := flag.Int("lo", 1, "first seed")
	seedHi := flag.Int("hi", 4, "last seed")
	budget := flag.Int("budget", 6000000, "hard cap on decision expansions per arm")
	auditN := flag.Int("audit", 300, "audit tasks")
	out := flag.String("out", "res.json", "output")
	trainFile := flag.String("train", "train.json", "collected labelled edits")
	wfile := flag.String("weights", "weights.json", "frozen transfer weights")
	arms := flag.String("arms", "random,ucb,freqrule,learned,transfer,transfer_shuf", "arms")
	flag.Parse()
	if *mode == "collect" {
		var S []Sample
		collect = &S
		tot := 0
		for s := *seedLo; s <= *seedHi; s++ {
			r := runArm(s, "collect", *budget, nil)
			tot += r.Spent
			fmt.Printf("collect seed %d samples %d spent %d\n", s, len(S), r.Spent)
		}
		f, _ := os.Create(*trainFile)
		json.NewEncoder(f).Encode(S)
		f.Close()
		w := fitW(S, false, nil)
		ws := fitW(S, true, rand.New(rand.NewSource(424242)))
		f, _ = os.Create(*wfile)
		json.NewEncoder(f).Encode(map[string]interface{}{"transfer": w, "transfer_shuf": ws, "n_samples": len(S), "train_expansions": tot})
		f.Close()
		fmt.Println("weights", w)
		return
	}
	var wm map[string]interface{}
	if b, err := os.ReadFile(*wfile); err == nil {
		json.Unmarshal(b, &wm)
	}
	getW := func(k string) []float64 {
		var o []float64
		if a, ok := wm[k].([]interface{}); ok {
			for _, v := range a {
				o = append(o, v.(float64))
			}
		}
		return o
	}
	var all []Result
	for s := *seedLo; s <= *seedHi; s++ {
		wr := rand.New(rand.NewSource(int64(s)*1000 + 1))
		w := newWorld(wr)
		aur := rand.New(rand.NewSource(int64(s)*1000 + 555))
		audit := make([]Task, *auditN)
		for i := range audit {
			audit[i] = w.task(aur)
		}
		b0, _ := evalLib(baseLib(), audit, nil)
		for _, a := range strings.Split(*arms, ",") {
			transferW = getW(a)
			r := runArm(s, a, *budget, audit)
			all = append(all, r)
			fmt.Printf("seed %d %-15s audit %3d/%d (base %d) rounds %2d evals %3d acc %2d lib %2d spent %d wallms %d ops %d\n", s, a, r.AuditSolve, *auditN, b0, r.Rounds, r.Evals, r.Accepts, r.LibSize, r.Spent, r.WallMs, r.Ops)
		}
	}
	f, _ := os.Create(*out)
	json.NewEncoder(f).Encode(all)
	f.Close()
}
