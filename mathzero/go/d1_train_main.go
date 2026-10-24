//go:build !js

package main

// P3 runner (new code). `mathzero d1train -seed S [-rounds 6 -perround 40 -cap 100
// -evalcap 100 -evallevels 6,7,8 -secseed 9001 -secperlevel 30]` prints one JSON
// report: frozen configuration, bank digests, weights hashes pre/post, per-case
// pre/post results on a PRIMARY (levels.json) and a SECONDARY (GenBank) bank, the
// training log and train/eval overlaps. Progress lines go to stderr. It never
// selects checkpoints and the evaluation bank is only used to exclude pairs.

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "d1train" {
		if err := RunD1Train(os.Args[2:], os.Stdout, os.Stderr); err != nil {
			fmt.Fprintln(os.Stderr, "d1train:", err)
			os.Exit(2)
		}
		os.Exit(0)
	}
}

type D1BankReport struct {
	Name       string         `json:"name"`
	Digest     string         `json:"bank_digest"`
	DigestAlgo string         `json:"bank_digest_algo"`
	Cases      int            `json:"cases"`
	Pre        []CaseResult   `json:"pre"`
	Post       []CaseResult   `json:"post"`
	PreSolved  int            `json:"pre_solved"`
	PostSolved int            `json:"post_solved"`
	Delta      int            `json:"delta"`
	LostSolved int            `json:"solved_to_unsolved"`
	GainSolved int            `json:"unsolved_to_solved"`
	PerLevel   map[int][3]int `json:"per_level_pre_post_n"`
}

type D1TrainReport struct {
	Seed         int64                       `json:"seed"`
	Rounds       int                         `json:"rounds"`
	PerRound     int                         `json:"per_round"`
	TrainCap     int                         `json:"train_cap"`
	EvalCap      int                         `json:"eval_cap"`
	Library      string                      `json:"library"`
	PreWeights   string                      `json:"weights_sha256_pre"`
	PostWeights  string                      `json:"weights_sha256_post"`
	FinalLevel   int                         `json:"final_level"`
	Train        TrainLog                    `json:"train_log"`
	Overlap      map[string]TrainEvalOverlap `json:"train_eval_overlap"`
	Banks        []D1BankReport              `json:"banks"`
	SecondsPhase map[string]float64          `json:"seconds_by_phase"`
}

func parseIntList(s string) ([]int, error) {
	var out []int
	for _, f := range strings.Split(s, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(f))
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func buildPrimary(levels []int) []BankCase {
	var out []BankCase
	for _, k := range levels {
		for i, p := range LoadProblems("levels.json", fmt.Sprint(k)) {
			out = append(out, BankCase{EvalCase{ID: fmt.Sprintf("L%d-%d", k, i), Start: p.Cur, Target: p.Target}, k})
		}
	}
	return out
}

func evalBankReport(name string, bank []BankCase, pre, post []CaseResult) D1BankReport {
	r := D1BankReport{Name: name, Digest: BankDigest(bank), DigestAlgo: BankDigestAlgo, Cases: len(bank), Pre: pre, Post: post, PerLevel: map[int][3]int{}}
	for i := range bank {
		a, b := pre[i].Solved, post[i].Solved
		if a {
			r.PreSolved++
		}
		if b {
			r.PostSolved++
		}
		if a && !b {
			r.LostSolved++
		}
		if !a && b {
			r.GainSolved++
		}
		v := r.PerLevel[bank[i].Level]
		if a {
			v[0]++
		}
		if b {
			v[1]++
		}
		v[2]++
		r.PerLevel[bank[i].Level] = v
	}
	r.Delta = r.PostSolved - r.PreSolved
	return r
}

func RunD1Train(args []string, w, prog io.Writer) error {
	fs := flag.NewFlagSet("d1train", flag.ContinueOnError)
	seed := fs.Int64("seed", 0, "network and training seed (required, non-zero)")
	rounds := fs.Int("rounds", 6, "training rounds")
	per := fs.Int("perround", 40, "problems per round")
	capT := fs.Int("cap", 100, "training search budget")
	capE := fs.Int("evalcap", 100, "evaluation search budget")
	evl := fs.String("evallevels", "6,7,8", "levels of the primary (levels.json) and secondary banks")
	secSeed := fs.Int64("secseed", 9001, "GenBank seed of the secondary bank (0 = none)")
	secPer := fs.Int("secperlevel", 30, "cases per level in the secondary bank")
	stage := fs.String("stage", "all", "all | pre-primary | pre-secondary | train | post-primary | post-secondary | report")
	dir := fs.String("dir", "", "state directory for staged runs (required unless -stage all)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *seed == 0 {
		return fmt.Errorf("-seed is required")
	}
	if *rounds < 1 || *per < 1 || *capT < 2 || *capE < 0 || *secPer < 1 {
		return fmt.Errorf("invalid rounds/perround/cap/evalcap/secperlevel")
	}
	lv, err := parseIntList(*evl)
	if err != nil {
		return fmt.Errorf("bad -evallevels: %v", err)
	}
	primary := buildPrimary(lv)
	if len(primary) == 0 {
		return fmt.Errorf("empty primary bank")
	}
	banks := map[string][]BankCase{"primary": primary}
	names := []string{"primary"}
	if *secSeed != 0 {
		banks["secondary"] = GenBank(BankRule{Seed: *secSeed, PerLevel: *secPer, Levels: lv})

package main

// P3 runner (new code). `mathzero d1train -seed S [-rounds 6 -perround 40 -cap 100
// -evalcap 100 -evallevels 6,7,8 -secseed 9001 -secperlevel 30]` prints one JSON
// report: frozen configuration, bank digests, weights hashes pre/post, per-case
// pre/post results on a PRIMARY (levels.json) and a SECONDARY (GenBank) bank, the
// training log and train/eval overlaps. Progress lines go to stderr. It never
// selects checkpoints and the evaluation bank is only used to exclude pairs.

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "d1train" {
		if err := RunD1Train(os.Args[2:], os.Stdout, os.Stderr); err != nil {
			fmt.Fprintln(os.Stderr, "d1train:", err)
			os.Exit(2)
		}
		os.Exit(0)
	}
}

type D1BankReport struct {
	Name       string         `json:"name"`
	Digest     string         `json:"bank_digest"`
	DigestAlgo string         `json:"bank_digest_algo"`
	Cases      int            `json:"cases"`
	Pre        []CaseResult   `json:"pre"`
	Post       []CaseResult   `json:"post"`
	PreSolved  int            `json:"pre_solved"`
	PostSolved int            `json:"post_solved"`
	Delta      int            `json:"delta"`
	LostSolved int            `json:"solved_to_unsolved"`
	GainSolved int            `json:"unsolved_to_solved"`
	PerLevel   map[int][3]int `json:"per_level_pre_post_n"`
}

type D1TrainReport struct {
	Seed         int64                       `json:"seed"`
	Rounds       int                         `json:"rounds"`
	PerRound     int                         `json:"per_round"`
	TrainCap     int                         `json:"train_cap"`
	EvalCap      int                         `json:"eval_cap"`
	Library      string                      `json:"library"`
	PreWeights   string                      `json:"weights_sha256_pre"`
	PostWeights  string                      `json:"weights_sha256_post"`
	FinalLevel   int                         `json:"final_level"`
	Train        TrainLog                    `json:"train_log"`
	Overlap      map[string]TrainEvalOverlap `json:"train_eval_overlap"`
	Banks        []D1BankReport              `json:"banks"`
	SecondsPhase map[string]float64          `json:"seconds_by_phase"`
}

func parseIntList(s string) ([]int, error) {
	var out []int
	for _, f := range strings.Split(s, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(f))
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func buildPrimary(levels []int) []BankCase {
	var out []BankCase
	for _, k := range levels {
		for i, p := range LoadProblems("levels.json", fmt.Sprint(k)) {
			out = append(out, BankCase{EvalCase{ID: fmt.Sprintf("L%d-%d", k, i), Start: p.Cur, Target: p.Target}, k})
		}
	}
	return out
}

func evalBankReport(name string, bank []BankCase, pre, post []CaseResult) D1BankReport {
	r := D1BankReport{Name: name, Digest: BankDigest(bank), DigestAlgo: BankDigestAlgo, Cases: len(bank), Pre: pre, Post: post, PerLevel: map[int][3]int{}}
	for i := range bank {
		a, b := pre[i].Solved, post[i].Solved
		if a {
			r.PreSolved++
		}
		if b {
			r.PostSolved++
		}
		if a && !b {
			r.LostSolved++
		}
		if !a && b {
			r.GainSolved++
		}
		v := r.PerLevel[bank[i].Level]
		if a {
			v[0]++
		}
		if b {
			v[1]++
		}
		v[2]++
		r.PerLevel[bank[i].Level] = v
	}
	r.Delta = r.PostSolved - r.PreSolved
	return r
}

func RunD1Train(args []string, w, prog io.Writer) error {
	fs := flag.NewFlagSet("d1train", flag.ContinueOnError)
	seed := fs.Int64("seed", 0, "network and training seed (required, non-zero)")
	rounds := fs.Int("rounds", 6, "training rounds")
	per := fs.Int("perround", 40, "problems per round")
	capT := fs.Int("cap", 100, "training search budget")
	capE := fs.Int("evalcap", 100, "evaluation search budget")
	evl := fs.String("evallevels", "6,7,8", "levels of the primary (levels.json) and secondary banks")
	secSeed := fs.Int64("secseed", 9001, "GenBank seed of the secondary bank (0 = none)")
	secPer := fs.Int("secperlevel", 30, "cases per level in the secondary bank")
	stage := fs.String("stage", "all", "all | pre-primary | pre-secondary | train | post-primary | post-secondary | report")
	dir := fs.String("dir", "", "state directory for staged runs (required unless -stage all)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *seed == 0 {
		return fmt.Errorf("-seed is required")
	}
	if *rounds < 1 || *per < 1 || *capT < 2 || *capE < 0 || *secPer < 1 {
		return fmt.Errorf("invalid rounds/perround/cap/evalcap/secperlevel")
	}
	lv, err := parseIntList(*evl)
	if err != nil {
		return fmt.Errorf("bad -evallevels: %v", err)
	}
	primary := buildPrimary(lv)
	if len(primary) == 0 {
		return fmt.Errorf("empty primary bank")
	}
	banks := map[string][]BankCase{"primary": primary}
	names := []string{"primary"}
	if *secSeed != 0 {
		banks["secondary"] = GenBank(BankRule{Seed: *secSeed, PerLevel: *secPer, Levels: lv})
		names = append(names, "secondary")
	}
	if *stage != "all" {
		return runStage(*stage, *dir, stageParams{Seed: *seed, Rounds: *rounds, Per: *per, CapT: *capT, CapE: *capE, Lv: lv, SecSeed: *secSeed, SecPer: *secPer}, banks, names, w, prog)
	}
	var all []EvalCase
	toCases := func(b []BankCase) []EvalCase {
		cs := make([]EvalCase, len(b))
		for i, c := range b {
			cs[i] = c.EvalCase
		}
		return cs
	}
	for _, n := range names {
		all = append(all, toCases(banks[n])...)
	}
	rep := D1TrainReport{Seed: *seed, Rounds: *rounds, PerRound: *per, TrainCap: *capT, EvalCap: *capE, Library: "none",
		Overlap: map[string]TrainEvalOverlap{}, SecondsPhase: map[string]float64{}}
	net := NewMLP(*seed)
	rep.PreWeights = WeightsSHA(net)
	phase := func(name string, f func()) {
		t0 := time.Now()
		f()
		rep.SecondsPhase[name] = time.Since(t0).Seconds()
		fmt.Fprintf(prog, "phase %s done in %.1fs\n", name, rep.SecondsPhase[name])
	}
	pre := map[string][]CaseResult{}
	for _, n := range names {
		n := n
		phase("eval_pre_"+n, func() { pre[n] = EvalBank(net, toCases(banks[n]), EvalConfig{Cap: *capE}) })
	}
	var terr error
	phase("train", func() {
		rep.Train, terr = TrainRounds(net, TrainConfig{Seed: *seed, Rounds: *rounds, PerRound: *per, Cap: *capT,
			Exclude: ExcludeSetFor(all), ExcludeRender: RenderExcludeFor(all)})
	})
	if terr != nil {
		return terr
	}
	rep.PostWeights, rep.FinalLevel = WeightsSHA(net), net.Level
	for _, n := range names {
		n := n
		rep.Overlap[n] = OverlapTrainEval(rep.Train, toCases(banks[n]))
		var post []CaseResult
		phase("eval_post_"+n, func() { post = EvalBank(net, toCases(banks[n]), EvalConfig{Cap: *capE}) })
		rep.Banks = append(rep.Banks, evalBankReport(n, banks[n], pre[n], post))
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	return enc.Encode(rep)
}
