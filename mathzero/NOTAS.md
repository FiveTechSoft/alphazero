# Notas de diseño y backlog (privado)

## Estado medido (12 oct 2026)
- Tres en raya (Python+NumPy): red entrenada 6000 partidas autojuego, 0 derrotas vs minimax (50 partidas por lado). Sin entrenar: 30/50 derrotas moviendo segundo.
- Motor Go (interfaz Game generica, MCTS PUCT): ~1.5-2.2M sims/s vs 17.6K sims/s en Python, mismo test sin red (~90x o mas). Sin red; con red el speedup sera menor.
- MathRing: 14 reglas, juez exacto por forma canonica de polinomio; 0 fallos de reglas. Baseline BFS cap 2000 (k=2..8): ver learn.log / RESULTADOS_MATH.md cuando termine.

## Backlog pedido por Anto (nada hecho salvo lo indicado)
1. Red + autojuego con curriculo y reentrenamiento iterativo desde el modelo guardado (learn.go, en curso). Opcional: guardar caminos como lemas/acciones nuevas.
2. API HTTP/WebSocket local (solo localhost) en el mismo binario Go + visualizacion web; laberinto como tercer Game; verificar con captura real.
3. Sudoku como Game (rejilla, colocar digito, juez exacto, comparar con solucionador de referencia).
4. Motor generico: red y busqueda configurables, accion/estado variable, backends intercambiables.
5. Tensor Logic (idea futura, NO integrar ahora): prueba barata posterior con juez de sudoku en NumPy antes de tocar el motor Go.
6. Lean como verificador enchufable (descartado por ahora por coste de instalacion; misma interfaz).

## Limites honestos
- No hay red neuronal dentro del MCTS Go todavia; la red actual puntua pares (estado, accion) para busqueda mejor-primero.
- Resultados de MathRing se miden con 30 problemas fijos por nivel; muestra pequena.
