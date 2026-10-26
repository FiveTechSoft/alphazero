"""Tests del estado de tres en raya (game.py)."""
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import unittest
import numpy as np

from game import TicTacToe


class TestGame(unittest.TestCase):
    def test_tablero_vacio(self):
        g = TicTacToe()
        self.assertEqual(len(g.legal_moves()), 9)
        self.assertIsNone(g.winner())
        self.assertIsNone(g.terminal_value())

    def test_play_alternacia(self):
        g = TicTacToe()
        g.play(0)
        self.assertEqual(g.player, -1)
        self.assertEqual(g.board[0], 1)
        g.play(4)
        self.assertEqual(g.board[4], -1)
        with self.assertRaises(AssertionError):
            g.play(0)  # casilla ocupada

    def test_victoria_x(self):
        g = TicTacToe()
        for a in [0, 3, 1, 4, 2]:
            g.play(a)
        self.assertEqual(g.winner(), 1)
        # gano X (jugador previo): para quien toca ahora vale -1
        self.assertEqual(g.terminal_value(), -1.0)

    def test_victoria_o(self):
        g = TicTacToe(board=np.array([1, 1, 0, -1, -1, 0, 0, 0, 0], dtype=np.int8), player=-1)
        g.play(5)
        self.assertEqual(g.winner(), -1)
        self.assertEqual(g.terminal_value(), -1.0)

    def test_tablas(self):
        g = TicTacToe(board=np.array([1, -1, 1, 1, -1, -1, -1, 1, 1], dtype=np.int8), player=1)
        self.assertEqual(g.winner(), 0)
        self.assertEqual(g.terminal_value(), 0.0)

    def test_encode_perspectiva(self):
        g = TicTacToe()
        g.play(0)  # X en 0, toca O
        e = g.encode()
        self.assertEqual(e.shape, (27,))
        self.assertEqual(int(e[0]), 0)       # propia del que mueve (O): no
        self.assertEqual(int(e[9]), 1)       # rival (X) en 0: si
        self.assertEqual(int(e[18:27].sum()), 9)  # plano de turno

    def test_clone_independiente(self):
        g = TicTacToe()
        g.play(0)
        h = g.clone()
        h.play(1)
        self.assertEqual(g.legal_moves().size, 8)


if __name__ == "__main__":
    unittest.main()
