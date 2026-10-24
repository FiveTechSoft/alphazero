# Nota de correccion (2026-10-24) — MANIFEST.sha256 de pilot-101

La version original de MANIFEST.sha256 incluia una auto-entrada con el hash del
archivo vacio (e3b0c44298fc...), puesta como marcador por construccion. Esa
entrada no verificaba nada (el archivo no esta vacio) y se ha eliminado.

Sin tocar: los 28 hashes originales ni ningun dato/resultado del piloto; todos
siguen verificando byte a byte contra los archivos de este directorio
(sha256sum -c MANIFEST.sha256).

Convencion corregida (la misma aplicada en pilots/p3/validation-11-18/): el
manifiesto no se lista a si mismo; su propio hash se registra fuera del archivo
(historial de commits e informes de auditoria). Esta nota se anade como archivo
nuevo y si figura en el MANIFEST corregido.
