//go:build !js

package main

// `mathzero d1a1train -seed S -stage STAGE -dir D` runs ONE stage of protocol A1
// (new; not P3/G1). Frozen: bank G2 (seed 20261025, L6-8 x30), schedule 2..7, 6x40,
// train cap 100, eval cap 100. Stages: pre-g2 (shared by both arms), train-A,
// train-F, post-A, post-F, report. Fails closed on infeasible generation or a
// digest that differs from the frozen one. Design: pilots/a1/DESIGN-A1-frozen.md.

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const G2FrozenDigest = "f58161c5133bffeb2838ee9b5da9a516f09022bbe7d94367c0b4c658d4bb67b6"

var A1Schedule = []int{2, 3, 4, 5, 6, 7}

func init() {
	if len(os.Args) > 1 && os.Args[1] == "d1a1train" {
		if err := RunD1A1Train(os.Args[2:], os.Stdout, os.Stderr); err != nil {
			fmt.Fprintln(os.Stderr, "d1a1train:", err)
			os.Exit(2)
		}
		os.Exit(0)
	}
}

type A1Params struct {
	Seed        int64
	Schedule    []int
	Per         int
	CapT, CapE  int
	BankDigest  string
	ExcludeKeys map[string]bool
}

func (p A1Params) fingerprint() string {
	h := sha256.New()
	ks := make([]string, 0, len(p.ExcludeKeys))
	for k := range p.ExcludeKeys {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	fmt.Fprintf(h, "A1 seed=%d sched=%v per=%d capT=%d capE=%d bank=%s keys=%d %x\n", p.Seed, p.Schedule, p.Per, p.CapT, p.CapE, p.BankDigest, len(ks), sha256.Sum256([]byte(strings.Join(ks, "\n"))))
	return fmt.Sprintf("%x", h.Sum(nil))
}

type A1ArmReport struct {
	Arm          string           `json:"arm"`
	WeightsPost  string           `json:"weights_sha256_post"`
	GuideSHAPre  string           `json:"guide_sha_pre,omitempty"`
	GuideSHAPost string           `json:"guide_sha_post,omitempty"`
	Train        TrainLog         `json:"train_log"`
	Overlap      TrainEvalOverlap `json:"train_eval_overlap"`
	Bank         D1BankReport     `json:"bank"`
	SecondsTrain float64          `json:"seconds_train"`
	SecondsPost  float64          `json:"seconds_post"`
}

type A1Report struct {
	Protocol          string         `json:"protocol"`
	Seed              int64          `json:"seed"`
	Schedule          []int          `json:"schedule"`
	PerRound          int            `json:"per_round"`
	TrainCap          int            `json:"train_cap"`
	EvalCap           int            `json:"eval_cap"`
	BankSeed          int64          `json:"bank_seed"`
	BankDigest        string         `json:"bank_digest"`
	PriorKeys         int            `json:"prior_keys"`
	GenStats          []G1LevelStats `json:"generation_stats"`
	PreWeights        string         `json:"weights_sha256_pre"`
	SecondsPre        float64        `json:"seconds_pre"`
	ProblemsIdentical bool           `json:"training_problems_identical_across_arms"`
	DeltaA            int            `json:"delta_A"`
	DeltaF            int            `json:"delta_F"`
	D                 int            `json:"d_delta_A_minus_delta_F"`
	Arms              []A1ArmReport  `json:"arms"`
	Note              string         `json:"note"`
}

func RunD1A1Train(args []string, w, prog io.Writer) error {
	fs := flag.NewFlagSet("d1a1train", flag.ContinueOnError)
	seed := fs.Int64("seed", 0, "network and training seed (required, non-zero)")
	per := fs.Int("perround", 40, "problems per round")
	capT := fs.Int("cap", 100, "training search budget")
	capE := fs.Int("evalcap", 100, "evaluation search budget")
	stage := fs.String("stage", "", "pre-g2 | train-A | train-F | post-A | post-F | report")
	dir := fs.String("dir", "", "state directory (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *seed == 0 {
		return fmt.Errorf("-seed is required")
	}
	if *per < 1 || *capT < 2 || *capE < 0 {
		return fmt.Errorf("invalid perround/cap/evalcap")
	}
	prior, err := G2PriorKeys()
	if err != nil {
		return err
	}
	rule := G1Rule{Seed: G2Seed, PerLevel: 30, Levels: []int{6, 7, 8}}
	bank, st, err := GenBankG1(rule, prior)
	if err != nil {
		return err
	}
	return RunA1StageWithBank(*stage, *dir, *seed, A1Schedule, *per, *capT, *capE, bank, st, G2FrozenDigest, rule, w, prog)
}

func RunA1StageWithBank(stage, dir string, seed int64, sched []int, per, capT, capE int, bank []BankCase, st G1GenStats, wantDigest string, rule G1Rule, w, prog io.Writer) error {
	switch stage {
	case "pre-g2", "train-A", "train-F", "post-A", "post-F", "report":
	default:
		return fmt.Errorf("unknown A1 stage %q", stage)
	}
	if dir == "" {
		return fmt.Errorf("-dir is required")
	}
	if d := BankDigest(bank); d != wantDigest {
		return fmt.Errorf("bank digest %s differs from the frozen %s: refusing to run", d, wantDigest)
	}
	p := A1Params{Seed: seed, Schedule: sched, Per: per, CapT: capT, CapE: capE, BankDigest: wantDigest, ExcludeKeys: G1KeySetOf(bank)}
	fp := p.fingerprint()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	cases := make([]EvalCase, len(bank))
	for i, c := range bank {
		cases[i] = c.EvalCase
	}
	timed := func(f func()) float64 { t0 := time.Now(); f(); return time.Since(t0).Seconds() }
	done := func(name string, sf stageFile) error {
		sf.Fingerprint, sf.Stage = fp, name
		if _, err := os.Stat(filepath.Join(dir, name+".json")); err == nil {
			return fmt.Errorf("stage %s already exists in %s; refusing to overwrite", name, dir)
		}
		if err := writeJSON(filepath.Join(dir, name+".json"), sf); err != nil {
			return err
		}
		fmt.Fprintf(prog, "stage %s written (%.1fs)\n", name, sf.Seconds)
		return nil
	}
	switch {
	case stage == "pre-g2":
		net := NewMLP(seed)
		var rs []CaseResult
		sec := timed(func() { rs = EvalBank(net, cases, EvalConfig{Cap: capE}) })
		return done(stage, stageFile{Seconds: sec, Results: rs, Weights: WeightsSHA(net)})
	case strings.HasPrefix(stage, "train-"):
		arm := stage[6:]
		wp := filepath.Join(dir, "weights_"+arm+".json")
		if _, err := os.Stat(wp); err == nil {
			return fmt.Errorf("weights for arm %s already exist in %s; refusing to overwrite", arm, dir)
		}
		net := NewMLP(seed)
		var res A1Result
		var terr error
		sec := timed(func() {
			res, terr = TrainRoundsA1(net, A1Config{Seed: seed, Schedule: sched, PerRound: per, Cap: capT, FrozenGuide: arm == "F", ExcludeKeys: p.ExcludeKeys})
		})
		if terr != nil {
			return terr
		}
		if arm == "F" && res.GuideSHAPre != res.GuideSHAPost {
			return fmt.Errorf("the frozen guide changed during training")
		}
		net.Save(wp)
		return done(stage, stageFile{Seconds: sec, Train: &res.Log, Weights: WeightsSHA(net), Level: sched[len(sched)-1], GuideSHAPre: res.GuideSHAPre, GuideSHAPost: res.GuideSHAPost})
	case strings.HasPrefix(stage, "post-"):
		arm := stage[5:]
		tr, err := readStage(dir, "train-"+arm, fp)
		if err != nil {
			return err
		}
		net := LoadMLP(filepath.Join(dir, "weights_"+arm+".json"))
		if net == nil || WeightsSHA(net) != tr.Weights {
			return fmt.Errorf("stored weights of arm %s do not match the hash recorded by its train stage", arm)
		}
		var rs []CaseResult
		sec := timed(func() { rs = EvalBank(net, cases, EvalConfig{Cap: capE}) })
		if WeightsSHA(net) != tr.Weights {
			return fmt.Errorf("evaluation mutated the weights")
		}
		return done(stage, stageFile{Seconds: sec, Results: rs, Weights: WeightsSHA(net)})
	case stage == "report":
		pre, err := readStage(dir, "pre-g2", fp)
		if err != nil {
			return err
		}
		rep := A1Report{Protocol: "A1 (new protocol; not P3/G1)", Seed: seed, Schedule: sched, PerRound: per, TrainCap: capT, EvalCap: capE,
			BankSeed: rule.Seed, BankDigest: wantDigest, PriorKeys: st.PriorKeys, GenStats: st.PerLevel, PreWeights: pre.Weights, SecondsPre: pre.Seconds,
			Note: "total effect of an adaptive guide vs a frozen initial guide; includes mediators (solved, samples, hindsight); not autojuego-vs-SGD; disjointness modulo Key"}
		var problems [][]string
		deltas := map[string]int{}
		for _, arm := range []string{"A", "F"} {
			tr, err := readStage(dir, "train-"+arm, fp)
			if err != nil {
				return err
			}
			po, err := readStage(dir, "post-"+arm, fp)
			if err != nil {
				return err
			}
			br := evalBankReport("g2-"+arm, bank, pre.Results, po.Results)
			deltas[arm] = br.Delta
			problems = append(problems, tr.Train.RenderedProblems)
			rep.Arms = append(rep.Arms, A1ArmReport{Arm: arm, WeightsPost: tr.Weights, GuideSHAPre: tr.GuideSHAPre, GuideSHAPost: tr.GuideSHAPost, Train: *tr.Train,
				Overlap: OverlapTrainEval(*tr.Train, cases), Bank: br, SecondsTrain: tr.Seconds, SecondsPost: po.Seconds})
		}
		rep.ProblemsIdentical = strings.Join(problems[0], "\n") == strings.Join(problems[1], "\n")
		rep.DeltaA, rep.DeltaF = deltas["A"], deltas["F"]
		rep.D = rep.DeltaA - rep.DeltaF
		enc := json.NewEncoder(w)
		enc.SetIndent("", " ")
		return enc.Encode(rep)
	}
	return fmt.Errorf("unreachable stage %q", stage)
}
