import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import numpy as np,time,json,os,pickle,hashlib
import lang6 as M
T0=time.perf_counter()
def guard():
    if time.perf_counter()-T0>150: raise RuntimeError('internal wall guard')
OUT=(EXP_ROOT+'/lang6/out'); os.makedirs(OUT,exist_ok=True)
STEPS=10000; LR=3e-3; BS=32; SEEDS=[21,22]; ARMS=['full','empty','perm']
EV=M.build_eval(3001)
np.savez(OUT+'/eval_frozen.npz',**{k+'_'+f:v for k,eps in EV.items() for f,v in dict(
    sw=np.stack([e['sw'] for e in eps]),sc=np.stack([e['sc'] for e in eps]),sc2=np.stack([e['sc2'] for e in eps]),lay=np.stack([e['lay'] for e in eps]),
    X=np.array([e['p'][0] for e in eps]),Y=np.array([e['p'][1] for e in eps]),gold=np.array([e['gold'] for e in eps]),grp=np.array([e['grp'] for e in eps])).items()})
EVH=hashlib.sha256(open(OUT+'/eval_frozen.npz','rb').read()).hexdigest()
def evaluate(P,arm):
    out={};preds={}
    for name,eps in EV.items():
        sw,sc,lay,X,Y,gold=M.pack(eps,arm); pr=M.predict(P,sw,sc,lay,X,Y); preds[name]=pr; ok=(pr==gold).all(1); g=np.array([e['grp'] for e in eps])
        ga=[float(ok[g==u].mean()) for u in np.unique(g)]
        out[name]=dict(joint=float(ok.mean()),src=float((pr[:,0]==gold[:,0]).mean()),dst=float((pr[:,1]==gold[:,1]).mean()),groups=len(ga),min_group=min(ga),groups_ge80=int(sum(a>=.8 for a in ga)),n=int(len(ok)))
    return out,preds
def run(seed,arm):
    tag='%s_s%d'%(arm,seed); P=M.init(seed); np.savez(OUT+'/init_'+tag+'.npz',**P)
    rng=np.random.RandomState(1000+seed); prng=np.random.RandomState(2000+seed)
    m={k:np.zeros_like(v) for k,v in P.items()}; v2={k:np.zeros_like(v) for k,v in P.items()}; t=0; acc=0.0; curve=[]
    for step in range(STEPS):
        eps=[M.gen(rng,M.TRB,M.SEEN,M.TRL,prng) for _ in range(BS)]
        for e in eps:
            if M.validate(e,M.TRB,M.SEEN,M.TRL): raise RuntimeError('train validation')
        sw,sc,lay,X,Y,gold=M.pack(eps,arm)
        assert sw.shape==(BS,4) and lay.shape==(BS,4) and sc.max()<=4 and sw.max()<=4
        loss,c=M.fwd(P,sw,sc,lay,X,Y,gold)
        if not np.isfinite(loss): raise RuntimeError('nonfinite')
        g=M.bwd(P,c); t+=1
        for k in P:
            m[k]=0.9*m[k]+0.1*g[k]; v2[k]=0.999*v2[k]+0.001*g[k]**2
            P[k]-=(LR*(m[k]/(1-0.9**t))/(np.sqrt(v2[k]/(1-0.999**t))+1e-8)).astype(np.float32)
        acc+=float(loss)
        if (step+1)%500==0: curve.append(acc/500); acc=0.0; guard()
    np.savez(OUT+'/final_'+tag+'.npz',**P); np.savez(OUT+'/adam_'+tag+'.npz',t=t,**{'m_'+k:v for k,v in m.items()},**{'v_'+k:v for k,v in v2.items()})
    pickle.dump(dict(rng=rng.get_state(),prng=prng.get_state()),open(OUT+'/rng_'+tag+'.pkl','wb'))
    res,preds=evaluate(P,arm); np.savez(OUT+'/pred_'+tag+'.npz',**preds)
    L=dict(np.load(OUT+'/final_'+tag+'.npz')); _,p2=evaluate(L,arm); parity=bool(all((preds[k]==p2[k]).all() for k in preds))
    r=dict(final=res,loss_curve=curve,parity=parity)
    if arm=='full': r['on_empty']=evaluate(P,'empty')[0]['E3']; r['on_perm']=evaluate(P,'perm')[0]['E3']
    return r
R=dict(eval_sha=EVH,results={})
for s in SEEDS:
    for a in ARMS:
        R['results']['%s_s%d'%(a,s)]=run(s,a); guard()
B3=0.2916666666666667; adv={}
for s in SEEDS:
    f=R['results']['full_s%d'%s]['final']['E3']; e=R['results']['empty_s%d'%s]['final']['E3']; p=R['results']['perm_s%d'%s]['final']['E3']
    adv[s]=dict(E3=f['joint'],min_group=f['min_group'],vs_empty=f['joint']-e['joint'],vs_perm=f['joint']-p['joint'],vs_bound=f['joint']-B3,
        ok=bool(f['joint']>=.90 and f['min_group']>=.80 and f['joint']-e['joint']>=.30 and f['joint']-p['joint']>=.30 and f['joint']>=B3+.30))
R['advance']=adv; R['advance_all']=bool(all(a['ok'] for a in adv.values())); R['wall']=time.perf_counter()-T0
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): R['rss_mib']=int(l.split()[1])/1024
json.dump(R,open(OUT+'/result.json','w'),indent=1); print('done',R['advance_all'])
