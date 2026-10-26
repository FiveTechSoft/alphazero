"""Pruebas rapidas del prototipo Python del tres en raya.

Ejecutar desde la raiz del repositorio:

    python -m unittest discover -s tests -v
"""
import os
import sys
import tempfile
import unittest

import numpy as np

RAIZ = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if RAIZ not in sys.path:
    sys.path.insert(0, RAIZ)

from evaluate import _negamax, minimax_move  # noqa: E402
from game import TicTacToe  # noqa: E402
from mcts import mcts_probs  # noqa: E402
from model import Net  # noqa: E402


class TestJuego(unittest.TestCase):
    def test_inicio_legal_y_no_terminal(self):
        st = TicTacToe()
        self.assertEqual(len(st.legal_moves()), 9)
        self.assertIsNone(st.terminal_value())

    def test_x_gana_fila_superior(self):
        st = TicTacToe()
        for a in (0, 3, 1, 4, 2):
            st.play(a)
        self.assertEqual(st.winner(), 1)
        self.assertEqual(st.terminal_value(), -1.0)

    def test_tablero_lleno_es_tablas(self):
        st = TicTacToe(np.array([1, -1, 1, 1, -1, -1, -1, 1, 1], dtype=np.int8), 1)
        self.assertEqual(st.winner(), 0)
        self.assertEqual(st.terminal_value(), 0.0)

    def test_jugada_ilegal(self):
        st = TicTacToe()
        st.play(0)
        with self.assertRaises(AssertionError):
            st.play(0)

    def test_encode_es_canonico(self):
        st = TicTacToe(np.array([1, 0, 0, 0, 0, 0, 0, 0, 0], dtype=np.int8), -1)
        cod = st.encode()
        self.assertEqual(cod.shape, (27,))
        # Desde el punto de vista de O, la X es la ficha rival (plano 9:18).
        self.assertEqual(cod[0], 0.0)
        self.assertEqual(cod[9], 1.0)


class TestRed(unittest.TestCase):
    def test_politica_es_distribucion_y_valor_acotado(self):
        net = Net(seed=1)
        p, v = net.predict(TicTacToe().encode())
        self.assertEqual(p.shape, (9,))
        self.assertAlmostEqual(float(p.sum()), 1.0, places=5)
        self.assertTrue(np.all(p >= 0.0))
        self.assertGreaterEqual(v, -1.0)
        self.assertLessEqual(v, 1.0)

    def test_guardar_y_cargar_conserva_prediccion(self):
        net = Net(seed=2)
        x = TicTacToe().encode()
        p1, v1 = net.predict(x)
        with tempfile.TemporaryDirectory() as tmp:
            path = os.path.join(tmp, "modelo.npz")
            net.save(path)
            cargada = Net.load(path)
        p2, v2 = cargada.predict(x)
        np.testing.assert_allclose(p1, p2, atol=1e-6)
        self.assertAlmostEqual(v1, v2, places=6)


class TestBusqueda(unittest.TestCase):
    def test_mcts_devuelve_distribucion_sobre_legales(self):
        rng_estado = np.random.get_state()
        np.random.seed(3)
        try:
            st = TicTacToe()
            st.play(4)
            pi = mcts_probs(st, Net(seed=3), sims=12, temp=1.0, add_noise=False)
        finally:
            np.random.set_state(rng_estado)
        legal = st.legal_moves()
        self.assertAlmostEqual(float(pi.sum()), 1.0, places=6)
        self.assertTrue(np.all(pi[legal] >= 0.0))
        for a in range(9):
            if a not in legal:
                self.assertEqual(pi[a], 0.0)

    def test_minimax_no_pierde_desde_vacio(self):
        valor, _ = _negamax(TicTacToe().board.tobytes(), 1)
        self.assertEqual(valor, 0.0)

    def test_minimax_remata_victoria_inmediata(self):
        st = TicTacToe(np.array([1, 1, 0, -1, 0, 0, 0, 0, 0], dtype=np.int8), 1)
        # X tiene dos en raya: cualquier jugada optima debe ganar ya.
        mv = minimax_move(st)
        st.play(mv)
        self.assertEqual(st.winner(), 1)


if __name__ == "__main__":
    unittest.main()
