import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import numpy as np, time, json, sys, os, hashlib, pickle
from collections import Counter
import diag5 as D
T0=time.perf_counter()
def guard():
    if time.perf_counter()-T0>105: raise RuntimeError('internal wall guard')
OUT=(EXP_ROOT+'/lang5/pilot/out'); os.makedirs(OUT,exist_ok=True)
SEEDS=[21,22]; ARMS=['full','empty','perm']; STEPS=2500; LR=3e-3; B=32; PAD=12
CFG=dict(seeds=SEEDS,arms=ARMS,steps=STEPS,lr=LR,batch=B,d=16,H=64,ns=D.NS,eval_seed=777,
         train_bindings=D.TRB,new_bindings=D.NEWB,seen=D.SEEN,reserved=D.RES)
json.dump(CFG,open(OUT+'/config.json','w'))

def build_eval():
    rng=np.random.RandomState(777); rngp=np.random.RandomState(778); sets={}
    spec={'E3':(D.NEWB,D.RES,50,'res'),'E1':(D.NEWB,D.SEEN,25,'seen'),'E2':(D.TRB,D.RES,10,'res'),'TR':(D.TRB,D.SEEN,5,'seen')}
    for name,(bs,ps,K,pt) in spec.items():
        q=[];st=[];sp=[];tg=[];grp=[]
        for b in bs:
            for pi,p in enumerate(ps):
                for _ in range(K):
                    ep=D.gen_episode(rng,int(b),p)
                    if D.validate(ep,pt): raise RuntimeError('eval validation violation '+name)
                    b2=int(rngp.choice([x for x in range(24) if x!=b]))
                    ep2=dict(ep); ep2['sup']=[(pp,l,D.answer(b2,pp,l)) for pp,l,a in ep['sup']]
                    q.append(D.tok_query(ep)); st.append(D.tok_support(ep)); sp.append(D.tok_support(ep2)); tg.append(ep['qans']); grp.append(int(b)*100+D.PAIRS.index(p))
        sets[name]=dict(q=np.stack(q),sup=np.stack(st),sup_perm=np.stack(sp),tgt=np.array(tg),grp=np.array(grp))
        sets[name]['sup_empty']=np.full_like(sets[name]['sup'],PAD)
    return sets
def noctx_bound(bs,ps):
    s=0;m=0
    for p in ps:
        for L in range(24):
            c=Counter(D.answer(b,p,L) for b in bs); s+=c.most_common(1)[0][1]/len(bs); m+=1
    return s/m
EV=build_eval(); guard()
np.savez(OUT+'/eval_frozen.npz',**{k+'_'+f:v for k,s in EV.items() for f,v in s.items()})
ev_hash=hashlib.sha256(open(OUT+'/eval_frozen.npz','rb').read()).hexdigest()
BOUND={'E3':noctx_bound(D.NEWB,D.RES),'E1':noctx_bound(D.NEWB,D.SEEN),'E2':noctx_bound(D.TRB,D.RES),'TR':noctx_bound(D.TRB,D.SEEN)}
def predict(P,sup,q):
    dummy=np.zeros((len(q),2),dtype=np.int64)
    _,c=D.fwd(P,sup,q,dummy); return np.stack([c[6].argmax(1),c[7].argmax(1)],1)
def evaluate(P,arm,sets,mode=None):
    mode=mode or arm; out={}; preds={}
    key={'full':'sup','empty':'sup_empty','perm':'sup_perm'}[mode]
    for name,s in sets.items():
        pr=predict(P,s[key],s['q']); preds[name]=pr
        ok=(pr==s['tgt']).all(1); g=s['grp']
        ga=[float(ok[g==u].mean()) for u in np.unique(g)]
        out[name]=dict(joint=float(ok.mean()),src=float((pr[:,0]==s['tgt'][:,0]).mean()),dst=float((pr[:,1]==s['tgt'][:,1]).mean()),
                       n=int(len(ok)),groups=len(ga),min_group=min(ga),mean_group=float(np.mean(ga)),groups_ge80=int(sum(a>=0.8 for a in ga)))
    return out,preds
def mk_P(seed):
    rng=np.random.RandomState(seed); P=D.init(rng,16,64,np.float32)
    P['E']=np.vstack([P['E'],(rng.randn(1,16)*0.1).astype(np.float32)]); return P
def mk_batch(rng,rngp,arm):
    eps=[]
    for _ in range(B):
        b=int(D.TRB[rng.randint(len(D.TRB))]); qp=D.SEEN[rng.randint(len(D.SEEN))]
        ab=None
        if arm=='perm': ab=int(rngp.choice([x for x in range(24) if x!=b]))
        ep=D.gen_episode(rng,b,qp,ab)
        v=[k for k in D.validate(ep,'seen') if not(arm=='perm' and k=='oracle_wrong')]
        if v: raise RuntimeError('train validation violation '+str(v))
        eps.append(ep)
    sup=np.stack([D.tok_support(e) for e in eps]); q=np.stack([D.tok_query(e) for e in eps]); tgt=np.array([e['qans'] for e in eps],dtype=np.int64)
    assert q.max()<8 and sup.shape==(B,D.NS,8)
    assert sup[:,:,:6].max()<8 and sup[:,:,6:].min()>=8 and sup[:,:,6:].max()<12
    if arm=='empty': sup=np.full_like(sup,PAD)
    return sup,q,tgt
def run_arm(seed,arm):
    P=mk_P(seed); rng=np.random.RandomState(1000+seed); rngp=np.random.RandomState(2000+seed)
    tag='%s_s%d'%(arm,seed); np.savez(OUT+'/init_'+tag+'.npz',**P)
    ev0,_=evaluate(P,arm,{'E3':EV['E3']})
    m={k:np.zeros_like(v) for k,v in P.items()}; v2={k:np.zeros_like(v) for k,v in P.items()}; t=0; log=[]; acc=0.0
    for step in range(STEPS):
        sup,q,tgt=mk_batch(rng,rngp,arm)
        loss,c=D.fwd(P,sup,q,tgt)
        if not np.isfinite(loss): raise RuntimeError('nonfinite loss')
        g=D.bwd(P,sup,q,tgt,c); t+=1
        for k in P:
            m[k]=0.9*m[k]+0.1*g[k]; v2[k]=0.999*v2[k]+0.001*g[k]**2
            P[k]-= (LR*(m[k]/(1-0.9**t))/(np.sqrt(v2[k]/(1-0.999**t))+1e-8)).astype(np.float32)
        acc+=float(loss)
        if (step+1)%250==0: log.append(acc/250); acc=0.0; guard()
    np.savez(OUT+'/final_'+tag+'.npz',**P)
    np.savez(OUT+'/adam_'+tag+'.npz',t=t,**{'m_'+k:v for k,v in m.items()},**{'v_'+k:v for k,v in v2.items()})
    pickle.dump(dict(train_rng=rng.get_state(),perm_rng=rngp.get_state()),open(OUT+'/rng_'+tag+'.pkl','wb'))
    res,preds=evaluate(P,arm,EV)
    np.savez(OUT+'/raw_pred_'+tag+'.npz',**{k:v for k,v in preds.items()},**{k+'_gold':EV[k]['tgt'] for k in EV},**{k+'_grp':EV[k]['grp'] for k in EV})

    L=dict(np.load(OUT+'/final_'+tag+'.npz')); res2,preds2=evaluate(L,arm,EV)
    parity=all((preds[k]==preds2[k]).all() for k in preds)
    extra={}
    if arm=='full':
        extra['full_on_empty_support']=evaluate(P,arm,EV,'empty')[0]['E3']; extra['full_on_perm_support']=evaluate(P,arm,EV,'perm')[0]['E3']
    return dict(init_E3=ev0['E3'],final=res,loss_every250=log,reload_parity=bool(parity),extra=extra,steps=t)
R={'ev_hash':ev_hash,'noctx_bound_exact_by_split':BOUND,'results':{}}
for seed in SEEDS:
    for arm in ARMS:
        R['results']['%s_s%d'%(arm,seed)]=run_arm(seed,arm); guard()
adv={}
for seed in SEEDS:
    f=R['results']['full_s%d'%seed]['final']['E3']; e=R['results']['empty_s%d'%seed]['final']['E3']; p=R['results']['perm_s%d'%seed]['final']['E3']
    adv[seed]=dict(joint_ge90=f['joint']>=0.90,min_group_ge80=f['min_group']>=0.80,adv_over_empty=f['joint']-e['joint'],adv_over_perm=f['joint']-p['joint'],
        ok=bool(f['joint']>=0.90 and f['min_group']>=0.80 and f['joint']-e['joint']>=0.30 and f['joint']-p['joint']>=0.30))
R['advance']=adv; R['advance_all']=bool(all(a['ok'] for a in adv.values()))
R['wall_sec']=time.perf_counter()-T0
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): R['rss_peak_mib']=int(l.split()[1])/1024.0
json.dump(R,open(OUT+'/result.json','w'),indent=1); print('done',R['advance_all'])
