//go:build !js

package main

// P1 runner (new code, mathzero-next). `mathzero d1eval -level K -cap N
// [-model path | -seed S] [-lib path] [-max M]` prints one JSON document with
// the configuration, input digests, per-case results and a summary.
// It is wired by init() so that the base main.go stays unmodified.

import (
	"crypto/sha256"
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
	if len(os.Args) > 1 && os.Args[1] == "d1eval" {
		if err := RunD1Eval(os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "d1eval:", err)
			os.Exit(2)
		}
		os.Exit(0)
	}
}

type D1EvalReport struct {
	Level    string `json:"level"`
	Cap      int    `json:"cap"`
	Model    string `json:"model"`
	ModelSHA string `json:"model_sha256,omitempty"`
	Seed     int64  `json:"seed"`
	Library  string `json:"library"`
	LibSHA   string `json:"library_sha256,omitempty"`
	BankSHA  string `json:"bank_sha256"`
	BankRule string `json:"bank_rule"`
	// BankDigest is the unambiguous digest: BankDigestAlgo describes it. BankSHA is
	// kept unchanged for historical identifiers (see PROVENANCE-inc7).
	BankDigest     string       `json:"bank_digest"`
	BankDigestAlgo string       `json:"bank_digest_algo"`
	NodesTot       int          `json:"nodes_total"`
	ElapsedMs      int64        `json:"elapsed_ms"`
	Summary        BankSummary  `json:"summary"`
	Results        []CaseResult `json:"results"`
}

func fileSHA(path string) string {
	b := readDataFile(path)
	if b == nil {
		return ""
	}
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

func RunD1Eval(args []string, w io.Writer) error {
	fs := flag.NewFlagSet("d1eval", flag.ContinueOnError)
	level := fs.String("level", "2", "key in levels.json")
	capN := fs.Int("cap", 300, "node budget per case")
	model := fs.String("model", "", "model json (default: fresh MLP from -seed)")
	seed := fs.Int64("seed", 1, "seed for a fresh MLP")
	lib := fs.String("lib", "", "lemma library json (default: none)")
	max := fs.Int("max", 0, "use only the first N cases (0 = all)")
	bankSeed := fs.Int64("bankseed", 0, "if non-zero, use GenBank(seed) instead of levels.json")
	perLevel := fs.Int("perlevel", 10, "cases per level for -bankseed")
	levels := fs.String("levels", "4,6", "comma-separated levels for -bankseed")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *capN < 0 {
		return fmt.Errorf("negative cap")
	}
	var cases []EvalCase
	var bcases []BankCase
	h := sha256.New()
	ruleDesc := "levels.json level " + *level
	if *bankSeed != 0 {
		var lv []int
		for _, f := range strings.Split(*levels, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(f))
			if err != nil {
				return fmt.Errorf("bad -levels: %v", err)
			}
			lv = append(lv, n)
		}
		bank := GenBank(BankRule{Seed: *bankSeed, PerLevel: *perLevel, Levels: lv})
		if len(bank) == 0 {
			return fmt.Errorf("empty generated bank")
		}
		for _, c := range bank {
			cases = append(cases, c.EvalCase)
		}
		if *max > 0 && *max < len(cases) {
			cases = cases[:*max]
		}
		bcases = bank[:len(cases)]
		hh := BankDigest(bcases)
		fmt.Fprint(h, hh)
		ruleDesc = fmt.Sprintf("GenBank seed=%d perlevel=%d levels=%s", *bankSeed, *perLevel, *levels)
	} else {
		probs := LoadProblems("levels.json", *level)
		if len(probs) == 0 {
			return fmt.Errorf("no cases for level %q", *level)
		}
		if *max > 0 && *max < len(probs) {
			probs = probs[:*max]
		}
		for i, p := range probs {
			cases = append(cases, EvalCase{ID: fmt.Sprintf("L%s-%d", *level, i), Start: p.Cur, Target: p.Target})
			fmt.Fprintf(h, "%s\x00%s\n", RenderPrefix(p.Cur), RenderPrefix(p.Target))
			lv, _ := strconv.Atoi(*level)
			bcases = append(bcases, BankCase{cases[len(cases)-1], lv})
		}
	}
	rep := D1EvalReport{Level: *level, Cap: *capN, Seed: *seed, Library: "none", BankSHA: fmt.Sprintf("%x", h.Sum(nil)), BankRule: ruleDesc, BankDigest: BankDigest(bcases), BankDigestAlgo: BankDigestAlgo}
	var net *MLP
	if *model != "" {
		net = LoadMLP(*model)
		rep.Model, rep.ModelSHA = *model, fileSHA(*model)
	} else {
		net = NewMLP(*seed)
		rep.Model = "fresh-mlp"
	}
	cfg := EvalConfig{Cap: *capN}
	if *lib != "" {
		cfg.Library = LoadLibrary(*lib)
		rep.Library, rep.LibSHA = *lib, fileSHA(*lib)
	}
	t0 := time.Now()
	rep.Results = EvalBank(net, cases, cfg)
	rep.ElapsedMs = time.Since(t0).Milliseconds()
	for _, r := range rep.Results {
		rep.NodesTot += r.Nodes
	}
	rep.Summary = SummarizeBank(rep.Results, *capN)
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	return enc.Encode(rep)
}
