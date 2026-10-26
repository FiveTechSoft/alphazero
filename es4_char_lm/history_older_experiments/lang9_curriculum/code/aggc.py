import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import json
R={}
for s in (21,22):
    for a in ('cur','ctl'): R[(a,s)]=json.load(open((EXP_ROOT+'/lang9/runs/c9_%s_s%d/res.json')%(a,s)))
out={'seeds':{}}; ben=True; strong=True
for s in (21,22):
    c=R[('cur',s)]; k=R[('ctl',s)]; fc=c['final_d2']; fk=k['final_d2']
    d=dict(cur_exact={x:fc[x]['free_exact'] for x in fc},ctl_exact={x:fk[x]['free_exact'] for x in fk},E_bl_diff=fc['E_bl']['free_exact']-fk['E_bl']['free_exact'],
           cur_min_group={x:fc[x]['min_group'] for x in fc},ctl_min_group={x:fk[x]['min_group'] for x in fk},same_stream_from_switch=c['stream_hash_from_switch']==k['stream_hash_from_switch'],same_stream_all=c['stream_hash_all']==k['stream_hash_all'],
           d1_at_switch={x:v['free_exact'] for x,v in c['d1_at_switch'].items()},d1_retention={x:v['free_exact'] for x,v in c['d1_retention_final'].items()},cf_cur=c['counterfactual_d2'],cf_ctl=k['counterfactual_d2'],
           loss_last_cur=c['loss_curve'][-1],loss_last_ctl=k['loss_curve'][-1])
    b=d['E_bl_diff']>=.20 and fc['TR']['free_exact']>=.90; st=all(fc[x]['free_exact']>=.99 and fc[x]['min_group']>=.80 for x in ('TR','E_bl'))
    d['benefit']=bool(b); d['strong']=bool(st); ben=ben and b; strong=strong and st; out['seeds'][s]=d
out['benefit_both']=bool(ben); out['strong_both']=bool(strong); out['walls']={'%s_s%d'%k:v['wall'] for k,v in R.items()}; out['hwm']={'%s_s%d'%k:v['rss_mib'] for k,v in R.items()}; out['parity']={'%s_s%d'%k:v['parity'] for k,v in R.items()}
json.dump(out,open((EXP_ROOT+'/lang9/runs/result_c9.json'),'w'),indent=1); print(json.dumps({s:(d['benefit'],d['strong'],round(d['E_bl_diff'],3),d['same_stream_from_switch']) for s,d in out['seeds'].items()}),out['benefit_both'],out['strong_both'])
