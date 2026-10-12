"""Tres en raya (tic-tac-toe): estado del juego en forma canonica.

El tablero es un array de 9 casillas: 1 = X, -1 = O, 0 = vacia.
`player` es quien mueve (1 o -1). La forma "canonica" multiplica el
tablero por el jugador al que le toca, de modo que la red siempre ve
la posicion desde el punto de vista de quien mueve (truco clasico de
AlphaZero para no aprender dos veces el mismo juego).
"""
import numpy as np

WINS = [(0,1,2),(3,4,5),(6,7,8),(0,3,6),(1,4,7),(2,5,8),(0,4,8),(2,4,6)]

class TicTacToe:
    def __init__(self, board=None, player=1):
        self.board = np.zeros(9, dtype=np.int8) if board is None else board.copy()
        self.player = player  # 1 = X, -1 = O

    def clone(self):
        return TicTacToe(self.board, self.player)

    def legal_moves(self):
        return np.flatnonzero(self.board == 0)

    def play(self, a):
        assert self.board[a] == 0, "jugada ilegal"
        self.board[a] = self.player
        self.player = -self.player

    def winner(self):
        """1 si gana X, -1 si gana O, 0 tablas, None si no ha terminado."""
        for i, j, k in WINS:
            s = int(self.board[i]) + int(self.board[j]) + int(self.board[k])
            if s == 3: return 1
            if s == -3: return -1
        if (self.board != 0).all():
            return 0
        return None

    def terminal_value(self):
        """Valor terminal para el jugador al que le toca; None si sigue el juego."""
        w = self.winner()
        if w is None: return None
        if w == 0: return 0.0
        return -1.0  # gano el que acaba de mover: para quien toca, vale -1

    def encode(self):
        """27 entradas: 9 casillas propias + 9 del rival + 9 unos (plano de turno)."""
        c = (self.board * self.player).astype(np.float32)
        return np.concatenate([c == 1, c == -1, np.ones(9, np.float32)]).astype(np.float32)

    def render(self):
        sym = {1: "X", -1: "O", 0: "."}
        rows = [" ".join(sym[v] for v in self.board[r*3:r*3+3]) for r in range(3)]
        return "\n".join(rows)
