//go:build !js

package main

// `mathzero d1g1train -seed S -stage STAGE -dir D [-rounds 6 -perround 40 -cap 100
// -evalcap 100]` runs ONE stage of the G1 protocol (new; not P3). The only bank is
// the frozen G1 bank (seed 20261024, levels 6,7,8 x 30), regenerated each time and
// CHECKED against the frozen digest; the exclusion set is the Key of every G1 Start
// and Target. Stages: pre-g1, train, post-g1, report. Fails closed on infeasible
// generation or digest mismatch. See pilots/g1/DESIGN-G1-frozen.md.

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// G1FrozenDigest is the digest of the bank generated with the frozen rule
// (recorded in pilots/g1/BANK-G1.json when it was first generated).
const G1FrozenDigest = "8bc4db332ab0e6c240b0b9e788a091f282a695f31b777301d1d1bcc7962b209e"

type G1ReportInfo struct {
	Protocol   string         `json:"protocol"`
	BankSeed   int64          `json:"bank_seed"`
	BankDigest string         `json:"bank_digest"`
	PriorKeys  int            `json:"prior_keys"`
	GenStats   []G1LevelStats `json:"generation_stats"`
	Note       string         `json:"note"`
}

func init() {
	if len(os.Args) > 1 && os.Args[1] == "d1g1train" {
		if err := RunD1G1Train(os.Args[2:], os.Stdout, os.Stderr); err != nil {
			fmt.Fprintln(os.Stderr, "d1g1train:", err)
			os.Exit(2)
		}
		os.Exit(0)
	}
}

func RunD1G1Train(args []string, w, prog io.Writer) error {
	fs := flag.NewFlagSet("d1g1train", flag.ContinueOnError)
	seed := fs.Int64("seed", 0, "network and training seed (required, non-zero)")
	rounds := fs.Int("rounds", 6, "training rounds")
	per := fs.Int("perround", 40, "problems per round")
	capT := fs.Int("cap", 100, "training search budget")
	capE := fs.Int("evalcap", 100, "evaluation search budget")
	stage := fs.String("stage", "", "pre-g1 | train | post-g1 | report")
	dir := fs.String("dir", "", "state directory (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *seed == 0 {
		return fmt.Errorf("-seed is required")
	}
	if *rounds < 1 || *per < 1 || *capT < 2 || *capE < 0 {
		return fmt.Errorf("invalid rounds/perround/cap/evalcap")
	}
	rule := G1Rule{Seed: G1Seed, PerLevel: 30, Levels: []int{6, 7, 8}}
	bank, st, err := GenBankG1(rule, G1PriorKeys())
	if err != nil {
		return err
	}
	return RunG1StageWithBank(*stage, *dir, *seed, *rounds, *per, *capT, *capE, bank, st, G1FrozenDigest, rule, w, prog)
}

// RunG1StageWithBank is the testable core: the caller supplies the bank, the
// digest it must have, and the generation statistics.
func RunG1StageWithBank(stage, dir string, seed int64, rounds, per, capT, capE int, bank []BankCase, st G1GenStats, wantDigest string, rule G1Rule, w, prog io.Writer) error {
	switch stage {
	case "pre-g1", "train", "post-g1", "report":
	default:
		return fmt.Errorf("unknown G1 stage %q", stage)
	}
	if d := BankDigest(bank); d != wantDigest {
		return fmt.Errorf("G1 bank digest %s differs from the frozen %s: refusing to run", d, wantDigest)
	}
	info := &G1ReportInfo{Protocol: "G1 (new protocol, not P3)", BankSeed: rule.Seed, BankDigest: wantDigest, PriorKeys: st.PriorKeys, GenStats: st.PerLevel,
		Note: "Start/Target disjointness is modulo Key (structural associative-commutative normalization), not algebraic equivalence"}
	p := stageParams{Seed: seed, Rounds: rounds, Per: per, CapT: capT, CapE: capE, Lv: rule.Levels, ExcludeKeys: G1KeySetOf(bank), G1Info: info}
	return runStage(stage, dir, p, map[string][]BankCase{"g1": bank}, []string{"g1"}, w, prog)
}
