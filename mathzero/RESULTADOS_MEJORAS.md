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

## E9 enviado al pipeline real (16 oct): la biblioteca cruda penalizaba +218/700

E9 deja de vivir solo en `ab`: los comandos shipped (`curve`, `cycle`) usan
SOLO los lemas con uso medido (`Uses>=1`, top-8) para la metrica y la busqueda;
el autojuego ve los candidatos CRUDOS para medir uso y anadir candidatos. El
contador de uso de cada lema ahora es PERSISTENTE (`Lemma.Uses` se incrementa en
GuidedSolve y se salva/carga en `lemas.json`).

Medido con el modelo y la biblioteca commiteados, `curve 1 0 1000` (ciego, 700):

| Config | Total | k6 | k7 | k8 | nodos_med | tiempo |
|--------|-------|----|----|----|-----------|--------|
| CON todo crudo (94 lemas) [antes] | 336/700 | 39 | 26 | 14 | 580 | 767s |
| CON todo E9 (0 activos*) [despues] | 554/700 | 67 | 66 | 38 | 283 | 265s |

*Los 94 candidatos commiteados tienen Uses=0 (el contador nunca se incrementaba);
E9 no activa ninguno, luego la metrica coincide con SIN biblioteca. Ganancia
+218/700 y -54% de tiempo. Lo NO medido: si el top-8 de lemas que SI se usan
(tras acumular uso con `cycle`) bate a SIN biblioteca en diseno emparejado; E9
se acepto contra la biblioteca cruda (+12,3/700), no contra el vacio.

## E16 - Cierre de E9: el vacio bate al top-8 (17 oct): biblioteca OFF por defecto

Responde a lo NO medido de E9. Dos disenos, ambos emparejados (mismo net,
mismo ciego, solo cambia `LemmaLib`):

1) `cycle 2 150 1000` (ciego de niveles, 210), pares en el mismo ciclo:
   ciclo 1 CON biblio (0 lemas) = 158/210 == SIN biblio = 158/210;
   ciclo 2 CON biblio (8 lemas) = 135/210 vs SIN biblio = **161/210** (-26, -12%).

2) `ab e9 12345` (ciego canonico, 700), mismo net, brazo VACIO anadido:

| Brazo (mismo net) | Total ciego |
|-------------------|-------------|
| cruda (94 lemas) | 520/700 |
| top-8 (8 lemas) | 526/700 |
| vacio (sin lemas) | **539/700** |

Veredicto: la biblioteca NO bate al vacio: vacio > top-8 > cruda. Aunque el
juez verifica cada lema, inyectarlos como acciones amplia el branching y
consume el cap antes del objetivo. E9 ayudo porque pasa de CRUDA a top-8
(menos lastre), no porque la biblioteca aporte. Decision (E16): el pipeline
shipped (`curve`, `cycle`) NO usa la biblioteca; el top-8 queda como ablacion
en los comandos y en `ab e9`. Confirmado sin regresion: `curve 1 0 1000` da
554/700 (= previo SIN biblioteca).


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

## E11 - Motor con direccion en MAction + dedup de sucesores: ACEPTADO (3 semillas, 13 oct 06:20)

Variante: MAction.Dir + DirFix global + dedup por clave de sucesor completo
(corrige el bug medido: 30,06% de acciones duplicadas, 1.046 acciones-r
mal aplicadas; assoc/comm absorbidas gratis por la Key canonica). Emparejado
por semilla; ambos brazos SIN biblioteca; mismo ciego (700), cap 1000.

| Semilla | Base | Var (DirFix) | Dif | Base k6-8 | Var k6-8 | Dif k6-8 | Base nodos | Var nodos |
|---------|------|--------------|-----|-----------|----------|----------|------------|-----------|
| 1 | 529 | 561 | +32 | 158 | 183 | +25 | 306 | 285 |
| 2 | 536 | 557 | +21 | 161 | 178 | +17 | 295 | 271 |
| 3 | 537 | 559 | +22 | 164 | 175 | +11 | 297 | 273 |
| media | 534,0 | 559,0 | +25,0 | 161,0 | 178,7 | +17,7 | 299,3 | 276,3 |

Veredicto: ACEPTADO. 3/3 semillas muy positivas (+32/+21/+22), media
+25,0/700, k6-8 +17,7 y -7,7% de nodos. Mayor mejora medida hasta ahora:
la direccion en el motor elimina trabajo duplicado y descubre las acciones-r
que estaban enmascaradas. Validado en tests: modo clasico sin cambios,
DirFix 0,00% duplicadas (direction_diag_test.go).
Comando: `./mathzero ab e11 <semilla>` (log: go/ab_e11.log).

## Relajacion de costes en GuidedSolve (26 oct): DESCARTADO (1 semilla)

Idea (rama `fix/guidedsolve-best-cost`): sustituir el `seen` booleano por un
mapa `bestCost` y reabrir un estado solo si se alcanza con un coste
estrictamente menor. Se implemento sobre `GuidedSolve` y `SolvePath`,
conservando el contador persistente `Lemma.Uses` de E9.

Medicion (`./mathzero evalblind 1000`, mismo modelo nivel 6 ronda 12,
ciego 700, sin lemas):

| Variante | Total | k3 | k7 | nodos_med |
|----------|-------|----|----|-----------|
| `seen` actual | **554/700** | 99 | 66 | 283 |
| `bestCost` | 552/700 | 98 | 65 | 283 |

Veredicto: DESCARTADO. No mejora el ciego y pierde 2 problemas; el cambio
no se incorpora. La rama queda como referencia. El comando `evalblind` si
se mantiene: permite repetir esta comparacion sin entrenar ni escribir
`model_math.json`/`lemas.json`.
