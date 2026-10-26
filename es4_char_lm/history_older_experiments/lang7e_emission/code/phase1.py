import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import numpy as np,time,json,os,pickle,hashlib
import lang7 as L, lang7e as M
T0=time.perf_counter()
def guard():
    if time.perf_counter()-T0>128: raise RuntimeError('internal wall guard')
OUT=(EXP_ROOT+'/lang7e/out')
STEPS=int(open('steps.txt').read().strip()); LR=3e-3; BS=32; SEEDS=[21,22]; ARMS=['full','nocue','noprev']
EV=L.build_eval(3101)
cur=hashlib.sha256(open(OUT+'/eval_frozen.npz','rb').read()).hexdigest()
fz=np.load(OUT+'/eval_frozen.npz')
for k,eps in EV.items():
    assert (fz[k+'_gold']==np.array([e['gold'] for e in eps])).all() and (fz[k+'_T']==np.stack([e['T'] for e in eps])).all()
BASE={k:L.baselines(v) for k,v in EV.items()}
NOC={k:L.noctx_bound((L.SPECS[k][0],L.SPECS[k][1],L.SPECS[k][2],L.SPECS[k][5])) for k in L.SPECS}
def evaluate(P,arm,useprev,support='full',names=None):
    out={};preds={}
    pk='nocue' if arm=='nocue' else support
    for name,eps in EV.items():
        if names and name not in names: continue
        T,sw,sc,lay,gold=L.pack(eps,pk); fr=M.free_pred(P,T,sw,sc,lay,useprev); tf=M.tf_pred(P,T,sw,sc,lay,gold,useprev)
        tgt=np.concatenate([gold,np.full((len(gold),1),M.STOP)],1); ok=(fr==tgt).all(1); okt=(tf==tgt).all(1)
        inv=~((fr[:,0]<4)&(fr[:,1]<4)&(fr[:,2]==M.STOP)); g=np.array([e['grp'] for e in eps]); ga=[float(ok[g==u].mean()) for u in np.unique(g)]
        out[name]=dict(free_exact=float(ok.mean()),tf_exact=float(okt.mean()),invalid=float(inv.mean()),step_acc=[float((fr[:,i]==tgt[:,i]).mean()) for i in range(3)],
                       groups=len(ga),min_group=min(ga),groups_ge80=int(sum(a>=.8 for a in ga)),n=int(len(ok)))
        preds[name+'_free']=fr; preds[name+'_tf']=tf
    return out,preds
def agree(preds,name):
    eps=EV[name]; pr=preds[name+'_free']; res={}
    pol={'first_is_src':lambda f,s:(f,s),'second_is_src':lambda f,s:(s,f),'train_pair_template':lambda f,s:(f,s) if frozenset((f,s)) in L.ORDER0 else (s,f)}
    for k,fn in pol.items():
        ok=0
        for e,p in zip(eps,pr):
            words=[t for t in e['T'] if t<4]; src,dst=fn(int(words[0]),int(words[1])); ok+=(L.answer(e['b'],(src,dst),e['l'])==(int(p[0]),int(p[1])))
        res[k]=ok/len(eps)
    return res
def run(seed,arm):
    tag='%s_s%d'%(arm,seed); useprev=(arm!='noprev'); pk='nocue' if arm=='nocue' else 'full'
    P=M.init(seed); np.savez(OUT+'/init_'+tag+'.npz',**P)
    rng=np.random.RandomState(1000+seed); prng=np.random.RandomState(2000+seed)
    m={k:np.zeros_like(v) for k,v in P.items()}; v2={k:np.zeros_like(v) for k,v in P.items()}; t=0;acc=0.0;curve=[]
    for step in range(STEPS):
        eps=[L.gen(rng,L.TRB,L.TRS,L.TRL,'train',prng) for _ in range(BS)]
        for e in eps:
            if L.validate(e,L.TRB,L.TRS,L.TRL): raise RuntimeError('train validation')
        T,sw,sc,lay,gold=L.pack(eps,pk)
        assert T.shape==(BS,5) and T.max()<=6 and sw.shape==(BS,4) and sc.max()<=4
        loss,c=M.train_loss(P,T,sw,sc,lay,gold,useprev)
        if not np.isfinite(loss): raise RuntimeError('nonfinite')
        g=M.bwd(P,c); t+=1
        for k in P:
            m[k]=0.9*m[k]+0.1*g[k]; v2[k]=0.999*v2[k]+0.001*g[k]**2
            P[k]-=(LR*(m[k]/(1-0.9**t))/(np.sqrt(v2[k]/(1-0.999**t))+1e-8)).astype(np.float32)
        acc+=float(loss)
        if (step+1)%500==0: curve.append(acc/500); acc=0.0; guard()
    np.savez(OUT+'/final_'+tag+'.npz',**P); np.savez(OUT+'/adam_'+tag+'.npz',t=t,**{'m_'+k:v for k,v in m.items()},**{'v_'+k:v for k,v in v2.items()})
    pickle.dump(dict(rng=rng.get_state(),prng=prng.get_state()),open(OUT+'/rng_'+tag+'.pkl','wb'))
    res,preds=evaluate(P,arm,useprev); np.savez(OUT+'/pred_'+tag+'.npz',**preds)
    Lw=dict(np.load(OUT+'/final_'+tag+'.npz')); _,p2=evaluate(Lw,arm,useprev); parity=bool(all((preds[k]==p2[k]).all() for k in preds))
    r=dict(final=res,loss_curve=curve,parity=parity,agree_E3=agree(preds,'E3'),agree_E5=agree(preds,'E5'))
    if arm=='full':
        r['ood_empty_support']=evaluate(P,arm,useprev,'empty',['TR','E3'])[0]; r['ood_perm_support']=evaluate(P,arm,useprev,'perm',['TR','E3'])[0]
    return r
R=dict(eval_sha=cur,steps=STEPS,baselines=BASE,noctx=NOC,results={})
for s in SEEDS:
    for a in ARMS:
        R['results']['%s_s%d'%(a,s)]=run(s,a); guard()
adv={}
for s in SEEDS:
    f=R['results']['full_s%d'%s]['final']; n=R['results']['nocue_s%d'%s]['final']; npv=R['results']['noprev_s%d'%s]['final']
    ok=f['TR']['free_exact']>=.99 and f['TR']['invalid']<=.05; det={'TR':f['TR']['free_exact']}
    for sp in ['E3','E5']:
        x=f[sp]; b=max(BASE[sp]['first_is_src'],BASE[sp]['second_is_src'],BASE[sp]['train_pair_template'])
        d=dict(free_exact=x['free_exact'],tf_exact=x['tf_exact'],min_group=x['min_group'],invalid=x['invalid'],vs_nocue=x['free_exact']-n[sp]['free_exact'],vs_best_baseline=x['free_exact']-b,noprev_free_exact=npv[sp]['free_exact'])
        det[sp]=d
        ok=ok and x['free_exact']>=.90 and x['min_group']>=.80 and x['invalid']<=.05 and d['vs_nocue']>=.30 and d['vs_best_baseline']>=.30
    det['ok']=bool(ok); adv[s]=det
R['advance']=adv; R['advance_all']=bool(all(a['ok'] for a in adv.values())); R['wall']=time.perf_counter()-T0
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): R['rss_mib']=int(l.split()[1])/1024
json.dump(R,open(OUT+'/result.json','w'),indent=1); print('done',R['advance_all'])
