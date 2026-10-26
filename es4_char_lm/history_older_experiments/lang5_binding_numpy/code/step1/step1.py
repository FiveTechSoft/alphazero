import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import numpy as np, time, json, os, hashlib, pickle
import diag5 as D
T0=time.perf_counter()
OUT=(EXP_ROOT+'/lang5/step1/out'); os.makedirs(OUT,exist_ok=True)
STEPS=2500; LR=3e-3
rng=np.random.RandomState(31337); eps=[]
for _ in range(32):
    b=int(D.TRB[rng.randint(len(D.TRB))]); qp=D.SEEN[rng.randint(len(D.SEEN))]; ep=D.gen_episode(rng,b,qp)
    if D.validate(ep,'seen'): raise RuntimeError('validation')
    eps.append(ep)
sup=np.stack([D.tok_support(e) for e in eps]); q=np.stack([D.tok_query(e) for e in eps]); tgt=np.array([e['qans'] for e in eps],dtype=np.int64)
assert q.max()<8 and sup[:,:,:6].max()<8 and sup[:,:,6:].min()>=8
np.savez(OUT+'/batch.npz',sup=sup,q=q,tgt=tgt)
P=D.init(np.random.RandomState(21),16,64,np.float32); P['E']=np.vstack([P['E'],(np.random.RandomState(21).randn(1,16)*0.1).astype(np.float32)])

P0={k:v.copy() for k,v in P.items()}
def pred(P):
    _,c=D.fwd(P,sup,q,tgt); return np.stack([c[6].argmax(1),c[7].argmax(1)],1)
pr0=pred(P); acc0=float((pr0==tgt).all(1).mean()); loss0=float(D.fwd(P,sup,q,tgt)[0])
np.savez(OUT+'/init.npz',**P)
m={k:np.zeros_like(v) for k,v in P.items()}; v2={k:np.zeros_like(v) for k,v in P.items()}; t=0; curve=[]; step1_norm=None
for step in range(STEPS):
    loss,c=D.fwd(P,sup,q,tgt)
    if not np.isfinite(loss): raise RuntimeError('nonfinite')
    g=D.bwd(P,sup,q,tgt,c); t+=1; before={k:P[k].copy() for k in P} if step in (0,STEPS-1) else None
    for k in P:
        m[k]=0.9*m[k]+0.1*g[k]; v2[k]=0.999*v2[k]+0.001*g[k]**2
        P[k]-=(LR*(m[k]/(1-0.9**t))/(np.sqrt(v2[k]/(1-0.999**t))+1e-8)).astype(np.float32)
    if step==0: step1_norm={k:float(np.linalg.norm(P[k]-before[k])) for k in P}; g1={k:float(np.linalg.norm(g[k])) for k in P}
    if step==STEPS-1: last_norm={k:float(np.linalg.norm(P[k]-before[k])) for k in P}
    if step%100==0 or step==STEPS-1: curve.append((step,float(loss)))
    if time.perf_counter()-T0>25: raise RuntimeError('guard')
lossF=float(D.fwd(P,sup,q,tgt)[0]); prF=pred(P); accF=float((prF==tgt).all(1).mean())
np.savez(OUT+'/final.npz',**P); np.savez(OUT+'/adam.npz',t=t,**{'m_'+k:v for k,v in m.items()},**{'v_'+k:v for k,v in v2.items()})
pickle.dump(rng.get_state(),open(OUT+'/rng.pkl','wb')); np.savez(OUT+'/pred.npz',before=pr0,after=prF,gold=tgt)
L=dict(np.load(OUT+'/final.npz')); parity=bool((pred(L)==prF).all())
R=dict(init_loss=loss0,final_loss=lossF,init_joint=acc0,final_joint=accF,src_final=float((prF[:,0]==tgt[:,0]).mean()),dst_final=float((prF[:,1]==tgt[:,1]).mean()),
  pass_=bool(accF>=0.95 and lossF<=0.2),reload_parity=parity,
  step1_update_norm=step1_norm,step1_grad_norm=g1,last_step_update_norm=last_norm,
  total_change_norm={k:float(np.linalg.norm(P[k]-P0[k])) for k in P},curve=curve,wall=time.perf_counter()-T0)
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): R['rss_peak_mib']=int(l.split()[1])/1024
json.dump(R,open(OUT+'/result.json','w'),indent=1); print('done',R['pass_'])
