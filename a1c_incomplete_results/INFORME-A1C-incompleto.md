# A1C: confirmacion A-F con semillas 19-26 y banco G3. Ejecucion incompleta (0 de 8 pares)

Resumen: la ejecucion se detuvo pronto, tras el primer timeout de etapa y antes de completar ningun par; no hay pares A-F completos, ni p-valor, ni conclusion sobre adaptar la guia (A) frente a congelarla (F). Resultado primario: INCOMPLETO, no confirmatorio, segun el protocolo (pilots/a1c/PROTOCOL-A1C.md).

## Procedencia
- Fuente: rama exp/mathzero-next, commit bd7055a61969124b1a60280a92027a2481c9e763. Go 1.26.6 linux/amd64. Binario sha256 a3eec5f8f0c1e7b567aa266f30428eb0b7e4cd0d63528ddb29c0eef89b8188ed.
- Banco G3: semilla 20261027, 90 casos (niveles 6,7,8 x 30), digest 9df061e0f89aae8736ffda6e8b562d90dede192d0d9e2f928440a3f50bab7a19 (el binario lo exige).
- Configuracion congelada: semillas 19..26, calendario 2..7, 40 problemas por ronda, caps 100/100. Limites: 1 CPU, 1 GiB de RSS vigilado por sondeo sobre el arbol de procesos, 150 s por etapa, 5400 s en total, etapas secuenciales, sin red.
- Inicio 2026-10-26 17:52:35 CET; detencion 18:00:14-18:00:19 CET.

## Que se ejecuto (estados reales)
| semilla | etapa | estado | segundos observados | pico RSS |
|---|---|---|---|---|
| 19 | pre-g3 | CENSORED_TIMEOUT (limite 150 s) | 270 (reloj de la guardia) | 444 MiB |
| 19 | resto | no ejecutado (el protocolo no continua una semilla con una etapa no completada) | - | - |
| 20 | pre-g3 | COMPLETED | 130 (129.4 s internos) | 497 MiB |
| 20 | train-A | detenido a ~53 s por decision tecnica | 53 | no registrado |
| 20 | resto | no ejecutado | - | - |
| 21-26 | todas | no iniciadas | - | - |
La etiqueta CENSORED_SIGNAL que el sistema escribio en seed 20 train-A es automatica: la causa real es la detencion deliberada, no un timeout ni una etapa completada. No se reintento nada y no se cambio ningun parametro.

## Tiempos: observado frente a reservado
- Observado: semilla 19 pre-g3 270 s; semilla 20 pre-g3 130 s y train-A 53 s (183 s). Total 453 s; reloj ~464 s.
- Reservado en la contabilidad del protocolo (budget_used_s): semilla 19 = 150 s; semilla 20 = 280 s (130 usados + 150 reservados para train-A); suma 430 s.
- La reserva no mide tiempo real. Semilla 19: 150 reservados frente a 270 observados, subcuenta de 120 s. Semilla 20: 280 reservados frente a 183 observados, sobrecuenta de 97 s. El neto (430 frente a 453) no borra ninguno de los dos efectos.
- Incidencia: la etapa de la semilla 19 tenia limite de 150 s y la guardia registro 270 s de reloj antes de cerrarla. Por el codigo, la guardia sondea cada 0.1 s y corta a 150 s, con un timeout interno a 155 s (y kill a 158 s). No hay evidencia de que esas acciones se ejecutaran de forma efectiva antes de los 270 s; la causa es desconocida y no se afirma ningun mecanismo ni hipotesis sobre el entorno de ejecucion.
- Referencias de tiempo, sin comparacion controlada: en el protocolo A1 se midieron antes 38.6 s (etapa post-A de a1-val-17 en G2, ejecucion historica) y 67 s (reproduccion posterior de esa misma evaluacion con 1 CPU fijada). Son medidas de otro banco y de otras ejecuciones; la etapa pre-g3 de la semilla 20 (130 s) no es una comparacion controlada con ellas. Solo se constata que el tope de 150 s por etapa dejaba poco margen frente a 130 s.

## Dato completo disponible (solo descriptivo)
Semilla 20, evaluacion inicial en G3 (red NewMLP(20), sin entrenar), cap 100: 36 de 90 resueltos (nivel 6: 17/30, nivel 7: 10/30, nivel 8: 9/30), 7124 nodos en total. Hash de pesos 9772059848b3a04f8ef38b04caff8a3e459b68aa641d4f3aed809dabbcef3a9b, huella de configuracion de la etapa c085fa6117798a55897b455f6e29cbba9394f98a7b38fab2eb1ff5ca78f59601, fichero seed20_pre-g3.json sha256 26066f956b828c786feee09dbe2c0cacf94ed4f0dad64b38bd194c4832332f51. Es una sola semilla, una sola etapa previa al entrenamiento: no compara A con F y no sustenta ninguna inferencia.

## Que no se puede concluir
Nada sobre A frente a F: no hay entrenamiento completado ni evaluaciones post. Tampoco sobre aprendizaje general ni utilidad en una interfaz real. El tope de 150 s por etapa y la censura posible dependiente de la red o del banco (no ignorable) habrian sesgado cualquier subconjunto de pares que se hubiera completado.

## Verificacion
La procedencia de la ejecucion es la reportada por quien la realizo y no esta verificada de forma independiente; los estados y datos de este informe fueron cotejados con los ficheros de la ejecucion.
Se conservan los ficheros originales de la ejecucion sin modificar. Este paquete incluye solo este informe, estados_a1c.csv y seed20_pre-g3.json, con su manifiesto.
