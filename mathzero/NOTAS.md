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

## 12 oct 2026 (tarde) - decisiones y direccion (Anto)

- "Verificado" pasa a significar: medido SOLO sobre el conjunto ciego
  (blind.json, 700 problemas, 100/nivel, semilla 12345, excluido del entrenamiento).
- Benchmark lenguajes (mediana): Go 2,93M / C 5,51M (1,88x) / Rust 5,74M (1,96x) sims/s.
  **Correccion 26-oct:** la cifra Go era invalida (MCTS no exploraba por falta de
  `Expanded=true`); re-medido: Go 1,28M / C 3,70M (2,90x) / Rust 4,20M (3,28x).
  Juez portado a Rust: 0 discrepancias en 230 vectores. Decision de lenguaje: Go se
  queda como nucleo por ahora; Rust/C quedan medidos como opcion de aceleracion.
- 8 experimentos de la literatura aprobados ("toma todo lo que nos sirva"):
  E1 PUCT+red (HyperTree 2205.11491), E2 online, E3 curriculo al borde (STP 2502.00212),
  E4 replay priorizado (2006.00283), E5 sinteticos (AlphaGeometry), E6 DreamCoder,
  E7 test-time RL (AlphaProof), E8 transferencia. A/B mismo ciego+cap, 3 semillas.
- Panel de capacidades aprobado como hito (h18): 7 metricas, 3 semillas, un comando,
  tabla versionada, visible en la web. Solo medido.
- "Queremos un AlphaZero muy profesional, no un juguete": eje Q anadido (tests+CI
  primero, reproducibilidad, arquitectura, benchmarks, Guia para bobos, CLI, API,
  empaquetado, metricas, evaluacion rigurosa).
- MoE anotado como futuro CONDICIONADO (h22): expertos+router simple primero.
- Minilaboratorio multiagente (h23), largo plazo tras laberinto: con/sin canal de
  comunicacion, 3 semillas. Vision "placa de Petri" de Silver. Sin afirmar nada.
- Comparacion con LLM (h20): mismo conjunto ciego + mismo juez exacto.
- Idea Symbols (h21, vision): la red propone parches, el motor simbolico verifica.
- Web publica: https://fivetechsoft.github.io/alphazero/ (Pages desde main:/docs).
- Analisis de mejora continuo pedido por Anto: donde falla la red (k6-k8), cuellos
  de botella (juez, evaluacion, arquitectura, curriculo), experimentos medibles
  ordenados por impacto/coste, resumen periodico de que mejoro y que no.
