import json
B=json.load(open('cost8.json')); BASE=B['baselines']; R={}
for s in (21,22):
    for a in ('full','nocue'): R['%s_s%d'%(a,s)]=json.load(open('out/res_%s_s%d.json'%(a,s)))
adv={}
for s in (21,22):
    f=R['full_s%d'%s]['final']; n=R['nocue_s%d'%s]['final']; ok=f['TR']['free_exact']>=.99 and f['TR']['invalid']<=.05; det={'TR':f['TR']['free_exact']}
    for sp in ('E3','E5'):
        x=f[sp]; b=max(BASE[sp]['first_is_src'],BASE[sp]['second_is_src'],BASE[sp]['train_pair_template'])
        d=dict(free_exact=x['free_exact'],min_group=x['min_group'],invalid=x['invalid'],vs_nocue=x['free_exact']-n[sp]['free_exact'],vs_best_baseline=x['free_exact']-b); det[sp]=d
        ok=ok and x['free_exact']>=.9 and x['min_group']>=.8 and x['invalid']<=.05 and d['vs_nocue']>=.3 and d['vs_best_baseline']>=.3
    det['ok']=bool(ok); adv[s]=det
out=dict(advance=adv,advance_all=all(a['ok'] for a in adv.values()),walls={k:v['wall'] for k,v in R.items()},rss={k:v['rss_mib'] for k,v in R.items()},parity={k:v['parity'] for k,v in R.items()})
json.dump(out,open('out/result8.json','w'),indent=1); print(json.dumps(out))
