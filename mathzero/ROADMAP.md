# ROADMAP (privado) - actualizado 12 oct 2026 15:12

"Conseguido" = existe y corre. "Verificado" = medido y repetible (comando + numero).

| # | Hito | Criterio de verificado | Conseguido | Verificado |
|---|------|------------------------|-----------|-----------|
| 1 | Tres en raya AlphaZero (Python/NumPy) | 0 derrotas vs minimax, 50 partidas por lado | 100% | 100% |
| 2 | Motor generico Go (interfaz Game, MCTS) | 2 Games distintas en el mismo binario + bench sims/s | 80% | 60% |
| 3 | MathRing + juez exacto | 0 fallos de reglas; baseline BFS reproducible Python=Go | 100% | 90% |
| 4 | Red + autojuego (expert iteration) | tasa de resueltos > baseline BFS con mismo conjunto y cap | 50% | 0% |
| 5 | Curriculo automatico | sube de nivel solo al dominar; eval fija por nivel sin olvido | 40% | 0% |
| 6 | Biblioteca acumulativa (lemas/macros) | problemas nuevos resueltos con menos nodos usando la biblioteca | 0% | 0% |
| 7 | Core auto-mejorante (ciclo medido) | metrica repetible por ciclo, curva creciente | 10% | 0% |
| 8 | API local + visualizacion web | captura real de pantalla, solo localhost | 0% | 0% |
| 9 | Laberinto (Game) | resuelto en el motor, medido vs BFS | 0% | 0% |
| 10 | Sudoku (Game) | % resueltos vs solucionador de referencia | 0% | 0% |
| 11 | Velocidad Go vs Python | mismo test, sims/s | 100% | 80% |
| 12 | Medicion repetible | un comando reproduce todas las tablas | 30% | 0% |
| 13 | Idea futura: Tensor Logic con juez sudoku NumPy | solo nota, no integrar | 0% | 0% |
| 14 | Decision de lenguaje (Go vs C vs Rust) | micro-benchmark repetible del bucle caliente (MCTS + juez) en Go, C y Rust | 10% | 0% |
