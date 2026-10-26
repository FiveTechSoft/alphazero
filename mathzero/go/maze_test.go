package main

import "testing"

func TestMazeBFSyMCTSResuelven(t *testing.T) {
	m := NewMaze(7, 1)
	okB, _, stepsB := mazeBFS(m)
	if !okB {
		t.Fatal("BFS no resolvio el laberinto")
	}
	okM, stepsM := mazeMCTS(m, 150)
	if !okM {
		t.Fatalf("MCTS no resolvio el laberinto (BFS pasos=%d)", stepsB)
	}
	if stepsM <= 0 {
		t.Fatalf("MCTS devolvio pasos invalidos: %d", stepsM)
	}
}

func TestMazeNoRevisitaCeldas(t *testing.T) {
	m := NewMaze(5, 1)
	st := m.Initial()
	acts := m.Legal(st)
	if len(acts) == 0 {
		t.Fatal("sin acciones legales desde el inicio")
	}
	prev := st.Pos
	st = m.Play(st, acts[0]).(MazeState)
	for _, a := range m.Legal(st) {
		if a.(int) == prev {
			t.Fatal("Legal permite volver a una celda ya visitada")
		}
	}
}

func TestMazeEsSinglePlayer(t *testing.T) {
	var g Game = NewMaze(5, 1)
	if _, ok := g.(SinglePlayer); !ok {
		t.Fatal("Maze debe implementar SinglePlayer")
	}
}
