"""Juega tu contra el agente entrenado."""
import numpy as np
from game import TicTacToe
from mcts import mcts_probs
from model import Net

net = Net.load('model.npz')
st = TicTacToe()
human = 1
print("Tres en raya contra AlphaZero-mini. Tu eres X. Casillas:")
print("0 1 2\n3 4 5\n6 7 8\n")
while st.terminal_value() is None:
    print(st.render(), "\n")
    if st.player == human:
        legal = st.legal_moves()
        while True:
            try:
                a = int(input("Tu jugada (0-8): "))
                assert a in legal
                break
            except (ValueError, AssertionError):
                print("jugada ilegal, prueba otra")
    else:
        pi = mcts_probs(st, net, sims=50, temp=0.0, add_noise=False)
        a = int(pi.argmax())
        print(f"AlphaZero juega {a}")
    st.play(a)
print(st.render())
w = st.winner()
print("Tablas" if w == 0 else ("Ganas tu" if w == human else "Gana AlphaZero"))
