"""Bucle de entrenamiento AlphaZero: autojuego -> buffer -> SGD.

Cada partida la juega el MCTS contra si mismo (con ruido de Dirichlet
para explorar). Las posiciones visitadas se guardan con la politica que
encontro el MCTS y, al acabar la partida, con el resultado real. Eso es
todo el "aprendizaje por refuerzo": la red aprende a imitar al MCTS y a
predecir quien gana; el MCTS con una red mejor juega mejor; y vuelta a
empezar.
"""
import argparse
import random
import time
from collections import deque

import numpy as np

from game import TicTacToe
from mcts import mcts_probs
from model import Net


def self_play_game(net, sims, temp_moves=4):
    state = TicTacToe()
    samples = []
    while state.terminal_value() is None:
        temp = 1.0 if len(samples) < temp_moves else 0.1
        pi = mcts_probs(state, net, sims=sims, temp=temp)
        samples.append((state.encode(), pi, state.player))
        legal = state.legal_moves()
        probs = np.array([pi[a] for a in legal], dtype=np.float64)
        probs = probs / probs.sum()
        a = int(np.random.choice(legal, p=probs))
        state.play(a)
    w = state.winner()
    out = []
    for enc, pi, pl in samples:
        z = 0.0 if w == 0 else (1.0 if w == pl else -1.0)
        out.append((enc, pi, z))
    return out


def sample_batch(buffer, batch):
    idx = np.random.choice(len(buffer), size=min(batch, len(buffer)), replace=False)
    X = np.stack([buffer[i][0] for i in idx])
    PI = np.stack([buffer[i][1] for i in idx])
    Z = np.array([buffer[i][2] for i in idx], dtype=np.float32)
    return X, PI, Z


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--games', type=int, default=1000, help='partidas de autojuego')
    ap.add_argument('--sims', type=int, default=30, help='simulaciones MCTS por jugada')
    ap.add_argument('--batch', type=int, default=64)
    ap.add_argument('--steps-per-game', type=int, default=8)
    ap.add_argument('--lr', type=float, default=1e-3)
    ap.add_argument('--out', default='model.npz')
    ap.add_argument('--seed', type=int, default=0)
    ap.add_argument('--log-every', type=int, default=50)
    args = ap.parse_args()

    np.random.seed(args.seed)
    random.seed(args.seed)
    net = Net(seed=args.seed)
    buffer = deque(maxlen=20000)
    t0 = time.time()
    lp = lv = 0.0
    nsteps = 0
    for g in range(1, args.games + 1):
        buffer.extend(self_play_game(net, args.sims))
        if len(buffer) >= 2 * args.batch:
            for _ in range(args.steps_per_game):
                X, PI, Z = sample_batch(buffer, args.batch)
                a, b = net.train_step(X, PI, Z, lr=args.lr)
                lp += a; lv += b; nsteps += 1
        if g % args.log_every == 0:
            print(f"partida {g:5d}/{args.games}  buffer={len(buffer):5d}  "
                  f"loss_pol={lp/max(nsteps,1):.4f}  loss_val={lv/max(nsteps,1):.4f}  "
                  f"({time.time()-t0:.0f}s)", flush=True)
            lp = lv = 0.0; nsteps = 0
    net.save(args.out)
    print(f"modelo guardado en {args.out}  ({time.time()-t0:.0f}s total)")

if __name__ == '__main__':
    main()
