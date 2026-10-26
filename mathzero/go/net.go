// Red pequena en Go puro: puntua pares (estado, accion) para guiar la
// busqueda. MLP 24->48->1 con backprop y Adam escritos a mano.
// Se entrena por autojuego (expert iteration): la busqueda guiada por
// la red encuentra caminos; la red aprende a imitar las acciones que
// llevaron a solucion; y la busqueda con mejor red resuelve mas.
package main

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
)

const NFeat = 24
const NHid = 48

type MLP struct {
	W1    [][]float64 `json:"w1"`
	B1    []float64   `json:"b1"`
	W2    []float64   `json:"w2"`
	B2    float64     `json:"b2"`
	Level int         `json:"level"`
	Round int         `json:"round"`
	// Adam
	m1          [][]float64
	v1          [][]float64
	mb          []float64
	vb          []float64
	m2          []float64
	v2          []float64
	mb2, vb2, t float64
}

func NewMLP(seed int64) *MLP {
	r := rand.New(rand.NewSource(seed))
	n := &MLP{W1: make([][]float64, NHid), B1: make([]float64, NHid), W2: make([]float64, NHid),
		m1: make([][]float64, NHid), v1: make([][]float64, NHid), mb: make([]float64, NHid),
		vb: make([]float64, NHid), m2: make([]float64, NHid), v2: make([]float64, NHid), Level: 2}
	for i := range n.W1 {
		n.W1[i] = make([]float64, NFeat)
		n.m1[i] = make([]float64, NFeat)
		n.v1[i] = make([]float64, NFeat)
		for j := range n.W1[i] {
			n.W1[i][j] = r.NormFloat64() * math.Sqrt(2.0/NFeat)
		}
		n.W2[i] = r.NormFloat64() * math.Sqrt(2.0/NHid)
	}
	return n
}

// forward cache-less
func (n *MLP) hidden(x []float64) []float64 {
	h := make([]float64, NHid)
	for i := 0; i < NHid; i++ {
		s := n.B1[i]
		for j := 0; j < NFeat; j++ {
			s += n.W1[i][j] * x[j]
		}
		h[i] = math.Tanh(s)
	}
	return h
}

func (n *MLP) ScoreClean(x []float64) float64 {
	h := n.hidden(x)
	s := n.B2
	for i := 0; i < NHid; i++ {
		s += n.W2[i] * h[i]
	}
	return s
}

// TrainCE: entrenamiento por entropia cruzada sobre el conjunto de
// acciones legales de un estado: feat[i] = features de (estado, accion_i),
// target = indice de la accion que llevo a la solucion.
func (n *MLP) TrainCE(feats [][]float64, target int, lr float64) float64 {
	k := len(feats)
	hidden := make([][]float64, k)
	scores := make([]float64, k)
	maxs := -1e18
	for i, f := range feats {
		hidden[i] = n.hidden(f)
		s := n.B2
		for j := 0; j < NHid; j++ {
			s += n.W2[j] * hidden[i][j]
		}
		scores[i] = s
		if s > maxs {
			maxs = s
		}
	}
	sum := 0.0
	p := make([]float64, k)
	for i := range scores {
		p[i] = math.Exp(scores[i] - maxs)
		sum += p[i]
	}
	for i := range p {
		p[i] /= sum
	}
	// gradientes
	b1, b2, e1, e2 := 0.9, 0.999, 1e-8, 1e-4
	n.t++
	for i, f := range feats {
		d := p[i]
		if i == target {
			d -= 1.0
		}
		for j := 0; j < NHid; j++ {
			g2 := d * hidden[i][j]
			n.m2[j] = b1*n.m2[j] + (1-b1)*g2
			n.v2[j] = b2*n.v2[j] + (1-b2)*g2*g2
			n.W2[j] -= lr * (n.m2[j] / (1 - math.Pow(b1, n.t))) / (math.Sqrt(n.v2[j]/(1-math.Pow(b2, n.t))) + e1)
			dh := d * n.W2[j] * (1 - hidden[i][j]*hidden[i][j])
			for l := 0; l < NFeat; l++ {
				g1 := dh*f[l] + e2*n.W1[j][l]
				n.m1[j][l] = b1*n.m1[j][l] + (1-b1)*g1
				n.v1[j][l] = b2*n.v1[j][l] + (1-b2)*g1*g1
				n.W1[j][l] -= lr * (n.m1[j][l] / (1 - math.Pow(b1, n.t))) / (math.Sqrt(n.v1[j][l]/(1-math.Pow(b2, n.t))) + e1)
			}
		}
		gb := d
		n.mb2 = b1*n.mb2 + (1-b1)*gb
		n.vb2 = b2*n.vb2 + (1-b2)*gb*gb
		n.B2 -= lr * (n.mb2 / (1 - math.Pow(b1, n.t))) / (math.Sqrt(n.vb2/(1-math.Pow(b2, n.t))) + e1)
	}
	return -math.Log(p[target] + 1e-12)
}

func (n *MLP) Save(path string) {
	f, _ := os.Create(path)
	defer f.Close()
	json.NewEncoder(f).Encode(n)
}

func LoadMLP(path string) *MLP {
	data := readDataFile(path)
	if data == nil {
		return nil
	}
	n := &MLP{}
	if json.Unmarshal(data, n) != nil {
		return nil
	}
	n.m1 = make([][]float64, NHid)
	n.v1 = make([][]float64, NHid)
	n.mb = make([]float64, NHid)
	n.vb = make([]float64, NHid)
	n.m2 = make([]float64, NHid)
	n.v2 = make([]float64, NHid)
	for i := range n.m1 {
		n.m1[i] = make([]float64, NFeat)
		n.v1[i] = make([]float64, NFeat)
	}
	return n
}
