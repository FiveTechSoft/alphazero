# TRANSFER: evaluacion exploratoria visto/nuevo sobre las 8 redes de A1C-B. Sin diferencia detectada entre brazos

Resumen: 24/24 etapas COMPLETED (8 semillas x pre/A/F), sin censuras, sin reintentos y sin cambios durante la ejecucion. Brecha por brazo g = tasa en lista vista - tasa en lista nueva (niveles 6+7 agrupados); contraste d = g_A - g_F por semilla: P = 4, N = 4, Z = 0 (n efectivo 8), prueba de signos exacta bilateral p = 1, media d = -0.00037 (-0.04 puntos porcentuales). No se detecta diferencia entre la guia adaptativa (A) y la congelada (F) en la brecha visto/nuevo; esto NO es igualdad ni ausencia de efecto (n = 8, potencia baja).

## Pregunta y diseno (pilots/transfer/PROTOCOLO-TRANSFER-exploratorio.md)
Sobre las 8 redes ya entrenadas de A1C-B (sin reentrenar): cambia la brecha entre problemas intentados en la generacion de entrenamiento (lista "vista") y pares no entrenados del mismo generador (lista "nueva") entre el brazo A y el brazo F? Mismo solver, modo y limite (cap 100) que la evaluacion G3. Exploratoria y no confirmatoria: las semillas y el banco G3 ya estaban expuestos. Mide el paquete guia+datos, no el efecto puro de la guia. La red inicial (pre) se mide solo como control del instrumento.

## Procedencia y congelacion
- Freeze del instrumento: commit 86f7d4f518bf5c0ae7c20e730871d92e6659cb6a, rama exp/mathzero-next (hashes de codigo en pilots/transfer/HASHES-TRANSFER-v3.txt).
- Binario sha256 74f81ea071d0e6f0b20aa095f4db40ee1825c15f180bc84bbe1002869c077832 (un solo binario en las 8 semillas).
- Fuente: resultados de A1C-B (a1c_b_results/), solo lectura. Pesos pre/A/F ya entrenados; en las 24 etapas el hash de los pesos coincide con el registrado en A1C-B. Ningun entrenamiento.
- Listas por semilla: 77-80 casos por lista (631 por lista en total, niveles 6+7). Huella de cada lista identica en pre/A/F; lista vista verificada contra las listas almacenadas de train-A y train-F.
- Limites: 360 s por etapa, 900 s por semilla, 7200 s operativos, plazo civil 14400 s, 1 CPU, RSS 1024 MiB.

## Ejecucion (estados reales)
- 24/24 etapas COMPLETED, cada una con guardia status=exit rc=0 y 0 procesos residuales. Ninguna censura, ningun reintento, ninguna modificacion.
- Operativo usado 1551 s de 7200 (reloj de pared operativo 1556 s); civil 2137 s de 14400 (2026-10-26 22:46:57 UTC a 23:22:34 UTC).
- Etapa maxima 102 s (tope 360); semilla maxima 250 s (tope 900); pico RSS 548 MiB (tope 1024).

## Resultados por semilla (niveles 6+7 agrupados; valores redondeados a 4 decimales, exactos en confirm_transfer.json)
n = casos por lista. Tasas en [0,1]; g = vista - nueva; d = g_A - g_F.

| semilla | n | vista A | nueva A | vista F | nueva F | g_A | g_F | d |
|---|---|---|---|---|---|---|---|---|
| 19 | 79 | .9114 | .8481 | .8481 | .8101 | +.0633 | +.0380 | +.0253 |
| 20 | 77 | .6494 | .7662 | .7532 | .7792 | -.1169 | -.0260 | -.0909 |
| 21 | 78 | .9487 | .9359 | .8333 | .8333 | +.0128 | 0 | +.0128 |
| 22 | 78 | .9487 | .8718 | .8590 | .7692 | +.0769 | +.0897 | -.0128 |
| 23 | 80 | .8500 | .8250 | .8500 | .9125 | +.0250 | -.0625 | +.0875 |
| 24 | 79 | .7848 | .7848 | .8481 | .8608 | 0 | -.0127 | +.0127 |
| 25 | 80 | .7250 | .7375 | .8500 | .8500 | -.0125 | 0 | -.0125 |
| 26 | 80 | .8375 | .8500 | .8125 | .8000 | -.0125 | +.0125 | -.0250 |

## Analisis primario (analizador congelado; salida completa en confirm_transfer.json)
- status COMPLETE_EXPLORATORY; n efectivo 8; P = 4, N = 4, Z = 0 (n efectivo = P+N; los ceros se excluyen del test aunque las ocho semillas esten completas).
- media d = -0.00037 (-0.04 puntos porcentuales); prueba de signos exacta bilateral p = 1.

## Que se puede y que no se puede concluir
- Se puede decir: en estas 8 semillas, con este instrumento, la brecha visto/nuevo no difirio de forma detectable entre adaptar la guia y congelarla (diferencia media -0.04 puntos porcentuales; p = 1).
- No es igualdad ni ausencia de efecto, y no descarta memorizacion: con n = 8 la potencia es baja.
- No es confirmatorio ni ciego: semillas y G3 ya expuestos en A1C-B; mide el paquete guia+datos, no el efecto puro de la guia; no es causal.
- Lista "vista" = problemas INTENTADOS en la generacion de entrenamiento; no consta que entraran en el buffer de SGD. Pares "nuevos" = pares no entrenados; pueden compartir componentes con los vistos.
- Sesgo conocido: los pares nuevos son en promedio mas largos (+2 a +3.4 caracteres); efecto desconocido sobre el contraste.
- No es evidencia de aprendizaje general, de auto-mejora (RSI) ni de utilidad en una interfaz real (TUI es otra etapa).

## Verificacion y reproducibilidad
- Verificacion independiente: estados, presupuestos, hash del binario, hashes de pesos, conteos por nivel (Solved <= casos por nivel en las 24 salidas) y huellas de lista cotejados contra los datos crudos; el analisis se reprodujo con el analizador congelado y la salida fue identica a confirm_transfer.json.
- Para repetir el analisis sin ejecutar el solver, con el binario congelado: d1a1transferconfirm -root <raiz de la evaluacion> -seeds 19,20,21,22,23,24,25,26 -levels 6,7.
- Los datos crudos (estados, guardas, registros, configuracion de rutas) se conservan fuera del repositorio publico. Este paquete incluye solo este informe, confirm_transfer.json y su manifiesto (MANIFEST.sha256).
