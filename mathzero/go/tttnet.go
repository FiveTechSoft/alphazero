// Red del tres en raya compatible con la entrenada en Python
// (alphazero_prototipo/model.py): 27 -> 128 (ReLU) -> politica 9
// (softmax sobre legales) + valor (tanh). Mismos pesos exportados a
// JSON: es la MISMA red medida contra minimax, no una reescritura.
package main

import (
	"encoding/json"
	"math"
)

type TTTNet struct {
	W1 [][]float64 `json:"w1"` // 27 x 128
	B1 []float64   `json:"b1"`
	Wp [][]float64 `json:"wp"` // 128 x 9
	Bp []float64   `json:"bp"`
	Wv []float64   `json:"wv"` // 128
	Bv float64     `json:"bv"`
}

func LoadTTTNet(data []byte) *TTTNet {
	n := &TTTNet{}
	if json.Unmarshal(data, n) != nil || !n.valid() {
		return nil
	}
	return n
}

func (n *TTTNet) valid() bool {
	if n == nil || len(n.W1) != 27 || len(n.B1) != 128 ||
		len(n.Wp) != 128 || len(n.Bp) != 9 || len(n.Wv) != 128 {
		return false
	}
	for _, row := range n.W1 {
		if len(row) != 128 {
			return false
		}
	}
	for _, row := range n.Wp {
		if len(row) != 9 {
			return false
		}
	}
	return true
}

func (n *TTTNet) Predict(s any, acts []any) ([]float64, float64) {
	st := s.(TTTState)
	// 27 entradas: 9 propias, 9 del rival, 9 unos (plano de turno)
	var x [27]float64
	for i := 0; i < 9; i++ {
		c := st.Board[i] * st.Player
		if c == 1 {
			x[i] = 1
		}
		if c == -1 {
			x[9+i] = 1
		}
		x[18+i] = 1
	}
	h := make([]float64, len(n.B1))
	for j := range h {
		sv := n.B1[j]
		for i := 0; i < 27; i++ {
			sv += x[i] * n.W1[i][j]
		}
		if sv < 0 {
			sv = 0
		}
		h[j] = sv
	}
	var logits [9]float64
	for a := 0; a < 9; a++ {
		sv := n.Bp[a]
		for j := range h {
			sv += h[j] * n.Wp[j][a]
		}
		logits[a] = sv
	}
	mx := -1e18
	sc := make([]float64, len(acts))
	for i, a := range acts {
		sc[i] = logits[a.(int)]
		if sc[i] > mx {
			mx = sc[i]
		}
	}
	z := 0.0
	for i := range sc {
		sc[i] = math.Exp(sc[i] - mx)
		z += sc[i]
	}
	for i := range sc {
		sc[i] /= z
	}
	tv := n.Bv
	for j := range h {
		tv += h[j] * n.Wv[j]
	}
	return sc, math.Tanh(tv)
}
