# AlphaZero mini (tres en raya)

Prototipo pequeno pero completo del algoritmo **AlphaZero** (DeepMind, 2017),
escrito en Python + NumPy puro, sin PyTorch ni TensorFlow: la red, el
backprop, el MCTS y el bucle de autojuego se pueden leer enteros en una
tarde. El juego es el tres en raya porque esta resuelto: se puede medir el
aprendizaje contra un rival perfecto (minimax), no solo "parece que juega
bien".

## Como funciona

1. **Autojuego** (`train.py`): el MCTS juega contra si mismo. Cada jugada se
   elige con 60 simulaciones de busqueda guiadas por la red (PUCT + ruido de
   Dirichlet en la raiz para explorar).
2. **Memoria**: cada posicion se guarda con la politica que encontro el MCTS
   y, al final de la partida, con el resultado real (+1 / 0 / -1).
3. **Entrenamiento** (`model.py`): la red (MLP de una capa oculta con cabezas
   de politica y valor) aprende a imitar la politica del MCTS y a predecir el
   resultado. Con una red mejor, el MCTS juega mejor; y el ciclo se repite.

## Requisitos

Solo Python 3.9+ y NumPy (`pip install numpy`). Nada mas.

## Pasos para reproducirlo

1. Descomprime el archivo y entra en la carpeta: `cd alphazero_prototipo`
2. Instala la unica dependencia: `pip install numpy`
3. (Opcional) Mira el nivel de una red sin entrenar: `python evaluate.py`
4. Entrena desde cero: `python train.py --games 1500 --sims 40`
   (unos 2-3 minutos en un portatil; guarda `model.npz`)
5. Mide el resultado: `python evaluate.py --model model.npz`
   - contra un jugador aleatorio (100 partidas por lado)
   - contra minimax, el juego perfecto (20 partidas por lado):
     el objetivo es **no perder nunca** (el tres en raya optimo es tablas)
6. Juega tu contra el: `python play.py`

## Que se midio en esta copia

Red entrenada desde cero con 6000 partidas de autojuego (~5 minutos):

| Rival | Antes de entrenar | Despues de 6000 partidas |
|---|---|---|
| Aleatorio (moviendo primero) | 96G 4T 0P | 99G 1T 0P |
| Aleatorio (moviendo segundo) | 83G 13T 4P | 84G 16T 0P |
| Minimax perfecto (primero) | 0G 50T 0P | 0G 50T 0P |
| Minimax perfecto (segundo) | 0G 20T **30P** | 0G 50T **0P** |

La "tabla" que delata el aprendizaje no es ganarle al azar (el MCTS ya lo
consigue sin entrenar), sino la fila de minimax moviendo segundo: ahi el
agente pasa de perder la mitad de las partidas a no perder ninguna.

## Escala

Todo esta parametrizado: cambia el juego en `game.py` (misma interfaz:
`legal_moves`, `play`, `terminal_value`, `encode`) y el resto funciona igual.
Conecta 4 o un Othello 4x4 son buenos siguientes pasos; ahi la red ya
necesitaria ser convolucional y PyTorch empezaria a compensar.

## Archivos

- `game.py`    - reglas del tres en raya y forma canonica del tablero
- `model.py`   - red neuronal + backprop + Adam, todo en NumPy
- `mcts.py`    - Monte Carlo Tree Search PUCT guiado por la red
- `train.py`   - bucle de autojuego y entrenamiento
- `evaluate.py`- medicion contra aleatorio y contra minimax
- `play.py`    - partida interactiva contra el agente
