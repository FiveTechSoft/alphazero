# ROADMAP (privado) - actualizado 12 oct 2026 15:40

Vision (no es un resultado, es la direccion): "La meta es un
<<superaprendiz>> que descubra cosas <<tan profundas que no podemos
expresarlas con las palabras que tenemos hoy>>." - David Silver, FT.

"Conseguido" = existe y corre. "Verificado" = medido y repetible (comando + numero).

| # | Hito | Criterio de verificado | Conseguido | Verificado |
|---|------|------------------------|-----------|-----------|
| 1 | Tres en raya AlphaZero (Python/NumPy) | 0 derrotas vs minimax, 50 partidas por lado | 100% | 100% |
| 2 | Motor generico (interfaz Game, MCTS) (interfaz Game, MCTS) | 2 Games distintas en el mismo binario + bench sims/s | 80% | 60% |
| 3 | MathRing + juez exacto | 0 fallos de reglas; baseline BFS reproducible Python=Go | 100% | 90% |
| 4 | Red + autojuego (expert iteration) | tasa de resueltos > baseline BFS con mismo conjunto y cap | 90% | 70% |
| 5 | Curriculo automatico | sube de nivel solo al dominar; eval fija por nivel sin olvido | 70% | 50% |
| 6 | Biblioteca acumulativa (lemas/macros) | problemas nuevos resueltos con menos nodos usando la biblioteca | 90% | 50% |
| 7 | Core auto-mejorante (ciclo medido) | metrica repetible por ciclo, curva creciente | 80% | 60% |
| 8 | Web visual (Go real compilado a WASM) + GitHub Pages | captura real del sitio publico cargando y jugando | 60% | 30% |
| 9 | Laberinto (Game) | resuelto en el motor, medido vs BFS | 0% | 0% |
| 10 | Sudoku (Game) | % resueltos vs solucionador de referencia | 0% | 0% |
| 11 | Velocidad Go vs Python | mismo test, sims/s | 100% | 80% |
| 12 | Medicion repetible | un comando reproduce todas las tablas | 40% | 0% |
| 13 | Idea futura: Tensor Logic con juez sudoku NumPy | solo nota, no integrar | 0% | 0% |
| 14 | Decision de lenguaje (Go vs C vs Rust) | micro-benchmark repetible del bucle caliente (MCTS + juez) en Go, C y Rust | 10% | 0% |
| 15 | Descubrimiento | la biblioteca produce un lema NUEVO (no esta en el curriculo ni es instancia de las 14 reglas dadas), verificado por el juez exacto; prueba repetible: un comando lo genera y lo valida. Sin afirmar crecimiento ilimitado | 0% | 0% |
