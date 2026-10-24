//go:build !js

package main

// `mathzero d1g1gen [-seed 20261024 -perlevel 30 -levels 6,7,8]` prints the G1
// bank (rendered Start/Target), digest, rule and generation statistics as JSON.
// No network is evaluated and nothing is trained. Exit 2 on infeasibility, with
// the statistics still printed.

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "d1g1gen" {
		if err := RunG1Gen(os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "d1g1gen:", err)
			os.Exit(2)
		}
		os.Exit(0)
	}
}

type G1BankCaseJSON struct {
	ID, Start, Target string
	Level             int
}

type G1BankReport struct {
	Rule       G1Rule           `json:"rule"`
	Feasible   bool             `json:"feasible"`
	Error      string           `json:"error,omitempty"`
	Digest     string           `json:"bank_digest"`
	DigestAlgo string           `json:"bank_digest_algo"`
	Cases      []G1BankCaseJSON `json:"cases"`
	Stats      G1GenStats       `json:"stats"`
}

func RunG1Gen(args []string, w io.Writer) error {
	fs := flag.NewFlagSet("d1g1gen", flag.ContinueOnError)
	seed := fs.Int64("seed", G1Seed, "bank seed")
	per := fs.Int("perlevel", 30, "cases per level")
	levels := fs.String("levels", "6,7,8", "levels")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var lv []int
	for _, x := range strings.Split(*levels, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(x))
		if err != nil {
			return err
		}
		lv = append(lv, n)
	}
	rule := G1Rule{Seed: *seed, PerLevel: *per, Levels: lv}
	cs, st, gerr := GenBankG1(rule, G1PriorKeys())
	rep := G1BankReport{Rule: rule, Feasible: gerr == nil, Digest: BankDigest(cs), DigestAlgo: BankDigestAlgo, Stats: st}
	if gerr != nil {
		rep.Error = gerr.Error()
	}
	for _, c := range cs {
		rep.Cases = append(rep.Cases, G1BankCaseJSON{ID: c.ID, Start: RenderPrefix(c.Start), Target: RenderPrefix(c.Target), Level: c.Level})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	if err := enc.Encode(rep); err != nil {
		return err
	}
	return gerr
}
