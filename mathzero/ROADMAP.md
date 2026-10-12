# ROADMAP (privado) - actualizado 12 oct 2026 17:20

Vision (no es un resultado, es la direccion): "La meta es un
<<superaprendiz>> que descubra cosas <<tan profundas que no podemos
expresarlas con las palabras que tenemos hoy>>." - David Silver, FT.
Inspiracion a larguisimo plazo: la "placa de Petri" de Silver (FT) - ver h23.
Todo lo marcado "vision" es direccion, no un resultado.

"Conseguido" = existe y corre. "Verificado" = medido y repetible (comando + numero).
REGLA NUEVA (12 oct 16:00): en los hitos de matematicas, "verificado" solo
cuenta lo medido sobre el CONJUNTO CIEGO (blind.json, 700 problemas, 100/nivel,
semilla 12345, sin solape con levels.json, excluido del entrenamiento).

| # | Hito | Criterio de verificado | Conseguido | Verificado |
|---|------|------------------------|-----------|-----------|
| 1 | Tres en raya AlphaZero (Python/NumPy) | 0 derrotas vs minimax, 50 partidas por lado | 100% | 100% |
| 2 | Motor generico (interfaz Game, MCTS/PUCT) | 2 Games distintas en el mismo binario + bench sims/s | 90% | 70% |
| 3 | MathRing + juez exacto | 0 fallos de reglas; baseline BFS reproducible Python=Go | 100% | 90% |
| 4 | Red + autojuego (expert iteration) | tasa de resueltos > baseline BFS con mismo conjunto y cap | 90% | 70% |
| 5 | Curriculo automatico | sube de nivel solo al dominar; eval fija por nivel sin olvido | 70% | 50% |
| 6 | Biblioteca acumulativa (lemas/macros) | problemas nuevos resueltos con menos nodos usando la biblioteca | 90% | 50% |
| 7 | Core auto-mejorante (ciclo medido) | metrica repetible por ciclo, curva creciente (ciego: 330 -> 382 -> 450/700) | 85% | 70% |
| 8 | Web visual (Go real compilado a WASM) + GitHub Pages | captura real del sitio publico cargando y jugando | 100% | 70% |
| 9 | Laberinto (Game) | resuelto en el motor, medido vs BFS | 0% | 0% |
| 10 | Sudoku (Game) | % resueltos vs solucionador de referencia | 0% | 0% |
| 11 | Velocidad Go vs Python | mismo test, sims/s | 100% | 80% |
| 12 | Medicion repetible | un comando reproduce todas las tablas | 50% | 30% |
| 13 | Idea futura: Tensor Logic con juez sudoku NumPy | solo nota, no integrar | 0% | 0% |
| 14 | Decision de lenguaje (Go vs C vs Rust) | micro-benchmark repetible del bucle caliente MCTS en Go, C y Rust | 95% | 85% |
| 15 | Descubrimiento | la biblioteca produce un lema NUEVO (no esta en el curriculo ni es instancia de las 14 reglas), verificado por el juez exacto; un comando lo genera y lo valida. Sin afirmar crecimiento ilimitado | 0% | 0% |
| 16 | Medicion ciega (curva + ablaciones) | un comando (`./mathzero curve`): curva por ciclo (resueltos, nodos medios, longitud de prueba) + ablaciones con/sin red y con/sin biblioteca sobre 700 problemas ciegos | 100% | 70% |
| 17 | Juez exacto en Rust (y C si barato) | 0 discrepancias vs el juez Go en los 230 vectores + velocidades medidas: canon Rust 4,2x Go; BFS empate 1,06x | 95% | 85% |
| 18 | Panel de capacidades | un comando, 3 semillas, tabla versionada en resultados/*.json o csv, visible en la web: (1) % por nivel k2-k8 ciego; (2) nodos/tiempo y longitud de prueba; (3) generalizacion k2-k5 -> k6-k8 y otro dominio; (4) curva por ciclo; (5) version nueva vs anterior (Elo); (6) lemas nuevos verificados; (7) referencias externas (demostrador clasico, luego LLM con el mismo juez) | 10% | 0% |
| 19 | Eje de calidad profesional (Q1-Q10 abajo) | cada criterio Q verificable en CI | 40% | 25% |
| 20 | Comparacion con LLM | mismo conjunto ciego + mismo juez exacto; tabla lado a lado | 0% | 0% |
| 21 | Idea Symbols (nota/vision) | la red propone parches/reescrituras y un motor simbolico verifica; solo diseno hasta tener h18 | 0% | 0% |
| 22 | MoE (futuro, CONDICIONADO) | primero expertos por dominio + router simple por tipo de problema; MoE aprendido SOLO si el panel (h18) muestra beneficio de compartir parametros | 0% | 0% |
| 23 | Minilaboratorio multiagente (largo plazo, tras h9) | 2-3 agentes se coordinan para resolver un puzle con juez (uno ve el laberinto, otro mueve); criterio: tasa de resolucion con y sin canal de comunicacion, 3 semillas. Vision "placa de Petri"; sin afirmar nada sobre superinteligencia | 0% | 0% |
| 24 | Dominio futuro: geometria con la razon aurea phi | enunciados verificables en Q(raiz5) por el juez exacto (p.ej. phi^2=phi+1, coordenadas de dodecaedro/icosaedro con rectangulos aureos); SOLO cuando el core de algebra este solido. Anotado, sin trabajo actual. Vision del dominio (Anto, 17:04): darle lo MINIMO en geometria - axiomas + juez exacto - para que descubra por si mismo, como en algebra. Hito sonado (Anto, 17:05): que AlphaZero descubra POR SI SOLO phi en el dodecaedro (razon diagonal/arista = phi, verificable en Q(raiz5)) | 0% | 0% |

## Datos medidos (12 oct)

- Benchmark bucle MCTS TicTacToe (200 partidas x 100 sims, mediana):
  Go 2,93M sims/s; C 5,51M (1,88x); Rust 5,74M (1,96x). Detalle: RESULTADOS_BENCH.md.
- Juez Rust (h17): 230/230 vectores, 0 discrepancias (canonEq + BFS nodos exactos).
  Velocidad pendiente de medir en maquina quieta.
- Conjunto ciego (12 oct, `./mathzero curve 3 60 1000`, cap 1000):
  CON todo: 330 -> 382 -> 450/700 (ciclos 1-3; nodos_med 587/512/416; prueba_med 2,4/2,5/2,7).
  Ciclo 3: k2:100 k3:95 k4:81 k5:59 k6:47 k7:45 k8:23.
  Ablaciones: ciclo 1 SIN red 305, SIN biblio 551; ciclo 3 SIN red 300, SIN biblio 526.
  Hallazgo (repetido): los lemas como acciones crudas HUNDEN la busqueda (duplican
  branching); la red aporta +150/700 sobre busqueda uniforme. Ver RESULTADOS_MEJORAS.md.
- Juez (12 oct ~17:15, maquina quieta, 230 vectores, Rust 0 discrepancias):
  canon: Go 284.572 juicios/s, Rust 1.182.826 (4,2x). BFS: Go 5.160 nodos/s, Rust 5.472 (1,06x).
- CI (GitHub Actions): job rust-c verde; job go tras fix de net.go (gofmt), en verificacion.
- Fijo (levels.json, no ciego; historico): learn 12 rondas, cap 2000:
  k4 26/30, k5 25/30, k6 19/30, k7 16/30, k8 16/30 (BFS: 16, 15, 6, 5, 9).
  Ablacion biblioteca en ciclo 3 (fijo): 146 vs 146 (efecto marginal una vez la red
  interioriza los lemas; reportado con honestidad).

## Experimentos de la literatura (A/B: mismo conjunto ciego y cap, 3 semillas)

Se ejecutan por impacto/coste; se mantiene solo lo que mejore con cifras.
Resultados: RESULTADOS_MEJORAS.md. Estado: pendiente = no empezado.

| # | Experimento | Fuente | Coste | Estado |
|---|-------------|--------|-------|--------|
| E8 | Transferencia k2-k5 -> k6-k8 (entrenar solo bajos, medir altos) | propio | bajo | pendiente (prioridad 1) |
| E2 | Entrenamiento online tras cada lote | propio | bajo | pendiente (prioridad 1) |
| E4 | Replay priorizado + ponderacion por longitud de prueba | arXiv 2006.00283 | bajo | pendiente (prioridad 1) |
| E3 | Curriculo "al borde" 20-80% (tipo STP) vs niveles k fijos, midiendo olvido | arXiv 2502.00212 | medio | pendiente |
| E5 | Datos sinteticos por reescritura hacia delante como preentrenamiento | AlphaGeometry (Nature) | medio | pendiente |
| E6 | Biblioteca estilo DreamCoder (macros frecuentes) | DreamCoder | medio | pendiente |
| E1 | Red dentro del MCTS/PUCT (la red como prior en el arbol) | HyperTree, arXiv 2205.11491 | alto | pendiente |
| E7 | Test-time RL ligero para k7-k8 | AlphaProof | alto | pendiente |

## Eje de calidad profesional (Q)

| # | Criterio | Verificable como | Estado |
|---|----------|------------------|--------|
| Q1 | Tests automaticos | `go test ./...` verde (canon, reglas, 230 vectores, TTT, MLP, ciego sin solape) | hecho local, pendiente push |
| Q2 | CI en el repo | .github/workflows/ci.yml: gofmt + vet + test + build CLI/WASM + juez Rust vs vectores + bench C | escrito, pendiente push |
| Q3 | Reproducibilidad | semillas fijadas visibles en cada comando; un comando por tabla | parcial |
| Q4 | Arquitectura limpia | interfaz Game, modulos separados, sin duplicados | parcial |
| Q5 | Rendimiento medido | RESULTADOS_BENCH.md + bench/ repetible | hecho |
| Q6 | Documentacion "Guia para bobos" | pasos numerados, un diagrama simple, sin humo | pendiente |
| Q7 | CLI estable | subcomandos documentados y con tests de humo | pendiente |
| Q8 | API local documentada | endpoints + ejemplos curl | pendiente |
| Q9 | Empaquetado en un ejecutable | un binario sin dependencias | parcial (el CLI ya es un binario) |
| Q10 | Metricas, logs y evaluacion rigurosa | logs por corrida + ciego + varias semillas con intervalos | parcial |

## Prioridad inmediata (orden de trabajo)

1. Terminar curva ciega (h16) y velocidad del juez Go/Rust en maquina quieta (h17).
2. Push de todo lo local (codigo, tests, CI, ROADMAP, bench, blind.json, vectores).
3. E8, E2, E4 (baratos) -> RESULTADOS_MEJORAS.md.
4. Analisis de fallos k6-k8 (evalfail) -> guia de E3/E5/E6.
5. Panel de capacidades (h18) como comando unico con 3 semillas.
6. E3, E5, E6; luego E1, E7. Despues: laberinto (h9), sudoku (h10), descubrimiento (h15).
