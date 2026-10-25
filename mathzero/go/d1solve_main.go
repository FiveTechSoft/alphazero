//go:build !js

// d1solve: resuelve UN problema (start -> target) en notacion prefija y
// muestra el camino. Codigo NUEVO, local, no auditado: no forma parte del
// binario a1bin congelado ni de ninguna campana.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "d1solve" {
		if err := RunD1Solve(os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "d1solve:", err)
			os.Exit(2)
		}
		os.Exit(0)
	}
}

type D1SolveStep struct {
	From string `json:"from"`
	To   string `json:"to"`
	Rule string `json:"rule"`
}

type D1SolveReport struct {
	Start    string        `json:"start"`
	Target   string        `json:"target"`
	Model    string        `json:"model"`
	ModelSHA string        `json:"model_sha256,omitempty"`
	Cap      int           `json:"cap"`
	Solved   bool          `json:"solved"`
	Nodes    int           `json:"nodes"`
	Path     []D1SolveStep `json:"path,omitempty"`
}

func RunD1Solve(args []string, w io.Writer) error {
	fs := flag.NewFlagSet("d1solve", flag.ContinueOnError)
	start := fs.String("start", "", "expresion inicial, notacion prefija")
	target := fs.String("target", "", "expresion objetivo, notacion prefija")
	model := fs.String("model", "", "model json (por defecto MLP nuevo semilla 1)")
	capN := fs.Int("cap", 400, "presupuesto de nodos")
	if err := fs.Parse(args); err != nil {
		return err
	}
	s, err := ParsePrefix(*start)
	if err != nil {
		return fmt.Errorf("start: %v", err)
	}
	t, err := ParsePrefix(*target)
	if err != nil {
		return fmt.Errorf("target: %v", err)
	}
	var net *MLP
	if *model != "" {
		net = LoadMLP(*model)
	} else {
		net = NewMLP(1)
	}
	prevTraining := SearchTraining
	SearchTraining = false
	defer func() { SearchTraining = prevTraining }()
	ok, nodes, path := SolvePath(net, s, t, *capN)
	rep := D1SolveReport{Start: RenderPrefix(s), Target: RenderPrefix(t), Model: *model, Cap: *capN, Solved: ok, Nodes: nodes}
	if *model != "" {
		rep.ModelSHA = fileSHA(*model)
	}
	for _, st := range path {
		rep.Path = append(rep.Path, D1SolveStep{From: RenderPrefix(st.From), To: RenderPrefix(st.To), Rule: st.Rule})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	return enc.Encode(rep)
}
