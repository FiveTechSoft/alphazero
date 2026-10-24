# Piloto P3, seed 101 (RESULTADO EXPLORATORIO — n=1)

Salida cruda del primer piloto del diseno congelado P3 (pilots/p3/DESIGN-P3-frozen.md),
ejecutado por etapas bajo guardia inc8 + wrapper inc11, y verificado de forma independiente
(rejuicio exacto de todos los casos de transicion con ambas redes, recomputo de digests,
solapes y hashes de pesos).

Resultado medido: primario 43->61/90 (+19/-1); secundario 53->71/90 (+18/0).

LIMITACIONES:
- n=1: una sola semilla; sin valor de validacion estadistica.
- 82/240 (primario) y 89/240 (secundario) de los starts de entrenamiento aparecen en la
  evaluacion con OTRO target: transferencia cercana, no generalizacion limpia. Pair=0 solo
  garantiza que ningun PAR exacto de entrenamiento estaba en evaluacion.
- Niveles de entrenamiento 2..7 se solapan con los de evaluacion 6..8.
- Sin control SGD. cap 100 igual en train/eval.
- NO es aprendizaje demostrado ni verificacion independiente del diseno; NO equivale a
  ningun resultado historico (Reduced8 o anterior).
- El binario del piloto no fue reconstruido byte a byte (toolchain distinta); la
  reproduccion funcional (solved y nodos exactos) y el hash del binario son evidencias
  distintas, ninguna sustituye a la otra.
- La RSS era sondeo (guardia inc8): sobrepaso acotado documentado en P3-cost-measurements.md.

Transformacion respecto al paquete recibido: HASHES.txt depurado (rutas de trabajo
eliminadas, hashes conservados); el resto de archivos sin cambios byte a byte.
Normalizacion: a los 5 stage JSON se anadio un newline final (el editor web lo impone);
contenido JSON intacto. Los hashes de train.json en HASHES.txt corresponden al archivo
ORIGINAL sin ese newline; el MANIFEST.sha256 de este directorio refleja los bytes aqui
presentes.
