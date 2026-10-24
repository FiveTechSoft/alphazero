# Nota de estado final inc7-inc11 (exp/mathzero-next)

La rama incorpora inc7, inc8, inc9, inc10 e inc11 en una sola pasada, sobre base 48c87b8.

- scripts/guard.sh de inc7 (con defecto de limpieza de descendientes) NO se subio: fue
  reemplazada antes de tocar el remoto por la guardia corregida de inc8, unica version de
  scripts/guard.sh en la rama. La copia incluida en inc11 es byte-identica a la de inc8
  (sha256 prefijo da23456b580e7006); no se duplica.
- scripts/pilot_stage.sh y scripts/PILOT-INVOCATIONS.txt de inc10 NO se suben: inc11 los
  sustituye antes de tocar el remoto (correccion minima: huella de binario exigida, lock
  atomico, presupuesto atomico y fallo cerrado, reserva conservadora, registros durables).
  La version de la rama es la de inc11.
- PROVENANCE-incN.md y MANIFEST-incN.sha256 (N=7..11) son registros historicos de los
  paquetes tal como se recibieron; mencionan archivos omitidos o sustituidos, por lo que
  NO verifican por si solos el estado final combinado.
- MANIFEST-final-inc7-inc8.sha256 cubre exactamente los archivos del estado final
  incorporados por inc7-inc11 (incluidas procedencias y manifiestos historicos y esta
  nota; se excluye a si mismo). Es el manifiesto que verifica el arbol final. (El nombre
  conserva inc7-inc8 por la autorizacion original; el alcance es inc7-inc11.)
- El piloto P3 NO se ha lanzado: scripts/PILOT-INVOCATIONS.txt documenta las invocaciones
  exactas pendientes de autorizacion. Ningun hecho fue reescrito; sin registros internos.
  main (d93ea95) no se toca.
- scripts/PILOT-INVOCATIONS.txt incluye una correccion de documentacion posterior a la revision
  (requisito BIN_SHA explícito), hecha por el verificador el 2026-10-24 por instruccion del
  coordinador; no se altero codigo revisado. El texto original de inc10/inc11 no la incluia.
