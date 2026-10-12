"""Monte Carlo Tree Search guiado por la red (variante PUCT de AlphaZero).

Cada simulacion baja por el arbol escogiendo la jugada que maximiza
Q + U (explotacion + exploracion ponderada por la prior de la red),
expande la hoja evaluandola con la red, y propaga el valor hacia arriba
cambiando el signo en cada nivel (lo que es bueno para mi es malo para
el rival).
"""
import numpy as np
from game import TicTacToe

class Node:
    __slots__ = ('prior', 'n', 'w', 'children', 'expanded')
    def __init__(self):
        self.prior = {}     # jugada -> probabilidad a priori (red)
        self.n = {}         # jugada -> visitas
        self.w = {}         # jugada -> valor acumulado
        self.children = {}  # jugada -> Node
        self.expanded = False

def _expand(node, state, net, add_noise, dir_alpha, dir_eps):
    p, v = net.predict(state.encode())
    legal = state.legal_moves()
    mask = np.zeros(9, np.float32)
    mask[legal] = p[legal]
    mask = mask / mask.sum()
    priors = {int(a): float(mask[a]) for a in legal}
    if add_noise:
        # ruido de Dirichlet en la raiz: obliga a explorar jugadas raras
        noise = np.random.dirichlet([dir_alpha] * len(legal))
        for a, eta in zip(legal, noise):
            priors[int(a)] = (1 - dir_eps) * priors[int(a)] + dir_eps * float(eta)
    node.prior = priors
    node.n = {a: 0 for a in priors}
    node.w = {a: 0.0 for a in priors}
    node.expanded = True
    return v

def mcts_probs(state, net, sims=50, c_puct=1.5, temp=1.0,
               dir_alpha=0.5, dir_eps=0.25, add_noise=True):
    """Devuelve pi: distribucion de 9 probabilidades segun las visitas del MCTS."""
    root = Node()
    _expand(root, state, net, add_noise, dir_alpha, dir_eps)
    for _ in range(sims):
        s = state.clone()
        node = root
        path = []
        tv = s.terminal_value()
        while tv is None and node.expanded:
            total = sum(node.n.values())
            best_a, best_score = None, -np.inf
            for a, p in node.prior.items():
                q = node.w[a] / node.n[a] if node.n[a] else 0.0
                u = c_puct * p * np.sqrt(total + 1) / (1 + node.n[a])
                if q + u > best_score:
                    best_score, best_a = q + u, a
            path.append((node, best_a))
            s.play(best_a)
            node = node.children.setdefault(best_a, Node())
            tv = s.terminal_value()
        if tv is not None:
            v = tv
        else:
            v = _expand(node, s, net, False, 0, 0)
        for nd, a in reversed(path):
            v = -v
            nd.n[a] += 1
            nd.w[a] += v
    counts = np.array([root.n.get(a, 0) for a in range(9)], dtype=np.float64)
    if temp <= 1e-3:
        pi = np.zeros(9)
        pi[int(counts.argmax())] = 1.0
    else:
        c = counts ** (1.0 / temp)
        if c.sum() > 0:
            pi = c / c.sum()
        else:
            legal = state.legal_moves()
            pi = np.zeros(9); pi[legal] = 1.0 / len(legal)
    return pi
