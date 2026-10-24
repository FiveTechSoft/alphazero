# DISENO A1 (congelado ANTES de generar el banco G2 y antes de cualquier prueba con redes ni medicion de A1)
Protocolo NUEVO, independiente de P3 y de G1 (no reutiliza sus resultados como objetivo ni validacion). Arbol mathzero-next (base 62c089d + inc1..14), rama experimental.

## Pregunta (una sola intervencion)
Con problemas, calendario de niveles, presupuesto de busqueda, red inicial del alumno, optimizador y banco iguales: cambia el resultado segun que la GUIA de la busqueda de entrenamiento se ADAPTE durante el entrenamiento (brazo A) o sea una copia CONGELADA de la red inicial (brazo F)?
- Efecto TOTAL de la intervencion, incluidos sus mediadores (problemas resueltos, numero de muestras = pasos de SGD, longitud de caminos, hindsight conservado/descartado, orden de las muestras). No iguala computo ni muestras; no separa autojuego de SGD; no es "red vs sin red".
- En ambos brazos hay busqueda; solo cambia la red que la guia. Fuente de etiquetas del brazo F: busqueda guiada por la copia congelada, independiente de los pesos que se actualizan.

## Brazos
- A: la red que aprende es tambien la guia (se actualiza al final de cada ronda; igual que el bucle de P3/G1 pero con calendario fijo).
- F: guia = clon profundo de NewMLP(semilla) congelado (nunca entrena, nunca comparte memoria con el alumno; se comprueba su hash antes y despues). El alumno = NewMLP(semilla), se entrena con las muestras que genera la busqueda guiada por el clon.
- Ambos: sin library, lr 0.003, 3 epocas por ronda, TrainCE, checkpoint final unico, cap de entrenamiento 100, cap de evaluacion 100, 6 rondas x 40 problemas = 240.

## Calendario y problemas
- Calendario FIJO de niveles por ronda: 2,3,4,5,6,7 (eleccion de diseno informada por experiencia previa con P3/G1; NO se afirma que sea independiente de ella). Independiente de resultados de A1; no se cambia por resultados.
- RNG de PROBLEMAS: fuente rand propia sembrada con la semilla, que solo consume genTrain y el rechazo por clave; no depende de la busqueda, del brazo ni del cap. Por tanto la lista de 240 problemas es IDENTICA en A y F (se verifica con test sobre la lista exacta, incluso variando el cap).
- RNG de SGD: una fuente por (semilla, ronda), igual para ambos brazos; con muestras identicas (ronda 1) el orden y los pesos resultantes son identicos; con muestras distintas el orden difiere de forma natural. No consume el generador de problemas.
- Exclusion en entrenamiento: se rechaza y regenera un problema cuyo Start o Target (Key) este en el conjunto del banco G2 (hasta 1000 intentos por problema; si se agotan, la semilla FALLA, sin truncar ni relajar). Hindsight: se descarta el par si Key(Start) o Key(objetivo sustituto real) esta en el conjunto G2 o no se puede parsear. Identidad = Key (asociativa-conmutativa estructural; NO equivalencia algebraica).

## Banco G2 (regla fijada)
- Semilla 20261025 (fijada por main antes de generar). Niveles 6,7,8 x 30 = 90. Misma regla que G1 (genTrain por nivel; aceptar solo si Key(a)!=Key(b), ninguna de las dos en PRIOR, y ninguna repetida en el banco en ningun rol; limite 200*30+1000 = 7000 intentos por nivel; FALLA CERRADO si no llega a 30, sin buscar otra semilla ni relajar).
- PRIOR = PRIOR de G1 (levels.json, blind.json niveles 2..8, banco secundario P3 semilla 9001) + Key de Start y Target de los 90 casos del banco G1 congelado (digest 8bc4db332ab0e6c240b0b9e788a091f282a695f31b777301d1d1bcc7962b209e, comprobado).
- Se registran intentos/aceptados/rechazos por nivel. Digest BankDigest.

## Evaluacion y criterio
- Pre-evaluacion de NewMLP(semilla) sobre G2: UNA sola, compartida por ambos brazos (la red inicial es identica). Post-evaluacion por brazo sobre G2. delta_X = resueltos post - pre (90 casos, cap 100). Sin evaluar bancos antiguos.
- Semillas de validacion 11..18 (el piloto de una semilla, 101, es aparte y fuera del criterio).
- Criterio: d = delta_A - delta_F por semilla. Test de signos exacto bilateral estandar sobre d: ceros excluidos, n efectivo = P+N explicito, p = min(1, 2*P(X<=min(P,N))), X~Binomial(n efectivo,1/2); NOMINAL, condicional al banco fijo y configuracion. Direccion por la mayoria: P>N favorece adaptacion (A>F), N>P favorece guia fija; P=N sin direccion; n efectivo 0 = NO INFORMATIVO. Se reporta siempre P/N/Z por semilla integros y delta_A, delta_F por semilla. (Misma convencion que la Adenda 1 de G1.)
- Si el brazo F genera pocas muestras, eso es resultado/mediador de la intervencion; no motiva redisenar segun resultados.

## Limites declarados
Efecto total de la guia adaptativa frente a una guia inicial aleatoria congelada (no frente a una guia fija entrenada); no atribuye autojuego vs SGD ni compara igual computo; disjuncion modulo AC estructural, no generalizacion matematica; una configuracion; n=8; coste real desconocido hasta un piloto autorizado.
