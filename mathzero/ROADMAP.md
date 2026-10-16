# ROADMAP (privado) - actualizado 12 oct 2026 21:50

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
  MCTS en Rust: PRIORIDAD APARCADA (decision de Anto 22:31) - va detras de lemas/negaciones;
  se reabre solo si las medidas muestran que la busqueda es el cuello de botella.
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
- Checkpoints (higiene, aprobado 00:04): ciclo.go guarda model_math.json.best cuando la metrica fija mejora y model_math.json.nivel<N> al superar nivel. No toca motor ni eval.
- CI (GitHub Actions): go y rust-c verdes (runs 30-40). Causa del fallo 'vet fantasma':
  directorios espurios mathzero/go/mathzero/go y thzero (pushes con ruta en el nombre);
  borrados, con go/go raiz incluido.
- A/B noche 12 oct (ciego, cap 1000, emparejado): E8 descartado (media +2,7, ruido);
  E2 descartado (media -23,3); E4 descartado (media 0,0/700 en 6 semillas, ruido).
- Ejecutable autonomo (datos embebidos + flag -json), Release v0.1.1: PROBADO EN WINDOWS
  REAL por Anto (12 oct 23:24): arranque sin panic, 136 reglas OK, BFS 16/30 y 7/30
  identicos a Linux. (v0.1.0 daba panic sin argumentos por problems.json no embebido ->
  corregido embebiendolo.) Servidor: tanda 1 con wrapper de reloj de pared.
- Fijo (levels.json, no ciego; historico): learn 12 rondas, cap 2000:
  k4 26/30, k5 25/30, k6 19/30, k7 16/30, k8 16/30 (BFS: 16, 15, 6, 5, 9).
  Ablacion biblioteca en ciclo 3 (fijo): 146 vs 146 (efecto marginal una vez la red
  interioriza los lemas; reportado con honestidad).
- Geometria (16-oct, revision): el soporte vive solo en la rama `research/geometry-bridge`
  (no en main). Compila y los tests pasan. Las 4 identidades del banco son validas por el
  juez exacto (4/4) pero la busqueda NO prueba ninguna (0/4; prior uniforme, cap 2000-20000).
  Causa raiz: los fixtures contienen neg(const) y las 14 reglas no lo pliegan. Ver E15.
- E9 enviado al pipeline real (16-oct): `curve`/`cycle` activan SOLO los lemas con uso
  medido (top-8; `Lemma.Uses` ahora es persistente y se incrementa en GuidedSolve). Con el
  modelo/biblioteca commiteados, `curve 1 0 1000` pasa de 336 a 554/700 en el ciego y -54%
  de tiempo. Detalle en RESULTADOS_MEJORAS.md.

## Experimentos de la literatura (A/B: mismo conjunto ciego y cap, 3 semillas)

Se ejecutan por impacto/coste; se mantiene solo lo que mejore con cifras.
Resultados: RESULTADOS_MEJORAS.md. Estado: pendiente = no empezado. Dominio nuevo de evaluacion (logica booleana u otro con juez exacto): PROPUESTO, sin aprobar.

| # | Experimento | Fuente | Coste | Estado |
|---|-------------|--------|-------|--------|
| E8 | Transferencia k2-k5 -> k6-k8 (entrenar solo bajos, medir altos) | propio | bajo | DESCARTADO (medido: media +2,7/700, dentro del ruido; RESULTADOS_MEJORAS.md) |
| E2 | Entrenamiento online tras cada lote | propio | bajo | DESCARTADO (medido: media -23/700; 2 de 3 semillas muy negativas) |
| E4 | Replay priorizado + ponderacion por longitud de prueba | arXiv 2006.00283 | bajo | DESCARTADO (medido: media 0,0/700 en 6 semillas, ruido puro; RESULTADOS_MEJORAS.md) |
| E9 | Poda de lemas-accion por uso medido (top-8) vs biblioteca cruda | propio (hallazgo ablacion) | bajo | ACEPTADO (13-oct 00:45): +12,3/700 media, 3/3 semillas +12/+16/+9, k6-8 +4,7, nodos -5,3%. ENVIADO al pipeline real (16-oct): con el modelo/biblioteca commiteados, `curve 1 0 1000` pasa de 336 a 554/700 (los 94 lemas commiteados tenian Uses=0 -> E9 no activa ninguno) |
| E10 | Datos sinteticos neg-neg/add0 (25% del stream) contra el punto ciego de limpieza | propio (evalfail) | bajo | ACEPTADO (13-oct 01:07): +10,3/700 media, 3/3 semillas +6/+8/+17, k6-8 +3,3, nodos -3,7% |
| E11 | Motor con direccion en MAction + dedup de sucesores (bug medido: 30% acciones duplicadas, 1.044 r enmascaradas) | revision externa + direction_diag_test.go | bajo | ACEPTADO (13-oct 06:20): +25,0/700 media, 3/3 semillas +32/+21/+22, k6-8 +17,7, nodos -7,7% |
| E12 | Ablacion factorial 2x2 (red x biblioteca) multisemilla, eval-only | revision externa | bajo | PROPUESTO, sin aprobar |
| E13 | Puerta de aceptacion por set fijo (levels.json): quedarse con la red nueva solo si no baja (arena alpha-zero-general traducida a matematicas) | suragnair/alpha-zero-general | bajo | EN CURSO (00:15): puerta implementada (smoke verificado); A/B 3 semillas en cadena local tras E11 |
| E3 | Curriculo "al borde" 20-80% (tipo STP) vs niveles k fijos, midiendo olvido | arXiv 2502.00212 | medio | pendiente |
| E5 | Datos sinteticos por reescritura hacia delante como preentrenamiento | AlphaGeometry (Nature) | medio | pendiente |
| E6 | Biblioteca estilo DreamCoder (macros frecuentes) | DreamCoder | medio | pendiente |
| E1 | Red dentro del MCTS/PUCT (la red como prior en el arbol) | HyperTree, arXiv 2205.11491 | alto | pendiente |
| E7 | Test-time RL ligero para k7-k8 | AlphaProof | alto | pendiente |
| E14 | Senal de aprendizaje tipo RL sobre la iteracion de expertos: aprovechar los problemas fallidos (relabeling) y/o una cabeza de valor para ordenar la busqueda; relacionada con E1 y E7 | hipotesis propia | medio | CANDIDATO (13-oct): sin aprobar ni medir. Solo tras E13 y el checkpoint completo. Un unico cambio, A/B desde cero, mismo ciego de 700, semillas pareadas; el criterio de aceptacion se fija y justifica antes de medir |
| E15 | Axioma neg(const)->const (plegado de constantes negadas) para que el motor alcance las identidades que hoy no reduce; medido en el banco de geometria (hoy 0/4 por busqueda) y en el ciego de algebra | propio (revision geometria 16-oct) | alto | CANDIDATO (16-oct): sin aprobar ni medir. Cambia el numero de reglas -> cambia NFeat -> invalida model_math.json: exige reentrenar. A/B 3 semillas en el ciego de 700 + comprobar el banco de geometria. Criterio de aceptacion a fijar antes de medir |

## Eje de calidad profesional (Q)

| # | Criterio | Verificable como | Estado |
|---|----------|------------------|--------|
| Q1 | Tests automaticos | `go test ./...` verde (canon, reglas, 230 vectores, TTT, MLP, ciego sin solape) | hecho (verde en CI) |
| Q2 | CI en el repo | .github/workflows/ci.yml: gofmt + vet + test + build CLI/WASM + juez Rust vs vectores + bench C | hecho (runs 30-40 verdes) |
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

## Vision a largo plazo (direccion/idea de Anto, 13-oct ~00:15-00:20)

**Estado: direccion/idea. NO medido, NO aprobado como trabajo.** Cualquier experimento concreto necesitara aprobacion explicita y medicion en ciego como el resto.

- Que AlphaZero, partiendo de reglas basicas + juez exacto, desarrolle sobre Tensor Logic (Domingos, arXiv 2510.12269) como dominio nuevo con juez exacto.
- Y, en general, sobre matematicas, algebra y geometria (AlphaGeometry/AlphaProof como referencias de direccion, no como plan).
- Tensor Logic tambien se valora como NOTACION para describir el diseno del proyecto (reglas, juez, busqueda, aprendizaje) con mas claridad.

## Capa conversacional (vision a largo plazo; direccion decidida por Anto via main, 13-oct 05:55)

**Estado: vision futura, DESPUES de TODOS los hitos actuales de este ROADMAP. NO es el siguiente paso ni trabajo actual.** Anto dejo la eleccion a main ("lo dejo a tu eleccion", WhatsApp 05:55) y main eligio: opcion 4 empezando por la 1. Anto matizo despues (06:02): en el ROADMAP hay muchas otras cosas antes.

- Un LLM encima traduce lenguaje natural a problemas formales; AlphaZero razona; el juez verifica cada paso (sin alucinaciones: cada afirmacion pasa por el verificador).
- Antes quedan todos los hitos actuales (motor, experimentos de la tabla, eje de calidad). Cuando algun dia toque, entrara como propuesta con aprobacion explicita y medicion en ciego como todo lo demas.
