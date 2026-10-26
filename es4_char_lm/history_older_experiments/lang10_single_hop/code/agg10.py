import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import json,sys
R=(EXP_ROOT+'/lang10/runs/'); base=json.load(open((EXP_ROOT+'/lang10/base10.json'))); V={}
for s in (21,22):
    c=json.load(open(R+'c10_ctx_s%d/res.json'%s)); t=json.load(open(R+'c10_ctl_s%d/res.json'%s)); F=c['final']; Tc=t['final']
    S1=F['E_b']['ctx']['exact']>=.90 and F['E_bs']['ctx']['exact']>=.90
    S2=F['E_b']['ctx']['min_group']>=.80
    S3=F['E_b']['cf_support']['exact_new']>=.90
    S4=F['E_b']['nocontext']['exact']<=base['E_b']['prior_exact']+.10
    S5=Tc['TR']['ctx']['exact']<=base['TR']['prior_exact']+.05 and Tc['E_b']['ctx']['exact']<=base['E_b']['prior_exact']+.05 and F['E_b']['ctx']['exact']-Tc['E_b']['ctx']['exact']>=.50
    V[s]=dict(S1=S1,S2=S2,S3=S3,S4=S4,S5=S5,all=all([S1,S2,S3,S4,S5]),ctx_TR=F['TR']['ctx']['exact'],ctx_E_b=F['E_b']['ctx']['exact'],ctx_E_s=F['E_s']['ctx']['exact'],ctx_E_bs=F['E_bs']['ctx']['exact'],ctx_Eb_min_group=F['E_b']['ctx']['min_group'],
              ctx_nocontext_E_b=F['E_b']['nocontext']['exact'],ctx_cf_E_b=F['E_b']['cf_support'],ctl_TR=Tc['TR']['ctx']['exact'],ctl_E_b=Tc['E_b']['ctx']['exact'],loss_ctx=c['loss_curve'][-1],loss_ctl=t['loss_curve'][-1],
              hwm=[c['rss_mib'],t['rss_mib']],parity=[c['parity'],t['parity']],stream_equal=c['stream_hash']==t['stream_hash'])
both=all(V[s]['all'] for s in V); mem=all(V[s]['ctx_TR']>=.90 for s in V)
verdict='RETRIEVAL_LEARNED' if both else ('BINDINGS_SEEN_ONLY' if mem else 'NOT_LEARNED')
out=dict(verdict=verdict,seeds=V,baseline=base); json.dump(out,open(R+'result_c10.json','w'),indent=1); print(json.dumps(out,indent=1))
