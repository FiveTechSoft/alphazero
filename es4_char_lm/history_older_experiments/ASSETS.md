# Assets de la release de archivo experimental y alcance

Release (prerelease experimental, no funcional): `es4-charlm-archivo-2026-10-26`
https://github.com/FiveTechSoft/alphazero/releases/tag/es4-charlm-archivo-2026-10-26

Tag sobre la rama `exp/es4-char-lm` (commit `005d3e72`). La rama `main` no se ha tocado.

## Assets verificados (descarga independiente, sha256 y bytes comprobados)

| Asset | Bytes | sha256 | Contenido |
|---|---|---|---|
| es4_char_lm_generations_g001-g025.tar | 1.455.349.760 | be28b04c6101dbb16165855f3ce1220c40bc81f1aba07a120ab6727182bbbcfe | Checkpoints intermedios ES4 g001-g025 (model.pt, adam.pt, train_state.json, rng_state.npz) |
| es4_char_lm_generations_g026-g050.tar | 1.455.349.760 | 1a393812f97f5657c955022ac11f5dcfb499ab5a9ef75be114c537fcd5b031c9 | Checkpoints g026-g050 |
| es4_char_lm_generations_g051-g074.tar | 1.397.135.360 | c5d39ca7534ee1fcad18d45fc0f8e8d6282aa092714d50d1d53e09cc849f2dea | Checkpoints g051-g074 |
| GENERATIONS_MANIFEST.json | 38.386 | d13d4444a1625f83d87f47a784e28884fc0b3cac0c2a03f6f105cd7d5f505896 | sha256 de los 296 ficheros de los 74 checkpoints |
| README_generations.md | 1.908 | 6a2f98f0277e472b4fda1ac63a65478f6ec1d21da1022b8a18ac278fe90f451e | Descripción y límites de la historia ES4 |
| ASSETS_MANIFEST.json | 554 | ab9fa2d8defcaf7a8da32e96dd54b62344748459032e76bad15fd3a31fdfe518 | sha256/bytes de los 3 tars de checkpoints |
| es4_char_lm_history_older_experiments_2B.tar | 135.936.000 | 6ce3a92ec68242fda6f2c70bf19a74a37f1b6a9358b8c93072c2d26a99dec234 | Pesos y datos de la historia de experimentos pequeños (lang5-lang12, calib, t0); 442 ficheros contra MANIFEST_2B.json de esta carpeta |
| b1_lineage_science_derived_retrospective.tgz | 3.606.248 | f5c0b5b9cb93827798d4d1dde9383dabe88bc0b5fb1fb109914f7c68f79f774b | Linea B1: checkpoint base ck_A94, fuentes Go mzab (archivadas, sin paridad de bytes), datasets y evaluaciones (198 ficheros contra su MANIFEST.sha256) |
| clean_2B_derived_retrospective_v3b.tgz | 16.068.793 | 8a078c1cb2cd890e8e9c7817278e24c23412c263ecaba59368013681170dfcbf | Export retrospectivo derivado lang0-lang4, lang11 y pilotos B1 (214 ficheros contra su MANIFEST.sha256) |

Los tars de checkpoints incluyen cada uno una copia idéntica de README.md y GENERATIONS_MANIFEST.json.

## Verificación y recuperación

- Cada asset se comprueba con `sha256sum` contra esta tabla (o contra ASSETS_MANIFEST.json para los 3 tars de checkpoints).
- Los 74 checkpoints se verifican miembro a miembro contra GENERATIONS_MANIFEST.json; la historia 2B contra MANIFEST_2B.json; el export retrospectivo contra su MANIFEST.sha256 interno.
- El estado Adam del test t0 va en dos partes: `cat adam.pt.part_00 adam.pt.part_01 > adam.pt`. Igual para `model/adam_g075.pt.part00/.part01` del paquete principal.
- Cargas: `torch.load(..., weights_only=True)` y `numpy.load(..., allow_pickle=False)`.

## Alcance y brechas (archivo de la ciencia recuperable)

- Publicado y verificado: paquete principal ES4 (rama, `es4_char_lm/`), etapa 2A (ES3, test, prefijo fuente), indice ligero 2B (esta carpeta), historia de checkpoints ES4, historia 2B, export retrospectivo lang0-lang4/pilotos y la linea B1 completa recuperable (assets anteriores).
- Limitaciones conocidas (documentadas en cada paquete): los binarios compilados `mzab_v3` y `mzab_v3s` se omiten por rutas de build incrustadas (sha256 y tamano registrados en BINARIES.md; fuentes Go archivadas sin paridad de bytes); los checkpoints intermedios A87-A93 y ck_b1 no se encontraron; logs y registros de ejecucion originales no incluidos. Los `seq_comp` del piloto B1 si estan incluidos.
- Esto es un archivo de toda la ciencia recuperable, no una copia bit-exacta de todo lo que existio.
