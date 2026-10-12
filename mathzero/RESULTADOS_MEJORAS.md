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
