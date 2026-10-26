# Reproduccion de una evaluacion historica: A1 val-17 (brazo A), post-A, banco G2

Alcance: reproduccion de compatibilidad de evidencia ya publicada. NO es evidencia confirmatoria nueva, ni comparacion A-F, ni prueba TUI/OpenCode, ni aprendizaje en otra area. El banco G2 ya estaba expuesto.

## Protocolo y banco
A1 congelado (pilots/a1/DESIGN-A1-frozen.md). Semilla 17, calendario 2..7, 40/ronda, cap entrenamiento 100, cap evaluacion 100. Banco G2: semilla 20261025, niveles 6,7,8 x 30 = 90, digest f58161c5133bffeb2838ee9b5da9a516f09022bbe7d94367c0b4c658d4bb67b6 (el codigo se niega a correr si no coincide). Fingerprint del protocolo 272905109635807ba2e595a499a3c6cb61d0ce56c91441cf5065ca643808953a (igual en original y reproduccion).

## Codigo y entorno
Rama exp/mathzero-next, commit 081a407a88aba67e85f80f1eab6ce074e74a4b3c. Compilado con Go 1.26.6 linux/amd64 (el original uso Go 1.26.10; binarios distintos, no se afirma igualdad binaria). sha256 binario reproduccion: b2560eb2ff56810f30ea6bcc52fc5da6ad27ea40e406138f50fd599e0b8c2b82; binario original: 97060cbd94f93e4a6422afee0aecb6d635c9e95ac4e7cf09b3c3a7c809216624.
Maquina: 2 vCPU disponibles, ejecucion fijada a 1 CPU, GOMAXPROCS=1, GOMEMLIMIT=900MiB (aviso: GOMEMLIMIT no es limite duro de RSS), timeout 300 s.

## Comando
mathzero d1a1train -seed 17 -perround 40 -cap 100 -evalcap 100 -stage post-A -dir <dir con train-A.json y weights_A.json del paquete publicado>
Una sola ejecucion, sin reintentos.

## Medicion
rc 0; tiempo pared 67.4 s (stderr: 66.9 s; el original registro 38.6 s); CPU usuario 65.5 s, sys 1.9 s; RSS maximo observado 407 MiB. Es una observacion, no una garantia. El tiempo varia entre ejecuciones; 67 s no es una cota para otros checkpoints o casos.

## Integridad de pesos (lectura de codigo, a1_main.go RunA1StageWithBank, etapa post-)
La etapa post-A: lee train-A.json, carga weights_A.json, exige que WeightsSHA(net) == hash registrado por train-A (si no, error), evalua con EvalBank, vuelve a comprobar el hash (error si "la evaluacion modifico los pesos") y solo escribe post-A.json. No entrena ni guarda pesos. Comprobacion externa: sha256 de weights_A.json antes y despues de la ejecucion = ab33795077136eadc83aa1b74e1bae8af6779277c61cc32e96935fc8b087a0e1 (el archivo "despues" esta en weights_A.sha256_after; el "antes" coincide con el publicado en OPENCODE.md y con train-A.json weights_sha256 bb9955ca... del contenido del modelo).

## Resultado
90 casos: 65 resueltos en original y en reproduccion; los 90 pares (id, solved, nodes) son identicos (comparison_90_pairs.tsv). post-A.reproduced.json sha256 742cdc96243e65c198b6e456298b342d057ce894a385f0cc347522761904ff41 (incluye el campo de tiempo, que difiere del original).

## Limites
Un checkpoint, una semilla. val-17 es la semilla con mayor delta_A de las 8 (A-F total: P4 N4, p=1.0 en el paquete A1 original). No se probo main a4bba9d ni cbd325db. No se probo d1serve/OpenCode. Los 90 casos pueden compartir plantillas; la exclusion por Key no cubre equivalencia algebraica.
