# A1C tanda B: ejecucion completa (8 de 8 pares A-F). Resultado primario: sin ventaja detectada

Resumen: las 8 semillas completaron sus 5 etapas (40/40 COMPLETED), sin censuras, sin reintentos y sin cambios de parametros durante la ejecucion. La guia que se adapta durante el entrenamiento (A) gano en 3 semillas y perdio en 5 frente a la guia congelada (F); diferencia media d = solved_A - solved_F = -0.5 problemas de 90. Sign test bilateral exacto p = 0.7265625; permutacion de signos p = 0.9140625. Veredicto del analizador congelado: no_advantage_detected. Lectura: en este banco y esta configuracion no se detecta ventaja de adaptar la guia; con n = 8 la potencia es baja y esto NO es equivalencia ni evidencia de ausencia de efecto.

## Pregunta y brazos (de pilots/a1c/PROTOCOL-A1C.md, sin cambios)
Con problemas de entrenamiento, calendario, presupuesto de busqueda, red inicial, optimizador y banco iguales: brazo A = la guia se adapta DURANTE el entrenamiento; brazo F = guia congelada, copia de la red inicial. Unidad de analisis: par A-F por semilla (n = 8). d = problemas del banco G3 resueltos por A menos los resueltos por F tras entrenar.

## Procedencia y congelacion
- Rama exp/mathzero-next, freeze de ejecucion commit e51d39c286585670b3414d7d797c76a3bd66ffa3 (scripts y protocolo B; fuente cientifica sin cambios desde bd7055a61969124b1a60280a92027a2481c9e763).
- Binario sha256 a3eec5f8f0c1e7b567aa266f30428eb0b7e4cd0d63528ddb29c0eef89b8188ed (un solo binario en las 8 semillas).
- Banco G3: fichero pilots/a1c/g3bank.json sha256 13af4f77f1baaea76244bb17350c77ddbb2a39d0bc7b619d8f58c42b8537ad2a; digest interno del banco 9df061e0f89aae8736ffda6e8b562d90dede192d0d9e2f928440a3f50bab7a19 (el que el binario exige). Son magnitudes distintas: hash del fichero frente a digest del contenido del banco. 90 casos, niveles 6, 7 y 8 x 30, semilla de generacion 20261027.
- Configuracion: semillas 19..26, calendario 2..7, 40 problemas por ronda, caps 100/100, 1 CPU, sin red.
- Limites del protocolo B (pilots/a1c/PROTOCOL-A1C-B.md): 300 s por etapa, 7200 s operativos en total, plazo civil absoluto de 14400 s, RSS 1024 MiB vigilado por sondeo.

## Ejecucion (estados reales)
- 8/8 pares A-F completos; 40/40 etapas COMPLETED, cada una con guardia status=exit rc=0 y 0 procesos residuales. Ninguna censura, ningun reintento, ninguna modificacion de parametros.
- Inicio 2026-10-26 21:22:32 UTC (22:22:32 CET); fin civil 22:22:11 UTC (23:22:11 CET).
- Tiempo operativo usado 3575 s de 7200 (reloj de pared operativo 3580 s); civil 3579 s de 14400. Por semilla (budget_used_s): 19=415, 20=526, 21=438, 22=422, 23=401, 24=470, 25=470, 26=433.
- Ninguna etapa supero los 139 s (tope 300 s). Pico de RSS por etapa entre 322 y 563 MiB (tope 1024).
- La tanda A (ejecucion incompleta, a1c_incomplete_results/) se conservo aparte e intacta; nada de ella se mezclo ni se reutilizo.

## Resultados por semilla
pre = problemas de G3 resueltos por la red inicial, antes de entrenar (cap 100). Tabla completa en a1c_B_per_seed.csv.

| semilla | pre | solved_A | solved_F | d = A - F |
|---|---|---|---|---|
| 19 | 40 | 61 | 56 | +5 |
| 20 | 36 | 47 | 56 | -9 |
| 21 | 34 | 70 | 55 | +15 |
| 22 | 38 | 69 | 54 | +15 |
| 23 | 38 | 60 | 65 | -5 |
| 24 | 35 | 50 | 68 | -18 |
| 25 | 38 | 54 | 58 | -4 |
| 26 | 37 | 56 | 59 | -3 |

Totales: A = 467, F = 471.

## Analisis primario (analizador congelado; salida completa en confirm.json)
- status PRIMARY_COMPLETE; n efectivo 8; P = 3, N = 5, Z = 0.
- media d = -0.5; sign test bilateral exacto p = 0.7265625; permutacion de signos p = 0.9140625.
- verdict: no_advantage_detected. Nota del propio analizador: nominal, condicionado al banco G3 y a esta configuracion; baja potencia; no es equivalencia; no es aprendizaje general; no es TUI.

## Que se puede y que no se puede concluir
- Se puede decir: en estas 8 repeticiones, con este banco y esta configuracion, adaptar la guia durante el entrenamiento no mostro ventaja sobre dejarla congelada (diferencia media -0.5 problemas de 90; p nominal 0.73).
- No es equivalencia: con n = 8 la prueba tiene poca potencia; no demuestra ausencia de efecto.
- No es una confirmacion ciega independiente: las semillas 19 y 20 y el banco G3 quedaron parcialmente expuestos en la tanda A (reanudacion tecnica declarada en pilots/a1c/PROTOCOL-A1C-B.md).
- El contraste primario es A frente a F; no es un test de si el entrenamiento mejora respecto a no entrenar.
- No es evidencia de aprendizaje general, de auto-mejora, ni de utilidad en una interfaz real (TUI es otra etapa).

## Verificacion y reproducibilidad
- El analisis se repitio de forma independiente: el analizador se recompilo desde el arbol congelado y se ejecuto sobre los datos crudos de la tanda; la salida JSON fue identica a confirm.json. Los estados, presupuestos, huellas del binario y la ausencia de procesos residuales se cotejaron contra los ficheros crudos de la ejecucion.
- Para repetir el analisis sin entrenar, con el binario congelado: d1a1confirm -root <raiz de la tanda B> -seeds 19,20,21,22,23,24,25,26 -budget 7200 -bank pilots/a1c/g3bank.json (invocacion congelada en pilots/a1c/ANALYSIS-B-INVOCATION.txt).
- Los datos crudos de la ejecucion (estados, guardas, registros) se conservan fuera del repositorio publico. Este paquete incluye solo este informe, confirm.json y a1c_B_per_seed.csv, con su manifiesto (MANIFEST.sha256).
