# ADENDA 1 al diseno G1 (versionada; NO corrige retroactivamente)
Base: DESIGN-G1-frozen.md (SHA256 837aa5dd0133a8b524026a1eb6050c3680f141c3473cd81e8a612ae1ae5f4879), que fija "numero de semillas con delta>0, test de signos exacto bilateral NOMINAL, condicional al banco fijo" y no declara convencion de empates ni direccion de lectura.
Fecha de redaccion: 24 oct 2026, DESPUES del piloto G1 semilla 101 (completo, informado) y ANTES de lanzar cualquier semilla de validacion 11..18. El piloto 101 no entra en el criterio ni fue usado para elegir esta regla (la regla es la convencion estandar del test de signos).
Banco, regla del banco, semilla 20261024, codigo, configuracion y checkpoints NO cambian.

## Reglas
1. Por cada semilla se reporta delta = resueltos post - pre sobre el banco G1 (90 casos, cap 100), con signo: positiva (delta>0), negativa (delta<0), cero (delta=0). Se reporta el conteo completo P (positivas), N (negativas), Z (ceros) con las 8 semillas listadas, integras.
2. Test de signos exacto bilateral estandar: los ceros se EXCLUYEN; n efectivo = P+N, explicito en el informe. p = min(1, 2*P(X<=min(P,N))) con X~Binomial(n efectivo, 1/2). Nominal, condicional al banco fijo y a esta configuracion.
3. Direccion: el signo de la mayoria interpreta el resultado. P>N con p pequeno apunta a mejora; N>P con p pequeno apunta a EMPEORAMIENTO, no a exito (p=0.0078125 con 8/8 negativas es empeoramiento). P=N: sin direccion.
4. Si n efectivo = 0: resultado NO INFORMATIVO; no se afirma mejora ni empeoramiento.
5. El conteo "semillas con delta>0 de 8" del diseno original se conserva como criterio declarado y descriptivo; no sustituye al test de la regla 2 ni se reinterpreta tras ver los resultados.
6. Ejemplos de la formula (verificables): 8P/0N/0Z p=0.0078125; 7P/0N/1Z n=7 p=0.015625; 7P/1N/0Z n=8 p=0.0703125; 8N p=0.0078125 (empeoramiento); 0P/0N/8Z no informativo.
7. Sin comparaciones con P3 ni bancos antiguos. Limites del diseno (disjuncion modulo AC estructural, sin control SGD, una configuracion, n=8) siguen vigentes.
