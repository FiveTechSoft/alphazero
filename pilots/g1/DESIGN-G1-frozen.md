# DISENO G1 (congelado ANTES de generar el banco y antes de observar ningun resultado de G1)
Protocolo NUEVO. No es replica de P3 ni de Reduced8; no usa sus cifras como objetivo ni validacion. Arbol: mathzero-next (base 62c089d + inc1..11), rama experimental.
Pregunta: tras el protocolo de rondas P3, mejora un banco cuyos Start y Target (modulo Key) no aparecen en el entrenamiento? Disjuncion de EXPRESIONES modulo AC estructural, no generalizacion matematica.

## Identidad normalizada: Key (mathring.go)
Aplana add/mul anidados del mismo operador y ordena los argumentos (asociativa+conmutativa). neg no se colapsa; sin plegado de constantes, distributividad ni identidades. NO es equivalencia algebraica (neg(neg(x)) y x tienen Key distinto aunque igual polinomio). Se usa Key, no Canon.

## Banco G1 (regla fijada)
- Semilla del banco: 20261024 (fijada por main antes de generar). Niveles 6,7,8 x 30 = 90. Generador base genTrain(r,k), una fuente rand sembrada con la semilla, en orden de nivel.
- Candidato (a,b) aceptado solo si: Key(a)!=Key(b); Key(a) y Key(b) no pertenecen al conjunto PRIOR; Key(a) y Key(b) no han sido aceptados antes en G1 en ningun rol (cada Key aparece una sola vez en todo el banco, como Start o como Target).
- PRIOR = Key de Start y Target de todos los problemas de levels.json y blind.json (niveles 2..8) + Start y Target del banco secundario P3 (GenBank semilla 9001, niveles 6..8 x30). (El banco primario P3 son los casos L6-8 de levels.json: incluido.)
- Limite de intentos por nivel: 200*30+1000 = 7000. Si un nivel no llega a 30: FALLA CERRADO, sin relajar la regla, sin cambiar semilla ni nivel; se reportan intentos/rechazos y factibilidad.
- Se registran por nivel: intentos, aceptados, rechazos por razon (equal_key, start_prior, target_prior, repeat_in_bank). Digest: BankDigest (nivel, Start, Target en RenderPrefix, sin ids).

## Entrenamiento G1 (cuando se autorice)
- Esquema P3: NewMLP(semilla), 6 rondas x 40, cap 100, evalcap 100, lr 0.003, 3 epocas, sin library, checkpoint final unico.
- Se rechaza y regenera todo problema de train cuyo Start o Target (por Key) este en G1 (ambas orientaciones cubiertas porque se comprueban Starts y Targets de G1 como conjunto unico). Hasta 1000 intentos por problema; si se agotan: error, la semilla queda FALLIDA, sin truncar.
- Hindsight: se descarta el par si Key(Start) o Key(Target SUSTITUTO REAL, es decir, la expresion alcanzada, no el objetivo original) esta en el conjunto G1, ademas del filtro previo por par renderizado.
- Regenerar hasta 240 conserva la cantidad pero CAMBIA la distribucion por rechazo/condicionamiento; se reportan rechazos y intentos por ronda.
- Registro (solo en modo G1): rechazos, intentos, Start/Target renderizados de cada problema y pares hindsight conservados, para verificar solapes con una normalizacion INDEPENDIENTE de la que filtra.
## Criterio (G1 unico, preregistrado)
Semillas de validacion 11..18 (el piloto de una semilla 101 se mantiene separado de las 11..18). delta = resueltos post - pre sobre el banco G1 (90) con cap 100. Criterio: numero de semillas con delta>0, test de signos exacto bilateral NOMINAL, condicional al banco fijo. Bancos antiguos no se evaluan. Sin seleccion de dificultad por tasas pre/post; sin usar tasas anteriores (39-53/90) como objetivo.
## Limites declarados
Disjuncion modulo AC estructural, no matematica; sin control SGD vs autojuego (fuera de alcance); una sola configuracion; n=8; coste real del protocolo nuevo desconocido hasta un piloto autorizado.
