# RESULTADOS DE MEJORA (privado) - solo medido, conjunto ciego salvo que se diga

Reglas: A/B con mismo conjunto ciego (blind.json, 700) y mismo cap; 3 semillas
cuando se comparen variantes. Se mantiene lo que mejora con cifras; se descarta
lo que no, y se dice.

## Hallazgo 1 (12 oct 16:26): la biblioteca como acciones extra HUNDE la busqueda

Ablacion del ciclo 1 de `curve` (ciego, cap 1000, red ronda 12):

| Variante | Total | k2 | k3 | k4 | k5 | k6 | k7 | k8 | nodos_med |
|----------|-------|----|----|----|----|----|----|----|-----------|
| CON todo (94 lemas) | 330/700 | 95 | 63 | 55 | 38 | 39 | 26 | 14 | 587 |
| SIN red (con lemas) | 305/700 | 100 | 63 | 46 | 38 | 28 | 18 | 12 | 642 |
| SIN biblioteca (con red) | 551/700 | 100 | 98 | 96 | 86 | 67 | 66 | 38 | 294 |

Mecanismo medido: los lemas ~duplican el branching (raiz k3: 31 vs 15 acciones;
k7: 25 vs 13). Con cap bajo la busqueda llega a la mitad de profundidad.
Contradice la ablacion vieja en conjunto fijo (146 vs 146, cap 2000): alli era
neutral. Conclusion: la biblioteca NO se puede enchufar como acciones crudas;
necesita integracion con poda/puntuacion (candidata: E6 DreamCoder con filtro,
o lemas solo en la raiz). Estado: confirmado en 1 ciclo; pendiente repetir en
ciclos 2-3 y con 3 semillas antes de cambiar diseno.

## Diseno A/B (E2, E4, E8) - en ejecucion

Comando: `./mathzero ab <e2|e4|e8> <semilla> [rondas=8] [por_ronda=60] [cap=1000]`.
Dos brazos desde la MISMA semilla: init de red identico y mismo stream de
problemas de entrenamiento (emparejado). Evaluacion en el mismo conjunto ciego
(700 problemas), mismo cap, SIN biblioteca en los brazos (ver Hallazgo 1).
E2 variante: entrenamiento online tras cada problema. E4 variante: replay
priorizado por longitud de prueba. E8 variante: techo de nivel k5 en
entrenamiento (transferencia a k6-k8). Resultados: ab_e2.log, ab_e4.log,
ab_e8.log. Estado: corriendo en cadena tras la curva (chain.log).

## E8 - Transferencia k2-k5 -> k6-k8 (completo, 3 semillas)

Variante: entrenar SOLO con k2-k5 (techo k5), medir en el mismo ciego k2-k8.
Emparejado: misma semilla => misma init de red y mismo stream de problemas.

| Semilla | Base total | Var total | Dif | Base k6-8 | Var k6-8 | Dif k6-8 |
|---------|-----------|-----------|-----|-----------|----------|----------|
| 1 | 529 | 544 | +15 | 158 | 164 | +6 |
| 2 | 536 | 539 | +3 | 161 | 163 | +2 |
| 3 | 537 | 527 | -10 | 164 | 160 | -4 |
| media | 534,0 | 536,7 | +2,7 | 161,0 | 162,3 | +1,3 |

Veredicto: DESCARTADO. La media (+2,7/700) esta dentro del ruido entre
semillas (+-10); sin mejora robusta en k6-k8. Entrenar solo niveles bajos
no transfiere mejor que el curriculo completo.
Comando: `./mathzero ab e8 <semilla>` (log: go/ab_e8.log).

## E2 - Entrenamiento online tras cada problema (completo, 3 semillas)

Variante: entrenar tras CADA problema resuelto (3 epocas) en lugar de
acumular por ronda. Emparejado por semilla. Mismo ciego, cap 1000, sin biblioteca.

| Semilla | Base | Var | Dif | Base k6-8 | Var k6-8 |
|---------|------|-----|-----|-----------|----------|
| 1 | 529 | 470 | -59 | 158 | 132 |
| 2 | 536 | 518 | -18 | 161 | 158 |
| 3 | 537 | 544 | +7 | 164 | 165 |
| media | 534,0 | 510,7 | -23,3 | 161,0 | 151,7 |

Veredicto: DESCARTADO. El online empeora la media en -23/700 (2 de 3
semillas muy negativas): la red deriva hacia lo reciente. Ademas cuesta
~4x por semilla. Comando: `./mathzero ab e2 <semilla>` (log: go/ab_e2.log).

## E4 - Replay priorizado por longitud de prueba: DESCARTADO (6 semillas)

Variante: replay con peso = longitud de prueba (arXiv 2006.00283), tope 20000.
Mismo binario y diseno emparejado; semillas 1-3 en local, 4-6 en el servidor
(mismas rondas/cap; la inicializacion depende de la semilla, asi que el
baseline de cada semilla es su propio control).

| Semilla | Origen | Base | Var | Dif | Base k6-8 | Var k6-8 | Dif k6-8 |
|---------|--------|------|-----|-----|-----------|----------|----------|
| 1 | local | 529 | 543 | +14 | 158 | 166 | +8 |
| 2 | local | 536 | 527 | -9 | 161 | 159 | -2 |
| 3 | local | 537 | 537 | 0 | 164 | 163 | -1 |
| 4 | servidor | 549 | 553 | +4 | 166 | 171 | +5 |
| 5 | servidor | 532 | 529 | -3 | 164 | 156 | -8 |
| 6 | servidor | 551 | 545 | -6 | 170 | 163 | -7 |

Media 6 semillas: 0,0/700 (k6-8: -0,8). Ruido puro: el +14 de la semilla 1
era espejismo. Veredicto: DESCARTADO, no se incorpora.
Comando: `./mathzero ab e4 <semilla>` (logs: go/ab_e4.log, tanda 1 servidor).


## E9 - Poda de lemas por uso medido (top-8): ACEPTADO (3 semillas, 13 oct 00:45)

Variante: biblioteca podada a los 8 lemas con uso medido en el propio
entrenamiento (de 111), misma red y stream emparejados por semilla. Ambos
brazos CON biblioteca (cruda vs podada); mismo ciego (700), cap 1000.

| Semilla | Base (cruda) | Var (podada) | Dif | Base k6-8 | Var k6-8 | Dif k6-8 | Base nodos | Var nodos |
|---------|--------------|--------------|-----|-----------|----------|----------|------------|-----------|
| 1 | 519 | 531 | +12 | 152 | 157 | +5 | 324 | 304 |
| 2 | 500 | 516 | +16 | 140 | 147 | +7 | 346 | 328 |
| 3 | 515 | 524 | +9 | 149 | 151 | +2 | 332 | 317 |
| media | 511,3 | 523,7 | +12,3 | 147,0 | 151,7 | +4,7 | 334,0 | 316,3 |

Veredicto: ACEPTADO. 3/3 semillas positivas (+12/+16/+9), media +12,3/700,
k6-8 +4,7 y -5,3% de nodos. La poda por uso convierte la biblioteca de
lastre (Hallazgo 1) en mejora neta sobre la cruda. Lo NO medido aqui:
podada vs SIN biblioteca en diseno emparejado (Hallazgo 1 era otra red y
otro diseno; no comparable directamente). Propuesto como parte de E12
(factorial), NO aprobado aun.
Comando: `./mathzero ab e9 <semilla>` (log: go/ab_e9.log).

## E10 - Datos sinteticos neg-neg/add0 (25% del stream): ACEPTADO (3 semillas, 13 oct 01:07)

Variante: 25% de problemas sinteticos dirigidos al punto ciego de limpieza
(neg-neg, add0) en el stream de entrenamiento. Emparejado por semilla;
ambos brazos SIN biblioteca; mismo ciego (700), cap 1000.

| Semilla | Base | Var | Dif | Base k6-8 | Var k6-8 | Dif k6-8 | Base nodos | Var nodos |
|---------|------|-----|-----|-----------|----------|----------|------------|-----------|
| 1 | 529 | 535 | +6 | 158 | 159 | +1 | 306 | 297 |
| 2 | 536 | 544 | +8 | 161 | 165 | +4 | 295 | 286 |
| 3 | 537 | 554 | +17 | 164 | 169 | +5 | 297 | 282 |
| media | 534,0 | 544,3 | +10,3 | 161,0 | 164,3 | +3,3 | 299,3 | 288,3 |

Veredicto: ACEPTADO. 3/3 semillas positivas (+6/+8/+17), media +10,3/700,
k6-8 +3,3 y -3,7% de nodos. Nota: brazos SIN biblioteca, no comparable con
las cifras de E9 (ambos brazos CON biblioteca). E12 (factorial E9xE10)
mediria la interaccion: PROPUESTO, NO aprobado.
Comando: `./mathzero ab e10 <semilla>` (log: go/ab_e10.log).
