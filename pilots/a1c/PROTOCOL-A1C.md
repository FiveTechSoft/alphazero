# A1C: confirmacion A-F con semillas y banco nuevos (protocolo, congelar antes de ejecutar)

## Pregunta
Con problemas de entrenamiento, calendario, presupuesto de busqueda, red inicial, optimizador y banco iguales, adaptar la guia durante el entrenamiento (brazo A) frente a una guia congelada, copia de la red inicial (brazo F), cambia el numero de problemas resueltos en un banco nuevo? Es el efecto total de la intervencion (incluye problemas resueltos, muestras, hindsight y orden). No iguala computo, no separa autojuego de SGD, no mide aprender en otras areas.

## Cronologia fijada antes de generar o entrenar
1. Semilla del banco G3: 20261027, fijada antes de generarlo (declaracion previa). Semillas de entrenamiento: 19..26, nuevas para A1/MathZero: no aparecen en el repositorio ni en los registros de A1/P3/G1. Los numeros 21 y 22 se usaron como semillas en otros experimentos distintos (transformers de dialogo, otro codigo y dominio), sin relacion con NewMLP ni con A1.
2. Generacion de G3 y digest (sin modelos, sin resultados). Se ejecuto 3 veces en la preparacion, las 3 con el mismo digest candidato 9df061e0f89aae8736ffda6e8b562d90dede192d0d9e2f928440a3f50bab7a19; el digest es candidato hasta su congelacion. 3. Congelar el digest tras revision independiente. 4. Entrenar A y F por semilla. 5. Evaluar pre/post sobre G3. 6. Analisis.

## Diseno
- Entrenamiento A1 sin cambios (pilots/a1/DESIGN-A1-frozen.md): calendario 2..7, 40 problemas por ronda, cap 100 en entrenamiento y evaluacion, lr 0.003, 3 epocas por ronda, checkpoint final unico, guia F congelada con hash comprobado antes y despues.
- Banco G3: 90 casos, niveles 6, 7, 8 x 30, mismo generador y reglas que G2. Excluye por Key (normalizacion estructural asociativa-conmutativa, no equivalencia algebraica) todo lo de blind, levels, G1, G2 (casos y prior) y P3-9001; sin repeticiones dentro del banco. Falla cerrado si no llena.
- Los problemas de entrenamiento y los objetivos hindsight que tengan Start o Target (Key) en el banco se rechazan o descartan en el propio entrenamiento. La lista de problemas depende solo de la semilla y del banco, no del brazo; los objetivos hindsight dependen de la trayectoria y se filtran al generarse.
- Sin seleccion de checkpoints ni de semillas.

## Analisis primario
Unidad = semilla (n=8 pares A-F). d_s = resueltos_A - resueltos_F en G3. Prueba de signos exacta bilateral, ceros excluidos, n efectivo = P+N, p = min(1, 2*P(X<=min(P,N))), X~Binomial(n efectivo, 1/2), alfa 0.05. Significativos: 8:0 p=0.0078, 7:0 p=0.0156 (un cero), 6:0 p=0.0313 (dos ceros). No significativos: 7:1 p=0.0703, 6:2 p=0.289, 5:0 p=0.0625, 4:0 p=0.125. Se reportan siempre P/N/Z, d por semilla, media y rango. Secundario: prueba exacta de permutacion por cambio de signo (2^8). Los casos del banco no son unidades independientes: ningun test por problema ni agrupado.
Si falta o se censura cualquier par A-F previsto, el resultado primario es incompleto y no confirmatorio: no se reduce n, no se rellena, no se reintenta; se reporta lo completado con el posible sesgo.
Lectura: sin ventaja detectada no significa equivalencia ni ausencia de efecto. Potencia baja.

## Ejecucion (plan, sin ajustar despues)
1 CPU, 1 etapa a la vez, seed a seed (pre-g3, train-A, train-F, post-A, post-F), tope por etapa 150 s, tope total 90 min (5400 s: suma de budget_used_s de todas las semillas mas reloj de pared), vigilancia del RSS de todo el arbol de procesos con guard.sh en 1024 MiB (guardia por sondeo cada 0.1 s: puede sobrepasarse entre sondeos; GOMEMLIMIT no es limite). 40 etapas x 150 s serian 100 min: el tope de 90 min es duro y, si se agota, falta algun par y el resultado primario es INCOMPLETO y no confirmatorio; no se recorta n, no se reintenta, no se reparametriza. La censura puede depender del brazo o de la dificultad y por eso no se imputa. Medidas previas (no garantia): entrenamiento ~30 s por brazo, evaluacion de 90 casos 38.6 s y 67 s en dos maquinas/ejecuciones distintas. Script: scripts/a1c_run.sh (BANK=g3), invocado siempre con bash (los scripts del repositorio no tienen bit de ejecucion). Pruebas de las guardas con un binario sustituto, sin entrenamiento: scripts/a1c_run_test.sh y scripts/guard_test.sh.

## Comprobaciones del analizador (todas implementadas en a1_confirm.go)
Por semilla: los 5 status_* existen y dicen COMPLETED; los 5 json existen y llevan su propio nombre de etapa; una sola huella (fingerprint) igual a la ESPERADA para esa semilla y la configuracion G3, recomputada con los parametros congelados (semilla, calendario 2..7, 40 por ronda, caps 100/100, digest G3 y conjunto de claves del banco leido de pilots/a1c/g3bank.json tras verificar su digest) y no solo unica; los ids de resultados son exactamente los 90 casos de G3 en orden y los post coinciden con el pre; el hash de pesos post es el de train por brazo; el hash de la guia F antes y despues es el de la red inicial (pre-g3) y el brazo A no lleva hash de guia; los 240 problemas de entrenamiento son identicos en A y F; hay binary.sha256 y budget_used_s. Entre semillas: huella o red inicial repetida (directorio reutilizado), mas de un binario, o presupuesto total > 5400 s invalidan. No verificado por el analizador: el contenido de la huella (semilla, digest del banco, claves), que cada etapa del binario ya comprueba al leer la etapa previa.

## Limites
Condicional a G3 y a esta configuracion (una configuracion, 8 semillas, un banco de 90 casos que pueden compartir plantillas). Mismos niveles y generador no garantizan igual dificultad o distribucion realizada; excluir Keys previas cambia el muestreo respecto a G2. Disjuncion por Key estructural, no equivalencia algebraica. Algebra MathZero, caps 100, rama exp/mathzero-next. No es evidencia de aprendizaje autonomo general ni de utilidad en OpenCode (TUI es otra etapa).

## Implementacion
a1_main.go (-bank g2|g3, etapa pre-g3 para G3; G2 por defecto sin cambios), g3_bank.go (G3Seed, G3PriorKeys, comando d1g3bank que solo genera), a1_confirm.go (comando d1a1confirm -bank pilots/a1c/g3bank.json -root R, solo lectura de JSON), pruebas artificiales a1_confirm_test.go. Wrapper scripts/pilot_a1_stage.sh con BANK=g3.
