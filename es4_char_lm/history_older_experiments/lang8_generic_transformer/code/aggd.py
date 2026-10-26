import json
B=json.load(open('base_d.json')); R={}; out={'tasks':{}}
for t in ('d1','d2'):
    for s in (21,22): R['%s_s%d'%(t,s)]=json.load(open('out/res_%s_s%d.json'%(t,s)))
for t in ('d1','d2'):
    det={}
    for s in (21,22):
        f=R['%s_s%d'%(t,s)]['final']; ok=True
        for sp in ('TR','E_b','E_l','E_bl'):
            x=f[sp]; ok=ok and x['free_exact']>=.99 and x['min_group']>=.80 and x['invalid']<=.05
        det[s]=dict(ok=bool(ok),exact={k:f[k]['free_exact'] for k in f},min_group={k:f[k]['min_group'] for k in f},invalid={k:f[k]['invalid'] for k in f},loss_last=R['%s_s%d'%(t,s)]['loss_curve'][-1],cf=R['%s_s%d'%(t,s)]['counterfactual'])
    out['tasks'][t]=dict(seeds=det,pass_both=all(d['ok'] for d in det.values()))
out['walls']={k:v['wall'] for k,v in R.items()}; out['rss']={k:v['rss_mib'] for k,v in R.items()}; out['parity']={k:v['parity'] for k,v in R.items()}; out['prior']=B['prior']
json.dump(out,open('out/result8d.json','w'),indent=1); print(json.dumps({t:{'pass_both':v['pass_both'],**{s:(d['ok'],d['exact']['TR'],d['exact']['E_bl'],round(d['loss_last'],3)) for s,d in v['seeds'].items()}} for t,v in out['tasks'].items()}))
