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
	Level    string       `json:"level"`
	Cap      int          `json:"cap"`
	Model    string       `json:"model"`
	ModelSHA string       `json:"model_sha256,omitempty"`
	Seed     int64        `json:"seed"`
	Library  string       `json:"library"`
	LibSHA   string       `json:"library_sha256,omitempty"`
	BankSHA  string       `json:"bank_sha256"`
	Summary  BankSummary  `json:"summary"`
	Results  []CaseResult `json:"results"`
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
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *capN < 0 {
		return fmt.Errorf("negative cap")
	}
	probs := LoadProblems("levels.json", *level)
	if len(probs) == 0 {
		return fmt.Errorf("no cases for level %q", *level)
	}
	if *max > 0 && *max < len(probs) {
		probs = probs[:*max]
	}
	cases := make([]EvalCase, len(probs))
	h := sha256.New()
	for i, p := range probs {
		cases[i] = EvalCase{ID: fmt.Sprintf("L%s-%d", *level, i), Start: p.Cur, Target: p.Target}
		fmt.Fprintf(h, "%s\x00%s\n", RenderPrefix(p.Cur), RenderPrefix(p.Target))
	}
	rep := D1EvalReport{Level: *level, Cap: *capN, Seed: *seed, Library: "none", BankSHA: fmt.Sprintf("%x", h.Sum(nil))}
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
	rep.Results = EvalBank(net, cases, cfg)
	rep.Summary = SummarizeBank(rep.Results, *capN)
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	return enc.Encode(rep)
}
