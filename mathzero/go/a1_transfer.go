package main

// d1a1transfer: READ-ONLY instrumentation on top of the frozen A1C-B code (no training, never writes into the seed
// directory, never touches weights). Lists:
//   seen = problems ATTEMPTED by the training generation of that seed (same RNG, same G3 exclusion as TrainRoundsA1).
//          We do not claim each one entered the SGD buffer (no per-problem buffer record exists).
//   new  = pairs from the same generator and levels with another RNG stream, rejecting only G3 component keys (the same
//          rule the training generation uses) and any pair already in the seen list (pairKey). Components may overlap
//          seen components: "new" means untrained PAIR, not unseen components.
// Both lists are de-duplicated by pairKey (first occurrence kept); new has, per level, as many cases as seen has after
// de-duplication. Paired by generation rule: this does not guarantee equal real difficulty.
// Nets: pre (NewMLP(seed)), A and F (stored final weights), evaluated by EvalBank (same solver, mode and cap as G3).

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

type TransferList struct {
	Cases    []EvalCase
	Levels   []int
	Pairs    map[string]bool
	Drawn    map[int]int // per level: problems drawn
	Dups     map[int]int // per level: duplicate pairs dropped
	Rejected map[int]int // per level: draws rejected by the exclusion rule
	LenSum   map[int]int // per level: sum of len(RenderPrefix(a))+len(RenderPrefix(b)) over kept cases
	Kept     map[int]int
}

func newTL() TransferList {
	return TransferList{Pairs: map[string]bool{}, Drawn: map[int]int{}, Dups: map[int]int{}, Rejected: map[int]int{}, LenSum: map[int]int{}, Kept: map[int]int{}}
}

func (tl *TransferList) add(id string, level int, a, b *Expr) bool {
	pk := pairKey(a, b)
	if tl.Pairs[pk] {
		tl.Dups[level]++
		return false
	}
	tl.Pairs[pk] = true
	tl.Cases = append(tl.Cases, EvalCase{ID: id, Start: a, Target: b})
	tl.Levels = append(tl.Levels, level)
	tl.Kept[level]++
	tl.LenSum[level] += len(RenderPrefix(a)) + len(RenderPrefix(b))
	return true
}

// transferSeen reproduces the problem loop of TrainRoundsA1 (rp stream depends only on seed and exclusion keys).
// stored (optional) receives the full undeduplicated rendered list for comparison with the stored training log.
func transferSeen(seed int64, sched []int, per int, exclude map[string]bool) (TransferList, []string, error) {
	rp := rand.New(rand.NewSource(seed))
	tl := newTL()
	var rendered []string
	for ri, level := range sched {
		for i := 0; i < per; i++ {
			var a, b *Expr
			ok := false
			for attempt := 0; attempt < 1000; attempt++ {
				a, b = genTrain(rp, level)
				if exclude[Key(a)] || exclude[Key(b)] {
					tl.Rejected[level]++
					continue
				}
				ok = true
				break
			}
			if !ok {
				return tl, nil, fmt.Errorf("no problem outside exclusion set at level %d", level)
			}
			tl.Drawn[level]++
			rendered = append(rendered, RenderPrefix(a)+"|"+RenderPrefix(b))
			tl.add(fmt.Sprintf("seen-r%d-L%d-%d", ri+1, level, i), level, a, b)
		}
	}
	return tl, rendered, nil
}

// transferNew draws fresh PAIRS per level until it has want[level] distinct cases; rejects G3 component keys and seen pairs.
func transferNew(newSeed int64, sched []int, want map[int]int, g3 map[string]bool, seenPairs map[string]bool) (TransferList, error) {
	rp := rand.New(rand.NewSource(newSeed))
	tl := newTL()
	for ri, level := range sched {
		for tl.Kept[level] < want[level] {
			var a, b *Expr
			ok := false
			for attempt := 0; attempt < 5000; attempt++ {
				a, b = genTrain(rp, level)
				tl.Drawn[level]++
				pk := pairKey(a, b)
				if g3[Key(a)] || g3[Key(b)] || seenPairs[pk] {
					tl.Rejected[level]++
					continue
				}
				if tl.Pairs[pk] {
					tl.Dups[level]++
					continue
				}
				ok = true
				break
			}
			if !ok {
				return tl, fmt.Errorf("no new pair at level %d", level)
			}
			tl.add(fmt.Sprintf("new-r%d-L%d-%d", ri+1, level, tl.Kept[level]), level, a, b)
		}
	}
	return tl, nil
}

type ListStats struct {
	Drawn    map[int]int     `json:"drawn_per_level"`
	Kept     map[int]int     `json:"kept_per_level"`
	Dups     map[int]int     `json:"duplicate_pairs_dropped_per_level"`
	Rejected map[int]int     `json:"rejected_draws_per_level"`
	MeanLen  map[int]float64 `json:"mean_rendered_length_per_level"`
}

func statsOf(tl TransferList) ListStats {
	s := ListStats{Drawn: tl.Drawn, Kept: tl.Kept, Dups: tl.Dups, Rejected: tl.Rejected, MeanLen: map[int]float64{}}
	for l, n := range tl.Kept {
		if n > 0 {
			s.MeanLen[l] = float64(tl.LenSum[l]) / float64(n)
		}
	}
	return s
}

type TransferOut struct {
	Seed         int64                       `json:"seed"`
	NewSeed      int64                       `json:"new_seed"`
	Net          string                      `json:"net,omitempty"`
	Levels       []int                       `json:"levels_evaluated"`
	Cap          int                         `json:"cap"`
	SeenStats    ListStats                   `json:"seen_stats"`
	NewStats     ListStats                   `json:"new_stats"`
	NCases       map[string]int              `json:"cases_per_list_pooled"`
	ListSHA      map[string]string           `json:"list_sha256"`
	SeenMatchesF *bool                       `json:"seen_equals_stored_train_F_list,omitempty"`
	Solved       map[string]map[string][]int `json:"solved_per_level,omitempty"` // net -> list -> counts by level
	WeightsSHA   map[string]string           `json:"weights_sha256,omitempty"`
	SeenMatches  *bool                       `json:"seen_equals_stored_training_list,omitempty"`
	Note         string                      `json:"note"`
}

func listSHA(cs []EvalCase) string {
	h := sha256.New()
	for _, c := range cs {
		h.Write([]byte(pairKey(c.Start, c.Target)))
		h.Write([]byte{'\n'})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func transferNewSeed(seed int64) int64 { return seed*1000033 + 424242 }

func transferG3Keys() (map[string]bool, error) {
	prior, err := G3PriorKeys()
	if err != nil {
		return nil, err
	}
	bank, _, err := GenBankG1(G1Rule{Seed: G3Seed, PerLevel: 30, Levels: []int{6, 7, 8}}, prior)
	if err != nil {
		return nil, err
	}
	if d := BankDigest(bank); d != G3FrozenDigest {
		return nil, fmt.Errorf("G3 digest mismatch %s", d)
	}
	return G1KeySetOf(bank), nil
}

func transferLists(seed int64, per int) (TransferList, TransferList, []string, error) {
	g3, err := transferG3Keys()
	if err != nil {
		return TransferList{}, TransferList{}, nil, err
	}
	seen, rendered, err := transferSeen(seed, A1Schedule, per, g3)
	if err != nil {
		return seen, TransferList{}, nil, err
	}
	nw, err := transferNew(transferNewSeed(seed), A1Schedule, seen.Kept, g3, seen.Pairs)
	return seen, nw, rendered, err
}

func RunA1Transfer(dir string, seed int64, per, capE int, levels []int, net string, genOnly bool, w *os.File) error {
	seen, nw, rendered, err := transferLists(seed, per)
	if err != nil {
		return err
	}
	out := TransferOut{Seed: seed, NewSeed: transferNewSeed(seed), Net: net, Levels: levels, Cap: capE, SeenStats: statsOf(seen), NewStats: statsOf(nw), NCases: map[string]int{},
		Note: "read-only; seen = problems attempted in training generation (not proven to be in the SGD buffer); new = untrained PAIR, components may overlap seen; paired by generation rule, equal real difficulty not guaranteed; same solver/mode/cap as G3 evaluation; measures package guide+data, not pure guide effect; exploratory, n=8"}
	want := map[int]bool{}
	for _, l := range levels {
		want[l] = true
	}
	sub := func(tl TransferList) ([]EvalCase, []int) {
		var cs []EvalCase
		var ls []int
		for i, c := range tl.Cases {
			if want[tl.Levels[i]] {
				cs, ls = append(cs, c), append(ls, tl.Levels[i])
			}
		}
		return cs, ls
	}
	sc, _ := sub(seen)
	nc, _ := sub(nw)
	out.NCases["seen"], out.NCases["new"] = len(sc), len(nc)
	out.ListSHA = map[string]string{"seen": listSHA(sc), "new": listSHA(nc)}
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	if genOnly {
		return enc.Encode(out)
	}
	// integrity: the regenerated seen list must equal the stored training list of BOTH arms (arm-independent by design)
	for _, arm := range []string{"A", "F"} {
		b, e := os.ReadFile(filepath.Join(dir, "train-"+arm+".json"))
		if e != nil {
			return fmt.Errorf("cannot read train-%s.json: %v", arm, e)
		}
		var sf struct {
			Train struct{ RenderedProblems []string } `json:"train_log"`
		}
		if err := json.Unmarshal(b, &sf); err != nil || len(sf.Train.RenderedProblems) != len(rendered) {
			return fmt.Errorf("stored training list of arm %s unreadable or wrong length", arm)
		}
		for i := range rendered {
			if sf.Train.RenderedProblems[i] != rendered[i] {
				return fmt.Errorf("regenerated seen list differs from stored list of arm %s at %d", arm, i)
			}
		}
	}
	t := true
	out.SeenMatches = &t
	out.SeenMatchesF = &t
	var n *MLP
	switch net {
	case "pre":
		n = NewMLP(seed)
	case "A", "F":
		p := filepath.Join(dir, "weights_"+net+".json")
		if _, e := os.Stat(p); e != nil {
			return fmt.Errorf("missing %s", p)
		}
		var st struct {
			Weights string `json:"weights_sha256"`
		}
		sb, _ := os.ReadFile(filepath.Join(dir, "train-"+net+".json"))
		json.Unmarshal(sb, &st)
		n = LoadMLP(p)
		if st.Weights == "" || WeightsSHA(n) != st.Weights {
			return fmt.Errorf("stored weights of arm %s do not match recorded sha", net)
		}
	default:
		return fmt.Errorf("-net must be pre, A or F")
	}
	out.WeightsSHA = map[string]string{net: WeightsSHA(n)}
	out.Solved = map[string]map[string][]int{net: {}}
	for lname, pair := range map[string][2]interface{}{"seen": {sc, seen}, "new": {nc, nw}} {
		cs := pair[0].([]EvalCase)
		tl := pair[1].(TransferList)
		var ls []int
		for i := range tl.Cases {
			if want[tl.Levels[i]] {
				ls = append(ls, tl.Levels[i])
			}
		}
		rs := EvalBank(n, cs, EvalConfig{Cap: capE})
		cnt := make([]int, len(levels))
		for i, r := range rs {
			if r.Solved {
				for j, l := range levels {
					if l == ls[i] {
						cnt[j]++
					}
				}
			}
		}
		out.Solved[net][lname] = cnt
	}
	return enc.Encode(out)
}

func init() {
	if len(os.Args) > 1 && os.Args[1] == "d1a1transfer" {
		fs := flag.NewFlagSet("d1a1transfer", flag.ExitOnError)
		dir := fs.String("dir", "", "finished seed directory (read only)")
		seed := fs.Int64("seed", 0, "seed")
		per := fs.Int("perround", 40, "problems per round")
		capE := fs.Int("evalcap", 100, "node cap, same as G3 evaluation")
		lv := fs.String("levels", "6,7", "levels pooled")
		net := fs.String("net", "", "pre | A | F")
		genOnly := fs.Bool("genonly", false, "list statistics only, no networks, no solver")
		outP := fs.String("out", "", "write JSON here atomically (tmp+rename); default stdout")
		fs.Parse(os.Args[2:])
		var levels []int
		for _, s := range strings.Split(*lv, ",") {
			var v int
			fmt.Sscan(s, &v)
			levels = append(levels, v)
		}
		if *seed == 0 || (!*genOnly && (*dir == "" || *net == "")) {
			fmt.Fprintln(os.Stderr, "need -seed and (-dir and -net, or -genonly)")
			os.Exit(2)
		}
		w := os.Stdout
		tmp := ""
		if *outP != "" {
			tmp = *outP + ".tmp"
			f, err := os.Create(tmp)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			w = f
		}
		if err := RunA1Transfer(*dir, *seed, *per, *capE, levels, *net, *genOnly, w); err != nil {
			fmt.Fprintln(os.Stderr, "d1a1transfer:", err)
			os.Exit(1)
		}
		if tmp != "" {
			w.Close()
			if err := os.Rename(tmp, *outP); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		os.Exit(0)
	}
}
