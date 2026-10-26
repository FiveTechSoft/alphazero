"""Tests de la red NumPy pura (model.py): forward, entrenamiento y disco."""
import os
import sys
import tempfile

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import unittest
import numpy as np

from model import Net


class TestModel(unittest.TestCase):
    def setUp(self):
        np.random.seed(0)

    def test_predict_formas(self):
        net = Net(seed=0)
        x = np.random.rand(27).astype(np.float32)
        p, v = net.predict(x)
        self.assertEqual(p.shape, (9,))
        self.assertAlmostEqual(float(p.sum()), 1.0, places=5)
        self.assertTrue(-1.0 <= v <= 1.0)

    def test_train_step_baja_perdida(self):
        # tarea juguete: predecir politico fija y valor fijo
        net = Net(seed=0)
        B = 64
        X = np.random.rand(B, 27).astype(np.float32)
        PI = np.zeros((B, 9), np.float32)
        PI[:, 3] = 1.0
        Z = np.ones(B, np.float32)
        l0p, l0v = net.train_step(X, PI, Z, lr=3e-3, wd=0.0)
        for _ in range(200):
            lp, lv = net.train_step(X, PI, Z, lr=3e-3, wd=0.0)
        self.assertLess(lp, l0p)
        self.assertLess(lv, l0v)

    def test_save_load_roundtrip(self):
        net = Net(seed=7)
        x = np.random.rand(27).astype(np.float32)
        p1, v1 = net.predict(x)
        with tempfile.TemporaryDirectory() as d:
            path = os.path.join(d, "model.npz")
            net.save(path)
            net2 = Net.load(path)
        p2, v2 = net2.predict(x)
        np.testing.assert_allclose(p1, p2, atol=1e-6)
        self.assertAlmostEqual(v1, v2, places=6)

    def test_lr_wd_no_nan(self):
        net = Net(seed=2)
        X = np.random.rand(16, 27).astype(np.float32)
        PI = np.full((16, 9), 1 / 9, np.float32)
        Z = np.zeros(16, np.float32)
        for _ in range(50):
            lp, lv = net.train_step(X, PI, Z, lr=1e-3, wd=1e-4)
            self.assertTrue(np.isfinite(lp) and np.isfinite(lv))


if __name__ == "__main__":
    unittest.main()
