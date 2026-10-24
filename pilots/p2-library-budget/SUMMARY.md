# P2 piloto de diagnostico: biblioteca x presupuesto (mathzero-next)
Estado: diagnostico auxiliar, NO evidencia de aprendizaje, NO verificacion independiente. Codigo: exp/mathzero-next (bd00677 + inc6, ver PROVENANCE-inc6). Evaluacion pura (sin entrenamiento, sin pre/post).

## Configuracion
- Presupuesto (cap): 300 nodos. Brazos: sin biblioteca / con lemas.json.
- Pesos: model_math.json, SHA256 1466ee8be8f22e69f1457a8070d49e66cf6c931da92fd21633c2a04f51e0e516. Incluidos en el repo; origen de entrenamiento desconocido.
- Biblioteca: lemas.json, SHA256 5ce68ec3fdee10d7cea6ea71578f16e84b8fa2fac04a5756b0722054b50b4c3c (94 lemas, incluida en el repo, no regenerada).
- Bancos (GenBank, regla en bank_gen.go): seed 101 niveles 4,5,6 x4 (12 casos), hash runner cc88d99960e164b51d1b158b568260be1396c2f8238853e25137e91066fac97f; seed 303 niveles 6,8 x3 (6 casos), hash runner 6573afc1e97af9d7cd525585adefebcc7a5a2756bf944329dbc9d9402cde7bbc.
- Comando: mathzero d1eval -bankseed S -perlevel N -levels L -cap 300 -model model_math.json [-lib lemas.json]. Toolchain Go 1.26.10 linux/amd64.

## Resultados pareados completos (ambos brazos terminaron, mismo banco y pesos)
| banco | casos | ambos resueltos | solo sin biblioteca | solo con biblioteca | ninguno | nodos sin / con |
|---|---|---|---|---|---|---|
| seed 101 | 12 | 11 | 1 (s101-L5-0: 75 nodos vs agota 300) | 0 | 0 | 192 / 464 |
| seed 303 | 6 | 4 | 2 (s303-L8-1: 98 vs 300; s303-L8-2: 6 vs 300) | 0 | 0 | 360 / 850 |
Total 18 casos: 0 mejoras con biblioteca, 3 regresiones, 15 empates resueltos por ambos. Sin p-valor ni inferencia global: N pequeno, bancos no independientes entre si por construccion de Start.
Detalle por caso: paired_results.json. Salidas crudas del runner: raw/.

## Censurado / no pareado
- seed 202 (niveles 4-8 x6, 30 casos, hash runner 19c51b5e44d6730771527748175efd064978c2109f13a23671645887155945e1): brazo sin biblioteca completo (25/30, 5 agotados, 2,183 nodos; raw/seed202_none_unpaired.json). Brazo con biblioteca: proceso terminado por SIGKILL a ~55 s, causa DESCONOCIDA (sin evidencia del sistema). No entra en ninguna comparacion.

## Coste (solo mediciones de entorno limpio: ningun otro proceso mio en marcha)
Repetidas en entorno limpio: seed 101 ambos brazos (sin biblioteca 594 ms/192 nodos; con biblioteca 8,183 ms/464 nodos); seed 303 sin biblioteca 5,587 ms/360 nodos; seed 303 con biblioteca 34,876 ms/850 nodos (con GOMEMLIMIT=1100MiB, limite blando); perfil de 1 caso (s303-L8-2, con biblioteca, un proceso): cap 60/120/200/250/280 = 1.38/2.70/4.7-5.2/7.1/7.3 s; memoria Sys maxima del runtime 599 MiB (cap 200), 794 (250), 856 (280).
INVALIDADOS como tiempos aislados: cualquier tiempo de las corridas en las que dos procesos huerfanos mios compartian la maquina (intentos con corte a 50-115 s sobre seed 303 con biblioteca, incluida la afirmacion ">27 s para L8-2"). No se usan.
Extrapolacion previa de 6 ms/nodo y 4,800 s: refutada. La maquina tiene ~2 GB totales.
Hotspot (perfil de 1 caso, cap 200, con biblioteca, ~87% del tiempo de la goroutine de busqueda): backwardEdgeAllowed -> exactAction (enumera todas las acciones del hijo y renderiza cada sucesor). GC ~38% de muestras totales. No se adopto ninguna optimizacion; cualquiera exige equivalencia con configuracion identica (resueltos, nodos por caso y camino).

## Solapes entre semillas (GenBank, seeds 1..8, niveles 4-8 x20 = 800 casos)
497 Start distintos, 793 pares (Start,Target) distintos, 7 pares compartidos entre semillas, 0 en levels.json/blind.json. Excluir pares entre semillas NO independiza los Start. Si se agrupa por semilla hay que tratar el Start como bloque explicitamente.

## Limites
Sin entrenamiento ni pre/post. La red y la biblioteca no fueron producidas por este arbol. Casos de ejecucion unica (determinista) sin repeticion estadistica. Sin equivalencia con Reduced8.
