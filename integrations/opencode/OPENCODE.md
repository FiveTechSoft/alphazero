# Usar AlphaZero (MathZero) desde OpenCode como modelo

Integracion equivalente a la de FiveTechSoft/symbols (ver su `opencode.json` y
`opencode.md`): OpenCode trata a AlphaZero como MODELO via un servidor local
OpenAI-compatible. AlphaZero NO es un modelo de lenguaje. Lo que responde es
determinista y limitado:

| Entrada | Respuesta | Origen |
|---|---|---|
| Saludo / "quien eres" | Plantilla fija de identidad | Ruta determinista declarada (patron Symbols: `ServerIsGreeting`/`ServerAnswerGreeting`); NO la genera la red |
| `<expresion> -> <objetivo>` (notacion prefija: add/mul/neg, constantes cN, variables va-vz) | Camino de reescritura con reglas, pasos y nodos | Busqueda guiada por la politica neuronal entrenada (SolvePath) |
| Cualquier otra cosa | Abstencion fija (fuera de alcance) | Plantilla declarada |

Limites: la igualdad "resuelto" es modulo la normalizacion asociativa-conmutativa
del motor (misma regla que en las evaluaciones); sin herramientas de OpenCode
(`tool_call: false`); sin conversacion general; sin garantia de cobertura fuera
de la distribucion de entrenamiento.

## Build y arranque

```bash
cd mathzero/go && go build -o mathzero .
cp ../../integrations/opencode/guard.sh .   # la guardia debe estar junto al binario (o flag -guard)
./mathzero d1serve -port 8100 -model <pesos.json> -cap 400
```

`-model` es un fichero de pesos MLP. Checkpoint de demo ya publico: en el
paquete persistido `pilots/a1/validation-11-18/`, directorio
`a1-val-17/weights_A.json` (sha256
ab33795077136eadc83aa1b74e1bae8af6779277c61cc32e96935fc8b087a0e1), de la
validacion A1. Es un checkpoint de demo, NO un modelo seleccionado como mejor.
Sin `-model` usa una red nueva semilla 1 (sin entrenar; solo para pruebas de
conduccion). El puerto 8100 evita colision con symbols-server (8099).

Flags de recursos (v3): `-cap` nodos por resolucion (defecto 200, maximo 500,
se reduce con aviso), `-rssmb` limite RSS supervisado del worker en MiB
(defecto 384), `-wtimeout` segundos maximos por worker (defecto 20). Requiere
`guard.sh` (por defecto junto al binario; ruta alternativa con `-guard`) y
bash: cada resolucion corre en un SUBPROCESO worker bajo la guardia.
Build tag `!js && linux`: solo Linux.

## Configuracion de OpenCode

El `opencode.json` de este directorio declara el proveedor `alphazero`
(`@ai-sdk/openai-compatible`, baseURL `http://127.0.0.1:8100/v1`) y el modelo
por defecto `alphazero/alphazero`, siguiendo la seccion "Custom provider" de
https://opencode.ai/docs/providers/ . Probado con opencode-ai 1.19.4
(`opencode run ...`). Comprueba la version que uses: el esquema puede cambiar.

## Modelo de recursos y limites reales (v3, 2026-10-25)

La v2 (motor en-proceso con mutex) fue rechazada en revision: entradas
admitidas (depth 60, cap 400) medidas a ~950 MiB de RSS en 2-3 s. Race verde
no implica uso seguro. La v3 aplica aislamiento de procesos:

- SUBPROCESO por resolucion: el servidor no ejecuta el motor en su propio
  proceso. Lanza `bash guard.sh <wtimeout> <rssmb> <out> <exe> d1solve ...`
  con grupo de procesos propio (Setpgid). Sin estado compartido: los
  globales del motor (Key cache, SearchTraining) viven y mueren en el worker.
- guard.sh: COPIA SIN MODIFICAR de la guardia ya probada en las campanas
  (sha256 da23456b580e70060bfbb191e73a03c57469e857bc862174b87b7cd42e7ad4d1).
  Supervisa RSS real del worker por muestreo (cada 0.1 s) y mata el grupo al
  superar el limite; aplica el timeout. Es supervision por muestreo: un pico
  de memoria muy rapido puede exceder brevemente el limite antes de ser
  detectado; NO es un limite duro garantizado, ni setrlimit, ni un timeout
  HTTP actua como limite de memoria.
- Cancelacion: si el cliente corta la conexion (o vence el plazo), el
  servidor envia TERM y luego KILL al GRUPO del worker; el worker nunca
  queda huerfano. Limpieza tambien en error y en salida del servidor.
- Admision acotada: 1 worker en paralelo + cola de 4; el exceso espera
  hasta 10 s y recibe HTTP 503 con Retry-After. Una peticion cancelada
  mientras espera en cola NUNCA arranca su worker (verificado en test).
- Entrada: cuerpo POST <= 256 KiB (413), mensaje de usuario <= 4096 bytes,
  expresiones <= 200 bytes y <= 16 niveles de anidamiento (400),
  cap <= 500 nodos. Conservadores a proposito tras la medicion de la v2.
- Un worker que muere (OOM, kill, error) devuelve HTTP 500/504 a SU
  peticion; el servidor y el resto de peticiones siguen vivos
  (verificado en test con hog sintetico).
- Timeouts HTTP: ReadHeader 5 s, Read 15 s, Write 90 s, Idle 30 s.
- Escucha solo en 127.0.0.1: limita la exposicion, NO es validacion.

## Mediciones (Linux, 2026-10-25)

- RSS pico del worker via guard.sh: problema normal (nivel 4, cap 400)
  25 MiB. La entrada mas costosa medida (depth 16, 200 B, cap 500) llego a
  164 MiB, sin demostrar que sea la peor admisible; 384 MiB es un margen
  operativo, no una cota.
- Hog sintetico (`d1hog -mb 128`): muerto por la guardia al superar 96 MiB
  (limite de prueba), HTTP 500, servidor vivo, residual 0.
- Worker lento (`d1hog -sleep 30` con wtimeout 2 s): HTTP 504, proceso
  recolectado, residual 0.
- Tras rafaga de 12 resoluciones concurrentes en vivo: 12 x HTTP 200,
  0 procesos worker/guard residuales.
- Verificado SOLO en Linux. Windows NO probado: guard.sh es bash y el
  build tag excluye no-Linux; no se afirma portabilidad.

## Pruebas

`cd mathzero/go && go test -race -run D1Serve .` (11 tests): concurrencia
16x solve con 503 acotado, mezcla de rutas, 413, expresion larga/anidada,
parseo sin `<nil>`, rutas, stream, /v1/models, hog matado por RSS sin
tumbar el servidor, worker lento a 504, cancelacion en cola sin arranque
de worker. Los tests localizan guard.sh por D1S_GUARD, por el layout del
repo (integrations/opencode/guard.sh) o junto al directorio de trabajo.

## Verificacion rapida

```bash
curl http://127.0.0.1:8100/v1/models
curl -X POST http://127.0.0.1:8100/v1/chat/completions -H 'Content-Type: application/json' \
  -d '{"model":"alphazero","messages":[{"role":"user","content":"add(va,c0) -> va"}]}'
```
