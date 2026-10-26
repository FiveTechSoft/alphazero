"""Tests del MCTS PUCT (mcts.py). Red sustituta determinista: prior
uniforme y valor 0, analogo al net=nil de la suite Go."""
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import unittest
import numpy as np

from game import TicTacToe
from mcts import mcts_probs, Node, _expand
from model import Net


class FakeNet:
    def predict(self, x):
        return np.full(9, 1 / 9, np.float32), 0.0


class TestMCTS(unittest.TestCase):
    def test_probas_suman_uno(self):
        s = TicTacToe()
        pi = mcts_probs(s, FakeNet(), sims=30, add_noise=False)
        self.assertAlmostEqual(float(pi.sum()), 1.0, places=6)
        self.assertEqual(pi.shape, (9,))

    def test_distribucion_con_red_sin_entrenar(self):
        s = TicTacToe()
        pi = mcts_probs(s, Net(seed=0), sims=30, add_noise=False)
        self.assertAlmostEqual(float(pi.sum()), 1.0, places=6)
        self.assertTrue((pi >= 0).all())

    def test_elige_jugada_ganadora_no_primer_legal(self):
        # X tiene 3,4 (gana en 5); O amenaza 0,1 con el 2 (primera legal).
        s = TicTacToe(board=np.array([-1, -1, 0, 1, 1, 0, 0, 0, 0], dtype=np.int8), player=1)
        pi = mcts_probs(s, FakeNet(), sims=200, temp=1.0, add_noise=False)
        legal = s.legal_moves()
        self.assertEqual(int(legal[0]), 2)       # primera legal no es la ganadora
        self.assertEqual(int(np.argmax(pi)), 5)  # el MCTS debe jugar 5
        s.play(5)
        self.assertEqual(s.winner(), 1)

    def test_temp_baja_dominio_de_argmax(self):
        # temp=0.1 agudiza: en posicion ganadora, el mas visitado domina claro
        s = TicTacToe(board=np.array([-1, -1, 0, 1, 1, 0, 0, 0, 0], dtype=np.int8), player=1)
        pi = mcts_probs(s, FakeNet(), sims=200, temp=0.1, add_noise=False)
        self.assertEqual(int(np.argmax(pi)), 5)
        self.assertGreater(float(pi.max()), 0.9)

    def test_expand_devuelve_valor_finito(self):
        s = TicTacToe()
        n = Node()
        v = _expand(n, s, Net(seed=1), False, 0.0, 0.0)
        self.assertTrue(np.isfinite(v))
        self.assertTrue(n.expanded)
        self.assertEqual(len(n.prior), 9)


if __name__ == "__main__":
    unittest.main()
