import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import sys,os,time,json,hashlib,numpy as np,torch
torch.set_num_threads(1)
sys.path.insert(0,(EXP_ROOT+'/lang10')); sys.path.insert(0,(EXP_ROOT+'/lang9')); sys.path.insert(0,(EXP_ROOT+'/lang8'))
import gt as G, gi, memguard, evalb
T0=time.perf_counter(); runid=sys.argv[1]; OUT=(EXP_ROOT+'/lang10p/')+runid; os.makedirs(OUT,exist_ok=False); memguard.check('start')
sha=lambda f:hashlib.sha256(open(f,'rb').read()).hexdigest()
Z=np.load((EXP_ROOT+'/lang10p/probe10.npz')); x1,x2,gold,kind,unit,R=Z['x1'],Z['x2'],Z['gold'],Z['kind'],Z['unit'],Z['R']
tgt=np.concatenate([gold+G.C0,np.full((len(gold),1),G.STOP)],1)
from lang7 import TRB
TR=set(TRB); res=dict(prompts_sha=sha((EXP_ROOT+'/lang10p/probe10.npz'))[:16],seeds={})
for s in (21,22):
    f=(EXP_ROOT+'/lang10/runs/c10_ctx_s%d/final_ctx_s%d.pt')%(s,s); h0=sha(f); exp=json.load(open((EXP_ROOT+'/lang10/runs/c10_ctx_s%d/res.json')%s))['final_sha']
    if h0!=exp: raise RuntimeError('final weights hash mismatch')
    m=gi.make(0); m.load_state_dict(torch.load(f)); m.eval()
    p1=evalb.greedy_batched(m,x1,128); p2=evalb.greedy_batched(m,x2,128); guard=time.perf_counter()-T0
    if guard>25: raise RuntimeError('internal wall guard')
    np.savez(OUT+'/raw_s%d.npz'%s,p1=p1,p2=p2,gold=tgt,kind=kind,unit=unit)
    c1=(p1==tgt).all(1); c2=(p2==tgt).all(1); same=(p1==p2).all(1)
    cat=dict(same_correct=same&c1&c2,same_wrong=same&~c1,differ_one=~same&(c1!=c2),differ_both_wrong=~same&~c1&~c2)
    out={}
    for k in ('SS','HS','HH'):
        mk=kind==k; U=np.unique(unit[mk]); r=dict(units=int(len(U)),replicates=int(mk.sum()))
        for nm,v in list(cat.items())+[('same_pred',same),('ctx1_correct',c1),('ctx2_correct',c2),('changed_wrong',~same)]:
            ur=np.array([v[mk&(unit==u)].mean() for u in U]); r[nm]=dict(mean_over_units=round(float(ur.mean()),3),min_unit=round(float(ur.min()),3),max_unit=round(float(ur.max()),3),pooled=round(float(v[mk].mean()),3))
        if k=='HS':
            seen1=np.array([a in TR for a in R[:,0]])[mk]; S_ok=np.where(seen1,c1[mk],c2[mk]); H_ok=np.where(seen1,c2[mk],c1[mk])
            r['seen_context_correct']=round(float(S_ok.mean()),3); r['heldout_context_correct']=round(float(H_ok.mean()),3)
            r['one_correct_and_it_is_seen']=round(float((S_ok&~H_ok).mean()),3); r['one_correct_and_it_is_heldout']=round(float((H_ok&~S_ok).mean()),3)
        out[k]=r
    if sha(f)!=h0: raise RuntimeError('weights changed')
    res['seeds'][s]=dict(strata=out,weights_sha_before_after=[h0[:16],sha(f)[:16]])
memguard.check('end'); _r,_h=memguard.usage(); res.update(rss_mib=_h,wall=time.perf_counter()-T0,script_sha=sha(os.path.abspath(__file__))[:16])
json.dump(res,open(OUT+'/res.json','w'),indent=1); print('done',round(res['wall'],1),round(_h,1))
