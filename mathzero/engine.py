"""MathZero: nucleo del motor de matematicas con juez EXACTO.

Dominio del checkpoint: identidades de anillos conmutativos (polinomios
con variables a,b,c y constantes enteras, con +, * y -).

JUEZ EXACTO: toda expresion se reduce a su forma canonica de polinomio
(diccionario monomio_ordenado -> coeficiente). Dos expresiones son
iguales si y solo si sus formas canonicas coinciden. Sin floats, sin
aproximaciones: es un verificador, no una heuristica.

PROBLEMA: dadas A y B (garantizadas iguales por construccion), encontrar
un camino de reescrituras desde A hasta B usando solo los axiomas.
El objetivo es alcanzar LA ESTRUCTURA de B (normalizada por
conmutatividad/asociatividad), no solo "ser igual": hay que encontrar
el camino de la demostracion.
"""
import random
from collections import deque

# expresiones: ('const', n) ('var', s) ('add', a, b) ('mul', a, b) ('neg', a)

def const(n): return ('const', int(n))
def var(s): return ('var', s)
def add(a, b): return ('add', a, b)
def mul(a, b): return ('mul', a, b)
def neg(a): return ('neg', a)

def canon(e):
    """Forma canonica: dict monomio(tuple ordenada de vars) -> coef entero."""
    op = e[0]
    if op == 'const': return {(): e[1]} if e[1] else {}
    if op == 'var': return {(e[1],): 1}
    if op == 'neg':
        return {m: -c for m, c in canon(e[1]).items()}
    if op == 'add':
        r = dict(canon(e[1]))
        for m, c in canon(e[2]).items():
            r[m] = r.get(m, 0) + c
        return {m: c for m, c in r.items() if c}
    if op == 'mul':
        d1, d2 = canon(e[1]), canon(e[2])
        r = {}
        for m1, c1 in d1.items():
            for m2, c2 in d2.items():
                m = tuple(sorted(m1 + m2))
                r[m] = r.get(m, 0) + c1 * c2
        return {m: c for m, c in r.items() if c}
    raise ValueError(op)

def key(e):
    """Clave estructural normalizada (aplana y ordena ops conmutativas)."""
    op = e[0]
    if op in ('const', 'var'): return str(e)
    if op == 'neg': return 'neg(' + key(e[1]) + ')'
    parts = []
    stack = [e]
    while stack:
        x = stack.pop()
        if x[0] == op: stack.extend([x[1], x[2]])
        else: parts.append(key(x))
    return op + '(' + ','.join(sorted(parts)) + ')'

def subpositions(e, path=()):
    """Todas las posiciones (caminos) del arbol, incluida la raiz ()."""
    yield path
    if e[0] in ('add', 'mul'):
        yield from subpositions(e[1], path + (0,))
        yield from subpositions(e[2], path + (1,))
    elif e[0] == 'neg':
        yield from subpositions(e[1], path + (0,))

def get(e, p):
    for i in p: e = e[i + 1]
    return e

def replace(e, p, new):
    if not p: return new
    i = p[0]
    if e[0] == 'neg': return ('neg', replace(e[1], p[1:], new))
    if i == 0: return (e[0], replace(e[1], p[1:], new), e[2])
    return (e[0], e[1], replace(e[2], p[1:], new))

def is0(x): return x == ('const', 0)
def is1(x): return x == ('const', 1)

# reglas bidireccionales: nombre, f(x)->y|None, finv(y)->x|None
def _r_add0_f(x): return x[1] if x[0]=='add' and is0(x[2]) else (x[2] if x[0]=='add' and is0(x[1]) else None)
def _r_add0_r(x): return add(x, const(0))
def _r_mul1_f(x): return x[1] if x[0]=='mul' and is1(x[2]) else (x[2] if x[0]=='mul' and is1(x[1]) else None)
def _r_mul1_r(x): return mul(x, const(1))
def _r_mul0_f(x): return const(0) if x[0]=='mul' and (is0(x[1]) or is0(x[2])) else None
def _r_comm_a(x): return add(x[2], x[1]) if x[0]=='add' else None
def _r_comm_m(x): return mul(x[2], x[1]) if x[0]=='mul' else None
def _r_assoc_a_f(x): return add(x[1][1], add(x[1][2], x[2])) if x[0]=='add' and x[1][0]=='add' else None
def _r_assoc_a_r(x): return add(add(x[1], x[2][1]), x[2][2]) if x[0]=='add' and x[2][0]=='add' else None
def _r_assoc_m_f(x): return mul(x[1][1], mul(x[1][2], x[2])) if x[0]=='mul' and x[1][0]=='mul' else None
def _r_assoc_m_r(x): return mul(mul(x[1], x[2][1]), x[2][2]) if x[0]=='mul' and x[2][0]=='mul' else None
def _r_dist_f(x): return add(mul(x[1], x[2][1]), mul(x[1], x[2][2])) if x[0]=='mul' and x[2][0]=='add' else None
def _r_dist_r(x):
    if x[0]=='add' and x[1][0]=='mul' and x[2][0]=='mul':
        a,b,c,d = x[1][1], x[1][2], x[2][1], x[2][2]
        if key(a)==key(c): return mul(a, add(b, d))
    return None
def _r_negneg(x): return x[1][1] if x[0]=='neg' and x[1][0]=='neg' else None
def _r_negadd_f(x): return add(neg(x[1][1]), neg(x[1][2])) if x[0]=='neg' and x[1][0]=='add' else None
def _r_negadd_r(x): return neg(add(x[1][1], x[2][1])) if x[0]=='add' and x[1][0]=='neg' and x[2][0]=='neg' else None
def _r_addneg(x): return const(0) if x[0]=='add' and x[2][0]=='neg' and key(x[1])==key(x[2][1]) else (const(0) if x[0]=='add' and x[1][0]=='neg' and key(x[2])==key(x[1][1]) else None)
def _r_negmul_f(x): return mul(neg(x[1][1]), x[1][2]) if x[0]=='neg' and x[1][0]=='mul' else None
def _r_negmul_r(x): return neg(mul(x[1][1], x[2])) if x[0]=='mul' and x[1][0]=='neg' else None
def _r_fold_a(x): return const(x[1][1]+x[2][1]) if x[0]=='add' and x[1][0]=='const' and x[2][0]=='const' else None
def _r_fold_m(x): return const(x[1][1]*x[2][1]) if x[0]=='mul' and x[1][0]=='const' and x[2][0]=='const' else None

RULES = [
    ('add0', _r_add0_f, _r_add0_r), ('mul1', _r_mul1_f, _r_mul1_r),
    ('mul0', _r_mul0_f, None), ('comm+', _r_comm_a, None), ('comm*', _r_comm_m, None),
    ('assoc+', _r_assoc_a_f, _r_assoc_a_r), ('assoc*', _r_assoc_m_f, _r_assoc_m_r),
    ('dist', _r_dist_f, _r_dist_r), ('negneg', _r_negneg, lambda x: neg(neg(x))),
    ('neg+', _r_negadd_f, _r_negadd_r), ('+-0', _r_addneg, None),
    ('neg*', _r_negmul_f, _r_negmul_r), ('fold+', _r_fold_a, None), ('fold*', _r_fold_m, None),
]

def actions(e):
    """(nombre_regla, posicion, expresion_resultado) para todo lo aplicable."""
    out = []
    for p in subpositions(e):
        sub = get(e, p)
        for name, f, finv in RULES:
            for fn in (f, finv):
                if fn is None: continue
                r = fn(sub)
                if r is not None and key(r) != key(sub):
                    out.append((name, p, replace(e, p, r)))
    return out

def random_expr(rng, depth):
    if depth == 0 or rng.random() < 0.3:
        if rng.random() < 0.6: return var(rng.choice('abc'))
        return const(rng.randint(-3, 4))
    op = rng.choice(['add', 'mul', 'neg'])
    if op == 'neg': return neg(random_expr(rng, depth-1))
    return (op, random_expr(rng, depth-1), random_expr(rng, depth-1))

def gen_problem(rng, k):
    """A y B con camino garantizado de <= k reescrituras entre ambos."""
    a = random_expr(rng, rng.randint(2, 3))
    cur = a
    for _ in range(k):
        acts = actions(cur)
        if not acts: break
        cur = rng.choice(acts)[2]
    return a, cur

def bfs_solve(a, b, cap):
    """BASELINE sin red: busqueda en anchura uniforme. Devuelve (ok, nodos)."""
    target = key(b)
    seen = {key(a)}
    q = deque([a])
    nodes = 0
    while q and nodes < cap:
        cur = q.popleft()
        nodes += 1
        if key(cur) == target: return True, nodes
        for _, _, nxt in actions(cur):
            k2 = key(nxt)
            if k2 not in seen:
                seen.add(k2)
                q.append(nxt)
    return False, nodes

if __name__ == '__main__':
    rng = random.Random(42)
    # chequeo del juez: toda regla preserva la forma canonica (identidad real)
    bad = 0
    for _ in range(200):
        e = random_expr(rng, 3)
        for name, p, r in actions(e):
            if canon(e) != canon(r):
                bad += 1; print('REGLA ROTA:', name, e, '->', r)
    print('reglas verificadas sobre 200 expr aleatorias: fallos =', bad)
    # baseline: conjunto fijo de 30 problemas
    rng = random.Random(7)
    for k, cap in ((4, 2000), (6, 4000)):
        rng = random.Random(7)
        probs = [gen_problem(rng, k) for _ in range(30)]
        ok = nodes = 0
        for a, b in probs:
            assert canon(a) == canon(b)  # garantia del generador
            solved, n = bfs_solve(a, b, cap)
            ok += solved; nodes += n
        print(f'baseline BFS: dificultad k={k}, cap={cap}: resueltos {ok}/30, media nodos {nodes//30}')
