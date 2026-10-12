"""Autojuego: la red entrenada contra si misma, partidas reales registradas.

Cada partida se guarda en selfplay_ttt.jsonl (una linea JSON por partida:
jugadas, ganador, longitud). Para que las partidas no sean todas identicas
(la red con temp=0 es determinista) se muestrean las dos primeras jugadas
con temp=1 y el resto se juega con temp=0 (la mejor jugada segun el MCTS).
"""
import argparse, json
import numpy as np
from game import TicTacToe
from mcts import mcts_probs
from model import Net

def play_one(net, sims, rng):
    st = TicTacToe()
    moves = []
    ply = 0
    while st.terminal_value() is None:
        temp = 1.0 if ply < 2 else 0.0
        pi = mcts_probs(st, net, sims=sims, temp=temp, add_noise=False)
        legal = st.legal_moves()
        if temp > 1e-3:
            probs = np.array([pi[a] for a in legal], dtype=np.float64)
            probs = probs / probs.sum()
            a = int(rng.choice(legal, p=probs))
        else:
            a = int(pi.argmax())
        moves.append(a)
        st.play(a)
        ply += 1
    return moves, st.winner()

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--games', type=int, default=200)
    ap.add_argument('--sims', type=int, default=50)
    ap.add_argument('--seed', type=int, default=7)
    ap.add_argument('--out', default='selfplay_ttt.jsonl')
    args = ap.parse_args()
    net = Net.load('model.npz')
    rng = np.random.default_rng(args.seed)
    res = {1: 0, -1: 0, 0: 0}
    with open(args.out, 'w') as f:
        for g in range(args.games):
            moves, w = play_one(net, args.sims, rng)
            res[w] += 1
            f.write(json.dumps({'game': g, 'moves': moves, 'winner': w,
                                'plies': len(moves)}) + '\n')
    n = args.games
    print(f"partidas={n} sims={args.sims} seed={args.seed}")
    print(f"X gana: {res[1]} ({100*res[1]/n:.1f}%)  O gana: {res[-1]} ({100*res[-1]/n:.1f}%)  tablas: {res[0]} ({100*res[0]/n:.1f}%)")
    print(f"registro: {args.out}")

if __name__ == '__main__':
    main()
