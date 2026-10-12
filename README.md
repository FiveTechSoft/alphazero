# Motor genérico de autojuego estilo AlphaZero

Motor genérico de autojuego estilo AlphaZero (MCTS + red neuronal + juez
exacto) que aprende solo: matemáticas, juegos y puzles.

> Visión (dirección del proyecto, no un resultado medido): "La meta es
> un «superaprendiz» que descubra cosas «tan profundas que no podemos
> expresarlas con las palabras que tenemos hoy»." - David Silver, FT.

Estado a 12 oct 2026. Todo lo listado abajo está medido en la misma
máquina; lo no medido no se afirma.

## Componentes

1. **Núcleo Go (`go/`)**: interfaz `Game` genérica + MCTS (PUCT)
   compilado. Dos juegos sobre el mismo motor: tres en raya y MathRing.
2. **MathRing (`go/mathring.go`)**: demostración de identidades
   algebraicas con juez exacto (forma canónica de polinomio). La red guía
   la búsqueda mejor-primero y aprende por expert iteration con
   currículo de niveles.
3. **Biblioteca de lemas (`go/lemas.go`)**: los caminos ya demostrados
   se guardan y se reutilizan como reescrituras de un solo paso.
4. **Ciclo de automejora (`go/ciclo.go`)**: autojuego → entrenamiento →
   verificación con juez exacto → biblioteca persistente → currículo
   automático → métrica fija por ciclo.
5. **Prototipo Python (raíz)**: AlphaZero completo del tres en raya en
   NumPy puro (red, backprop, MCTS y bucle de autojuego legibles),
   medido contra minimax.

## Medido hasta ahora

- **Tres en raya (Python, 6000 partidas de autojuego)**: 0 derrotas en
  100 partidas contra minimax (50 por lado); la misma red sin entrenar
  pierde 30 de 50 defendiendo. Detalle en `RESULTADOS.md`.
- **Autojuego red contra sí misma** (200 partidas reales, sims=50,
  seed=7, temp=1 solo en las 2 primeras jugadas): 95.5% tablas, 4.5%
  gana X, 0% gana O. Registro en `selfplay_ttt.jsonl`, reproducible con
  `selfplay.py`.
- **MathRing, baseline BFS (cap 2000)**: k4 16/30, k6 6/30.
  Evaluación intermedia tras 6 rondas de expert iteration con el mismo
  cap y conjunto fijo: k4 28/30, k6 18/30.
- **Ciclo de automejora** (3 ciclos, cap 1000, métrica fija de 210
  problemas): resueltos 97 → 140 → 146. Al entrar los primeros 49
  lemas, los nodos medios en k3 caen de 292 a 20; en el ciclo 3 la
  ablación CON/SIN biblioteca da el mismo total (146/210): la red
  acaba interiorizando lo que los lemas aportaban. La biblioteca
  comprime la búsqueda al principio; el efecto se diluye una vez
  aprendido.

## Pasos para ejecutar

1. Tres en raya (Python): `pip install numpy` y luego
   `python train.py --games 1500 --sims 40` (guarda `model.npz`),
   `python evaluate.py --model model.npz`, `python play.py`.
2. Autojuego del tres en raya contra sí misma:
   `python selfplay.py --games 200 --sims 50`.
3. Motor Go: `cd go && go build -o mathzero . && ./mathzero`
   (bench de MCTS en tres en raya + chequeo del juez exacto + baseline
   BFS de MathRing).
4. Aprendizaje de MathRing: `./mathzero learn`
   (12 rondas con currículo; guarda `model_math.json`).
5. Ciclo de automejora: `./mathzero cycle 3 60 1000`
   (3 ciclos, 60 problemas de entreno por ciclo, cap 1000, con ablación
   sin biblioteca en el ciclo 1).

## Hoja de ruta

`mathzero/ROADMAP.md` define hitos con criterio "verificado" medible
cada uno (incluido el benchmark Go vs C vs Rust antes de cualquier
decisión de lenguaje). `mathzero/DISENO_AUTOMEJORA.md` describe el
bucle de crecimiento y sus límites.
