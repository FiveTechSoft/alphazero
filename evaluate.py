"""Evaluacion honesta: el agente (red + MCTS, sin ruido) contra

  - random:  jugador que mueve al azar
  - minimax: jugador perfecto (el tres en raya esta resuelto: el juego
             optimo es tablas, asi que la mejor marca posible es
             "nunca pierde")

Se juega como primer y como segundo jugador, porque mover primero es
ventaja y no queremos maquillar el resultado.
"""
import argparse
import functools

import numpy as np

from game import TicTacToe
from mcts import mcts_probs
from model import Net


@functools.lru_cache(maxsize=None)
def _negamax(board_bytes, player):
    st = TicTacToe(np.frombuffer(board_bytes, dtype=np.int8), player)
    tv = st.terminal_value()
    if tv is not None:
        return tv, ()
    best_v, best_moves = -2, ()
    for a in st.legal_moves():
        child = st.clone(); child.play(a)
        v = -_negamax(child.board.tobytes(), child.player)[0]
        if v > best_v:
            best_v, best_moves = v, (int(a),)
        elif v == best_v:
            best_moves = best_moves + (int(a),)
    return best_v, best_moves

def minimax_move(state):
    _, moves = _negamax(state.board.tobytes(), state.player)
    return int(np.random.choice(moves))

def random_move(state):
    return int(np.random.choice(state.legal_moves()))

def agent_move(state, net, sims):
    pi = mcts_probs(state, net, sims=sims, temp=0.0, add_noise=False)
    return int(pi.argmax())

def play_one(agent_first, net, sims, opponent):
    state = TicTacToe()
    opp = minimax_move if opponent == 'minimax' else random_move
    while state.terminal_value() is None:
        is_agent_turn = (state.player == 1) == agent_first
        a = agent_move(state, net, sims) if is_agent_turn else opp(state)
        state.play(a)
    w = state.winner()
    if w == 0: return 'D'
    agent_player = 1 if agent_first else -1
    return 'W' if w == agent_player else 'L'

def evaluate(net, sims=50, games_vs_random=100, games_vs_minimax=20, seed=123):
    np.random.seed(seed)
    print(f"--- vs random ({games_vs_random} partidas por lado) ---")
    for first in (True, False):
        res = [play_one(first, net, sims, 'random') for _ in range(games_vs_random)]
        lado = 'primero' if first else 'segundo'
        print(f"  agente mueve {lado}: gana {res.count('W')}, tablas {res.count('D')}, pierde {res.count('L')}")
    print(f"--- vs minimax / juego perfecto ({games_vs_minimax} partidas por lado) ---")
    for first in (True, False):
        res = [play_one(first, net, sims, 'minimax') for _ in range(games_vs_minimax)]
        lado = 'primero' if first else 'segundo'
        print(f"  agente mueve {lado}: gana {res.count('W')}, tablas {res.count('D')}, pierde {res.count('L')}")
    st = TicTacToe()
    p, v = net.predict(st.encode())
    print("--- apertura: politica de la red en el tablero vacio ---")
    print("  " + "  ".join(f"{a}:{p[a]:.2f}" for a in st.legal_moves()))
    print(f"  valor estimado del tablero vacio: {v:+.2f} (0 = tablas)")

if __name__ == '__main__':
    ap = argparse.ArgumentParser()
    ap.add_argument('--model', default=None, help='model.npz; si se omite, red SIN entrenar (baseline)')
    ap.add_argument('--sims', type=int, default=50)
    args = ap.parse_args()
    if args.model:
        net = Net.load(args.model)
        print(f"Evaluando modelo entrenado: {args.model}")
    else:
        net = Net(seed=0)
        print("Evaluando red ALEATORIA sin entrenar (baseline)")
    evaluate(net, sims=args.sims)
