import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import numpy as np,time,json,hashlib
import lang7 as L, lang7e as M
T0=time.perf_counter(); res={}; OUT=(EXP_ROOT+'/lang7e/out')
res['code_sha']={f:hashlib.sha256(open(f,'rb').read()).hexdigest() for f in ['lang7.py','lang7e.py']}
res['gradcheck']={}
for up in (True,False):
    w,n=M.gradcheck(up); res['gradcheck']['useprev=%s'%up]=dict(worst=w,n=n)
    if w>1e-6: res['STOP']='gradcheck'; json.dump(res,open('cost.json','w'),indent=1); raise SystemExit
EV=L.build_eval(3101)
flat={}
for k,eps in EV.items():
    flat.update({k+'_T':np.stack([e['T'] for e in eps]),k+'_sw':np.stack([e['sw'] for e in eps]),k+'_sc':np.stack([e['sc'] for e in eps]),k+'_sc2':np.stack([e['sc2'] for e in eps]),
                 k+'_lay':np.stack([e['lay'] for e in eps]),k+'_gold':np.array([e['gold'] for e in eps]),k+'_grp':np.array([e['grp'] for e in eps])})
np.savez(OUT+'/eval_frozen.npz',**flat)
res['eval_sha']=hashlib.sha256(open(OUT+'/eval_frozen.npz','rb').read()).hexdigest()
_o=np.load((EXP_ROOT+'/lang7/out/eval_frozen.npz')); res['eval_content_equals_lang7']=bool(set(_o.files)==set(flat) and all((_o[k]==flat[k]).all() for k in flat))
res['counts']={k:len(v) for k,v in EV.items()}; res['groups']={k:len(set(e['grp'] for e in v)) for k,v in EV.items()}
res['noctx_bound']={k:L.noctx_bound((L.SPECS[k][0],L.SPECS[k][1],L.SPECS[k][2],L.SPECS[k][5])) for k in L.SPECS}
res['baselines_perfect_binding']={k:L.baselines(v) for k,v in EV.items()}

P=M.init(7); T,sw,sc,lay,gold=L.pack(EV['E3'],'full'); fp=M.free_pred(P,T,sw,sc,lay,True); tp=M.tf_pred(P,T,sw,sc,lay,gold,True)
res['contract']=dict(free_shape=list(fp.shape),tf_shape=list(tp.shape),free_range=[int(fp.min()),int(fp.max())])
if fp.shape!=(len(T),3) or tp.shape!=(len(T),3) or fp.min()<0 or fp.max()>=M.NV: res['STOP']='contract'; json.dump(res,open('cost.json','w'),indent=1); raise SystemExit
rng=np.random.RandomState(99); prng=np.random.RandomState(98); tg=[];tf=[];ta=[]
m={k:np.zeros_like(v) for k,v in P.items()}; v2={k:np.zeros_like(v) for k,v in P.items()}; t=0
for i in range(300):
    a=time.perf_counter(); eps=[L.gen(rng,L.TRB,L.TRS,L.TRL,'train',prng) for _ in range(32)]
    for e in eps:
        if L.validate(e,L.TRB,L.TRS,L.TRL): raise RuntimeError('train validation')
    T,sw,sc,lay,gold=L.pack(eps,'full'); b=time.perf_counter()
    loss,c=M.train_loss(P,T,sw,sc,lay,gold,True); g=M.bwd(P,c); e_=time.perf_counter(); t+=1
    for k in P:
        m[k]=0.9*m[k]+0.1*g[k]; v2[k]=0.999*v2[k]+0.001*g[k]**2
        P[k]-=(3e-3*(m[k]/(1-0.9**t))/(np.sqrt(v2[k]/(1-0.999**t))+1e-8)).astype(np.float32)
    f=time.perf_counter(); tg.append(b-a); tf.append(e_-b); ta.append(f-e_)
tot=float(np.median(tg)+np.median(tf)+np.median(ta))
t1=time.perf_counter(); allT=[L.pack(v,'full') for v in EV.values()]
for T,sw,sc,lay,gold in allT: M.free_pred(P,T,sw,sc,lay,True); M.tf_pred(P,T,sw,sc,lay,gold,True)
res['cost']=dict(gen_validate=float(np.median(tg)),fwd_bwd=float(np.median(tf)),adam=float(np.median(ta)),step_total=tot,eval_all_splits_s=time.perf_counter()-t1,params=int(sum(v.size for v in P.values())))
res['wall']=time.perf_counter()-T0
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): res['rss_mib']=int(l.split()[1])/1024
json.dump(res,open('cost.json','w'),indent=1); print('done',tot)
