package main

// P1 (new code, mathzero-next). Deterministic evaluation of a fixed case bank
// under an explicit configuration. It does not train and does not mutate the
// library. It is NEW code: no historical campaign result validates it.

import "encoding/json"

// EvalCase is one (Start, Target) task, identified by position and content.
type EvalCase struct {
	ID     string
	Start  *Expr
	Target *Expr
}

// EvalConfig fixes everything that can change a search outcome.
type EvalConfig struct {
	Cap     int      // node budget per case
	Library *Library // nil = no library
}

// CaseResult is the outcome of one case.
type CaseResult struct {
	ID     string `json:"id"`
	Solved bool   `json:"solved"`
	Nodes  int    `json:"nodes"`
}

// EvalBank runs every case with the given net and configuration in evaluation
// mode (SearchTraining=false). Global search state is restored on return.
func EvalBank(net *MLP, cases []EvalCase, cfg EvalConfig) []CaseResult {
	prevTraining, prevLib := SearchTraining, LemmaLib
	SearchTraining, LemmaLib = false, cfg.Library
	defer func() { SearchTraining, LemmaLib = prevTraining, prevLib }()
	out := make([]CaseResult, 0, len(cases))
	for _, c := range cases {
		ok, nodes, _ := GuidedSolve(net, c.Start, c.Target, cfg.Cap)
		out = append(out, CaseResult{ID: c.ID, Solved: ok, Nodes: nodes})
	}
	return out
}

// BankSummary counts solved cases and cases that used the whole budget.
type BankSummary struct {
	N, Solved, BudgetExhausted int
}

func SummarizeBank(rs []CaseResult, cap int) BankSummary {
	s := BankSummary{N: len(rs)}
	for _, r := range rs {
		if r.Solved {
			s.Solved++
		} else if r.Nodes >= cap {
			s.BudgetExhausted++
		}
	}
	return s
}

func BankJSON(rs []CaseResult) []byte {
	b, _ := json.Marshal(rs)
	return b
}
