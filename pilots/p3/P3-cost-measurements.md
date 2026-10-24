# P3: mediciones de coste (entorno limpio: un proceso, ningun otro proceso mio en marcha)
Maquina de ~1.98 GB de RAM, 2 CPUs. Go 1.26.10. Sin biblioteca salvo que se indique. No son resultados de aprendizaje.

## Evaluacion del banco fijo levels.json (210 casos), red model_math.json (incluida en el repo; nivel 6, ronda 12), cap 100
Salida original: L2 30/30 (0.10 s), L3 30/30 (0.13 s), L4 29/30 (2.12 s), L5 29/30 (5.11 s), L6 24/30 (12.73 s), L7 19/30 (23.61 s), L8 18/30 (20.57 s). TOTAL 179/210, ~64.3 s, RSS maximo 424 MiB.
(CORRECCION: un mensaje previo decia 149/210 por un error de suma mio; la suma de la salida original es 179.)
Intentos censurados (sin resultado): evaluacion de los 210 casos con cap 300 cortada a 55 s (RSS 940 MiB) y evaluacion monolitica a cap 100 cortada a 55 s antes de terminar; no dan tiempo.

## Autojuego de entrenamiento, nivel 6, 20 problemas (RNG seed 1), red model_math.json, cap 100
Sin biblioteca: 19 resueltos, 83 muestras, busqueda 4.48 s, SGD (3 epocas) 0.57 s, RSS maximo 197 MiB.
Con biblioteca (lemas.json): NO termino; cortado a 40 s con RSS 751 MiB. Cota inferior del tiempo de la fase de busqueda: >=~40 s frente a 4.48 s sin biblioteca (mismo conjunto de problemas por la misma semilla, misma red, cap y nivel; la ruta de busqueda difiere por la biblioteca), es decir al menos ~9x como cota inferior sobre la busqueda; no se da cociente exacto. Misma configuracion con cap 300 y biblioteca: el kernel mato el proceso por falta de memoria (dmesg, ver abajo).

## Curriculo desde red nueva NewMLP(1), cap 100, 6 rondas x 40 problemas
Niveles por ronda 2..7; resueltos 40,40,40,40,39,28 de 40; tiempo de busqueda por ronda 0.13, 0.17, 0.27, 2.28, 6.38, 19.13 s; SGD 0.40, 0.54, 0.73, 3.08, 2.83, 1.64 s; RSS maximo 346 MiB. Luego la red resuelve 30/30 en L2, L3 y L4 del banco fijo. (Semilla 1: no se usa en la validacion.)

## Memoria
~3 MiB por nodo en casos duros con biblioteca (s303-L8-2: Sys 599/794/856 MiB a cap 200/250/280). El Cycle heredado (cap 2000 por defecto, biblioteca activa) no cabe en 2 GB.

## Evidencia del sistema sobre kills (dmesg, PID enlazado a los procesos mios)
Cuatro kernel OOM-kills (global_oom, anon-rss ~1.83 GB): PID 8968 (mz, banco seed 202 con biblioteca), PID 9000 (mz, segundo intento con biblioteca seed 303 en segundo plano), PID 10369 (mz, seed 303 con biblioteca sin GOMEMLIMIT), PID 11382 (p3.test, entrenamiento cap 300 con biblioteca). Solo estos cuatro estan enlazados por PID; no se generaliza a otros SIGKILL.

## Guardia (scripts/guard.sh)
Aclaracion: la corrida que sufrio el OOM de 1.83 GB (PID 11382) usaba un supervisor ANTERIOR que solo medía el RSS pico y NO cortaba por memoria; el limite de 1200 MiB lo introduje despues. Guardia actual: sondea el RSS del grupo de procesos cada 0.1 s, mata el grupo al superar el limite o el tiempo, informa residual_procs. Pruebas reales: reserva de memoria a ~150 MiB/s con limite 300 MiB -> cortada a 308 MiB de pico, 0 procesos residuales; tope de tiempo de 3 s sobre un sleep 60 -> cortada, 0 residuales; proceso normal -> status=exit. Limite operativo para este entorno: 900 MiB de RSS, un proceso a la vez. No se promete pico <0.5 GB: depende de pesos/banco/cap y se mide por corrida.
