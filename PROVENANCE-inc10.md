# inc10: envoltorio de etapas del piloto P3 (NO EJECUTADO con el binario real)
- scripts/pilot_stage.sh: una etapa por llamada bajo guard.sh inc8; presupuesto TOTAL compartido entre etapas (900 s, fichero budget_used_s; cada etapa recibe min(110, restante)); preflight (memoria >=1300 MiB, sin otro d1train); etapa censurada o fallida queda registrada (censored_<etapa>) y NO se reintenta; nunca cambia parametros (configuracion congelada fija en el script).
- scripts/PILOT-INVOCATIONS.txt: invocaciones exactas, sin ejecutar.
- Verificado solo la logica de control con un binario FALSO (ok, repeticion rechazada, timeout censurado sin reintento, presupuesto agotado, 0 residuales). No se ha ejecutado d1train real.
