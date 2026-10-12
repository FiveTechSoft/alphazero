"""Red neuronal minima en NumPy puro: un perceptron con una capa oculta
y dos cabezas, como en AlphaZero:

  - cabeza de politica: probabilidad de cada jugada (softmax sobre 9)
  - cabeza de valor:    quien gana desde esta posicion (tanh en [-1, 1])

Sin PyTorch ni TensorFlow: el backprop esta escrito a mano y el
optimizador es Adam. Todo cabe en una pantalla y se puede leer entero.
"""
import numpy as np

class Net:
    def __init__(self, n_in=27, n_h=128, n_out=9, seed=0):
        rng = np.random.default_rng(seed)
        self.W1 = (rng.normal(0, np.sqrt(2/n_in), (n_in, n_h))).astype(np.float32)
        self.b1 = np.zeros(n_h, np.float32)
        self.Wp = (rng.normal(0, np.sqrt(2/n_h), (n_h, n_out))).astype(np.float32)
        self.bp = np.zeros(n_out, np.float32)
        self.Wv = (rng.normal(0, np.sqrt(2/n_h), (n_h, 1))).astype(np.float32)
        self.bv = np.zeros(1, np.float32)
        self.m = {}  # momentos de Adam
        self.v = {}
        self.t = 0

    def forward(self, X):
        h_pre = X @ self.W1 + self.b1
        h = np.maximum(h_pre, 0)
        logits = h @ self.Wp + self.bp
        logits = logits - logits.max(axis=1, keepdims=True)
        e = np.exp(logits)
        p = e / e.sum(axis=1, keepdims=True)
        v = np.tanh(h @ self.Wv + self.bv)
        return p, v, (X, h_pre, h)

    def predict(self, x):
        p, v, _ = self.forward(x[None, :])
        return p[0], float(v[0, 0])

    def train_step(self, X, PI, Z, lr=1e-3, wd=1e-4):
        B = X.shape[0]
        p, v, (Xc, h_pre, h) = self.forward(X)
        # perdida: entropia cruzada (politica) + error cuadratico (valor)
        dlogits = (p - PI) / B
        dv = (v - Z.reshape(-1, 1)) * (2.0 / B)
        dh = dlogits @ self.Wp.T + (dv * (1 - v**2)) @ self.Wv.T
        dh = dh * (h_pre > 0)
        grads = {
            'W1': Xc.T @ dh + wd * self.W1,  'b1': dh.sum(0),
            'Wp': h.T @ dlogits + wd * self.Wp, 'bp': dlogits.sum(0),
            'Wv': h.T @ dv + wd * self.Wv,   'bv': dv.sum(0),
        }
        beta1, beta2, eps = 0.9, 0.999, 1e-8
        self.t += 1
        for name, g in grads.items():
            self.m[name] = beta1 * self.m.get(name, 0) + (1 - beta1) * g
            self.v[name] = beta2 * self.v.get(name, 0) + (1 - beta2) * g * g
            mh = self.m[name] / (1 - beta1 ** self.t)
            vh = self.v[name] / (1 - beta2 ** self.t)
            setattr(self, name, getattr(self, name) - lr * mh / (np.sqrt(vh) + eps))
        loss_p = float(-(PI * np.log(p + 1e-12)).sum(1).mean())
        loss_v = float(((v - Z.reshape(-1, 1)) ** 2).mean())
        return loss_p, loss_v

    def save(self, path):
        np.savez(path, W1=self.W1, b1=self.b1, Wp=self.Wp, bp=self.bp,
                 Wv=self.Wv, bv=self.bv)

    @classmethod
    def load(cls, path):
        d = np.load(path)
        n = cls()
        for k in ['W1', 'b1', 'Wp', 'bp', 'Wv', 'bv']:
            setattr(n, k, d[k].astype(np.float32))
        return n
