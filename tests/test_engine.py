"""Pruebas del prototipo Python del juez exacto de MathRing.

Ejecutar desde la raiz del repositorio:

    python -m unittest discover -s tests -v
"""
import os
import random
import sys
import unittest

RAIZ = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
MATHZERO = os.path.join(RAIZ, "mathzero")
for ruta in (RAIZ, MATHZERO):
    if ruta not in sys.path:
        sys.path.insert(0, ruta)

import engine as mz  # noqa: E402


class TestJuez(unittest.TestCase):
    def test_identidades(self):
        a, b, c = mz.var("a"), mz.var("b"), mz.var("c")
        iguales = [
            (mz.mul(mz.add(a, b), mz.add(a, b)),
             mz.add(mz.add(mz.mul(a, a), mz.mul(mz.const(2), mz.mul(a, b))), mz.mul(b, b))),
            (mz.mul(a, mz.add(b, c)), mz.add(mz.mul(a, b), mz.mul(a, c))),
            (mz.add(a, mz.neg(a)), mz.const(0)),
            (mz.neg(mz.neg(a)), a),
        ]
        for x, y in iguales:
            self.assertEqual(mz.canon(x), mz.canon(y))

    def test_no_identidades(self):
        a, b = mz.var("a"), mz.var("b")
        self.assertNotEqual(mz.canon(a), mz.canon(b))
        self.assertNotEqual(mz.canon(mz.add(a, b)), mz.canon(mz.mul(a, b)))
        self.assertNotEqual(mz.canon(mz.mul(a, a)), mz.canon(a))

    def test_key_es_conmutativa_y_asociativa(self):
        a, b, c = mz.var("a"), mz.var("b"), mz.var("c")
        self.assertEqual(mz.key(mz.add(a, b)), mz.key(mz.add(b, a)))
        self.assertEqual(mz.key(mz.mul(a, b)), mz.key(mz.mul(b, a)))
        self.assertEqual(mz.key(mz.add(mz.add(a, b), c)),
                         mz.key(mz.add(a, mz.add(b, c))))

    def test_toda_regla_preserva_la_forma_canonica(self):
        rng = random.Random(42)
        comprobadas = 0
        for _ in range(100):
            e = mz.random_expr(rng, 3)
            for nombre, _, r in mz.actions(e):
                self.assertEqual(mz.canon(e), mz.canon(r),
                                 msg=f"regla {nombre}: {e} -> {r}")
                comprobadas += 1
        self.assertGreater(comprobadas, 100)

    def test_get_replace_es_inverso(self):
        e = mz.add(mz.mul(mz.var("a"), mz.neg(mz.var("b"))),
                   mz.add(mz.const(1), mz.mul(mz.const(2), mz.var("c"))))
        for p in mz.subpositions(e):
            self.assertEqual(mz.key(mz.replace(e, p, mz.get(e, p))), mz.key(e))

    def test_problemas_generados_se_resuelven(self):
        rng = random.Random(7)
        for _ in range(5):
            a, b = mz.gen_problem(rng, 3)
            self.assertEqual(mz.canon(a), mz.canon(b))
            ok, nodos = mz.bfs_solve(a, b, 5000)
            self.assertTrue(ok, msg=f"BFS no resolvio {a} -> {b} en {nodos} nodos")


if __name__ == "__main__":
    unittest.main()
