import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import numpy as np,time,json,os,pickle,hashlib
import lang7 as M
T0=time.perf_counter()
def guard():
    if time.perf_counter()-T0>150: raise RuntimeError('internal wall guard')
OUT=(EXP_ROOT+'/lang7/out'); os.makedirs(OUT,exist_ok=True)
STEPS=6000; LR=3e-3; BS=32; SEEDS=[21,22]; ARMS=['full','nocue','empty','perm']
EV=M.build_eval(3101)
flat={}
for k,eps in EV.items():
    flat.update({k+'_T':np.stack([e['T'] for e in eps]),k+'_sw':np.stack([e['sw'] for e in eps]),k+'_sc':np.stack([e['sc'] for e in eps]),k+'_sc2':np.stack([e['sc2'] for e in eps]),
                 k+'_lay':np.stack([e['lay'] for e in eps]),k+'_gold':np.array([e['gold'] for e in eps]),k+'_grp':np.array([e['grp'] for e in eps])})
np.savez(OUT+'/eval_frozen.npz',**flat); EVH=hashlib.sha256(open(OUT+'/eval_frozen.npz','rb').read()).hexdigest()
BASE={k:M.baselines(v) for k,v in EV.items()}
NOC={k:M.noctx_bound((M.SPECS[k][0],M.SPECS[k][1],M.SPECS[k][2],M.SPECS[k][5])) for k in M.SPECS}
def evaluate(P,arm,names=None):
    out={};preds={}
    for name,eps in EV.items():
        if names and name not in names: continue
        T,sw,sc,lay,gold=M.pack(eps,arm); pr=M.predict(P,T,sw,sc,lay); preds[name]=pr; ok=(pr==gold).all(1); g=np.array([e['grp'] for e in eps])
        ga=[float(ok[g==u].mean()) for u in np.unique(g)]
        out[name]=dict(joint=float(ok.mean()),src=float((pr[:,0]==gold[:,0]).mean()),dst=float((pr[:,1]==gold[:,1]).mean()),groups=len(ga),min_group=min(ga),groups_ge80=int(sum(a>=.8 for a in ga)),n=int(len(ok)))
    return out,preds
def agree(preds,name):

    eps=EV[name]; pr=preds[name]; res={}
    pol={'first_is_src':lambda f,s:(f,s),'second_is_src':lambda f,s:(s,f),'train_pair_template':lambda f,s:(f,s) if frozenset((f,s)) in M.ORDER0 else (s,f)}
    for k,fn in pol.items():
        ok=0
        for e,p in zip(eps,pr):
            words=[t for t in e['T'] if t<4]; src,dst=fn(int(words[0]),int(words[1])); ok+=(M.answer(e['b'],(src,dst),e['l'])==tuple(p))
        res[k]=ok/len(eps)
    return res
def run(seed,arm):
    tag='%s_s%d'%(arm,seed); P=M.init(seed); np.savez(OUT+'/init_'+tag+'.npz',**P)
    rng=np.random.RandomState(1000+seed); prng=np.random.RandomState(2000+seed)
    m={k:np.zeros_like(v) for k,v in P.items()}; v2={k:np.zeros_like(v) for k,v in P.items()}; t=0;acc=0.0;curve=[]
    for step in range(STEPS):
        eps=[M.gen(rng,M.TRB,M.TRS,M.TRL,'train',prng) for _ in range(BS)]
        for e in eps:
            if M.validate(e,M.TRB,M.TRS,M.TRL): raise RuntimeError('train validation')
        T,sw,sc,lay,gold=M.pack(eps,arm)
        assert T.shape==(BS,5) and T.max()<=6 and sw.shape==(BS,4) and sc.max()<=4
        loss,c=M.fwd(P,T,sw,sc,lay,gold)
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
    r=dict(final=res,loss_curve=curve,parity=parity,agree_E3=agree(preds,'E3'),agree_E5=agree(preds,'E5'))
    if arm=='full':
        r['masked_cues']=evaluate(P,'nocue',['TR','E1','E3'])[0]; r['on_empty_support']=evaluate(P,'empty',['TR','E3'])[0]; r['on_perm_support']=evaluate(P,'perm',['TR','E3'])[0]
    return r
R=dict(eval_sha=EVH,baselines=BASE,noctx=NOC,results={})
for s in SEEDS:
    for a in ARMS:
        R['results']['%s_s%d'%(a,s)]=run(s,a); guard()
adv={}
for s in SEEDS:
    f=R['results']['full_s%d'%s]['final']; ok=f['TR']['joint']>=.99; det={'TR':f['TR']['joint']}
    for sp in ['E3','E5']:
        x=f[sp]; b=max(BASE[sp]['first_is_src'],BASE[sp]['second_is_src'],BASE[sp]['train_pair_template'],NOC[sp])
        d=dict(joint=x['joint'],min_group=x['min_group'])
        for ctl in ['nocue','empty','perm']:
            d['vs_'+ctl]=x['joint']-R['results']['%s_s%d'%(ctl,s)]['final'][sp]['joint']
        d['vs_best_baseline']=x['joint']-b; det[sp]=d
        ok=ok and x['joint']>=.90 and x['min_group']>=.80 and all(d['vs_'+c]>=.30 for c in ['nocue','empty','perm']) and d['vs_best_baseline']>=.30
    det['ok']=bool(ok); adv[s]=det
R['advance']=adv; R['advance_all']=bool(all(a['ok'] for a in adv.values())); R['wall']=time.perf_counter()-T0
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): R['rss_mib']=int(l.split()[1])/1024
json.dump(R,open(OUT+'/result.json','w'),indent=1); print('done',R['advance_all'])
