# P3: diseno CONGELADO antes de entrenar (aprendizaje medido, red nueva, sin biblioteca)
Congelado antes de ejecutar ningun entrenamiento del piloto ni de la validacion. Los resultados no pueden cambiar estos parametros dentro de esta version; un cambio exige una version nueva con motivo.
Codigo: exp/mathzero-next (tras inc7): TrainRounds (d1_train.go), EvalBank (d1_experiment_adapter.go), GenBank/BankDigest (bank_gen.go).

## Pregunta
Tras R rondas de autojuego bidireccional + TrainCE, resuelve la red mas casos de un banco de evaluacion no usado para entrenar, con la misma busqueda y presupuesto antes y despues?

## Configuracion fija
- Red inicial de cada semilla: NewMLP(seed) (nivel 2, ronda 0). Sin pesos del repo.
- Sin biblioteca (LemmaLib=nil en entrenamiento y evaluacion).
- Entrenamiento: 6 rondas x 40 problemas por ronda, cap de busqueda 100, lr 0.003, 3 epocas con barajado, subida de nivel si >=80% resueltos en la ronda (curriculo identico al de Cycle), problemas genTrain a nivel de la red, RNG sembrado con la semilla de entrenamiento. Hindsight: muestras de hindsight incluidas salvo que el par (render, ambas orientaciones) este en el banco de evaluacion.
- Evaluacion: EvalBank, cap 100, sin entrenamiento, antes (R=0, red NewMLP(seed)) y despues (R=6, unico checkpoint final). NINGUNA seleccion de checkpoints por rendimiento en el banco de evaluacion.
- R=6 y 40 por ronda se eligen por coste medido (curriculo desde red nueva: 6 rondas x 40 = ~38 s y RSS 346 MiB en una semilla distinta de las de validacion) y NO se ajustan para que suba el resultado.
- Un solo proceso a la vez; supervisor externo (scripts/guard.sh) con limite de RSS 900 MiB y tope de tiempo; un tope alcanzado es CENSURA con estado explicito, nunca fallo matematico.

## Bancos de evaluacion (hashes con BankDigest, algoritmo en bank_digest_algo)
- PRIMARIO: levels.json niveles 6,7,8 (90 casos), digest 1bd3a2dad9f4482e026ee85df8f137356bdf81123bbbcd5cd06364bfe94af86d. Fichero levels.json SHA256 c8ea037a7e87dea3da88... (completo en MANIFEST de la base).
- SECUNDARIO: GenBank seed 9001, niveles 6,7,8 x30 (90 casos), digest 8909948adae8f2b17e3151c943a59215197adf4aea787f11185eaedc3d2ee6f2.
- Separacion: los pares exactos de evaluacion se excluyen del entrenamiento (normalizacion Key, AC, ambas orientaciones). Esto NO garantiza separacion de Start ni de Target: se registran por semilla los solapes entrenamiento-evaluacion de Start, Target y par (OverlapTrainEval, normalizacion Key). Medido en los bancos: los 180 casos de evaluacion tienen 118 Start distintos (los casos comparten Start) y 180 Target distintos; 30 Start del banco secundario tambien estan en el primario. Los bancos NO son muestras independientes.

## Semillas y fases
- PILOTO (mide coste y comprueba que el codigo funciona; no entra en la validacion): semilla 101.
- VALIDACION: semillas 11..18 (8 semillas). No se usan las semillas 1..8 porque la medicion de coste ya vio el entrenamiento de la semilla 1.

## Criterio y analisis (declarados antes)
- Metrica por semilla: delta = resueltos(post) - resueltos(pre) en el banco PRIMARIO.
- Unico criterio de senal: numero de semillas con delta>0 de 8, test del signo exacto bilateral (p nominal), mas media de delta. Unidad = semilla.
- Descriptivos (no criterio): banco secundario (misma metrica), delta por nivel 6/7/8, transiciones por caso (resuelto->no, no->resuelto), nodos, agotamientos, nivel final y resueltos por ronda del curriculo, solapes entrenamiento-evaluacion.
- Resultado = exploratorio; los p son nominales y condicionales al banco fijo compartido (Start repetidos); analizar por semilla no elimina la dependencia del banco.

## Limites declarados
- No separa autojuego de SGD adicional: no hay brazo control "mismo numero de pasos de SGD sin autojuego". El control sin entrenar repetido prueba solo determinismo.
- Una sola configuracion (R, tamano, cap, lr); sin barrido. Sin biblioteca. Generalizacion limitada a estos bancos de niveles 6-8 y a la funcion de busqueda heredada.
- Niveles 2-4 se omiten del criterio porque una red nueva ya los resuelve casi completos (efecto suelo/techo).
- Sin equivalencia con Reduced8 ni aval de la campana historica.
