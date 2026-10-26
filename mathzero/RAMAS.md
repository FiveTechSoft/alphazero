# Estado de las ramas (auditoria 26 oct 2026)

Auditoria de las ramas remotas frente a `main` (`cbd325d` en el momento de
la revision; despues se han anadido commits de robustez). No se modifico
ninguna rama ni se hizo push a ellas. Conclusion general: **no mergear
ninguna en bloque**; casi todas son experimentos cerrados, ya superados o
lineas de investigacion sin validacion.

| Rama | Veredicto | Motivo principal |
|------|-----------|------------------|
| `fix/guidedsolve-best-cost` | Descartada | Medida: 552/700 vs 554/700 con `seen`; sin mejora. Documentado en `RESULTADOS_MEJORAS.md`. |
| `fix/mcts-expanded-perspective` | Archivar | El fix `Expanded` ya esta en main (`36d01d2`); la API de perspectiva no tiene consumidores ni tests. |
| `fix/wasm-persisted-library` | Rechazar | Carga los 94 lemas crudos, peor brazo de E16: demo 9/12 vs 11/12 y 41,8 s vs 5,2 s. |
| `research/traince-joint-contract` | Rebase + tests | Updater conjunto prometedor, pero sin tests in-repo y con `panic` de CLI heredado; no como default. |
| `research/historical-traince-preserved` | Tag/archivar | Snapshot H reproducible; no aporta delta funcional a main. |
| `research/architecture24-utility-bidirectional` | No mergear | Regresion E9: 0/94 usos medidos; cambia el solver global y choca con E16. |
| `research/geometry-bridge` | Esperar E15 | Juez 4/4 pero busqueda 0/4 por falta del axioma `neg(const)`; no mergear aun. |
| `exp/hypothesis-engine` | Archivar | Experimento cerrado y autocontenido; solo conservar PREREG/RESULTS como referencia. |
| `exp/dialogue-selftrain` | Archivar | Ejecucion incompleta 7/8; H1 no soportada y referencias a evidencia no versionada. |
| `exp/es4-char-lm` | Archivar | 164 MiB de assets; no toca codigo de main. |
| `exp/byte-lm` | Archivar/release | Sin codigo propio (identico a mathzero-next); 131 MiB de bundles. |
| `exp/mathzero-next` | No mergear en bloque | 147 commits no validados; conflicto semantico con E9/E11/E16. Extraer fixes concretos. |

## Acciones recomendadas

1. No borrar ramas con resultados hasta crear un tag anotado de archivo.
2. `research/geometry-bridge`: no tocar hasta decidir E15; si se mergea,
   descartar los hunks de `exp.go` ajenos a geometria.
3. `research/traince-joint-contract`: rebasar sobre main, sustituir
   `panic` por errores y anadir un test del contrato (`t++` unico, `B2`
   invariante, signo del gradiente) antes de plantear merge.
4. `exp/mathzero-next`: si se quiere integrar algo, abrir una rama nueva
   pequena; no mergear el historico de pilotos.
5. `exp/mathzero-next` y `research/architecture24-utility-bidirectional`
   tienen el mismo fallo de contabilidad de lemas: su `GuidedSolve`
   bidireccional no incrementa `Uses`/`UsoMap` igual que main.
