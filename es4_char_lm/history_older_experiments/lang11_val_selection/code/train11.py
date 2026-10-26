import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')

import sys,time,json,os,pickle,hashlib,numpy as np,torch
torch.set_num_threads(1)
sys.path.insert(0,(EXP_ROOT+'/lang11')); sys.path.insert(0,(EXP_ROOT+'/lang9')); sys.path.insert(0,(EXP_ROOT+'/lang8'))
import gt as G, gi, memguard, evalb, data11 as D
T0=time.perf_counter(); arm=sys.argv[1]; seed=int(sys.argv[2]); STEPS=int(sys.argv[3]); ROUND=int(sys.argv[4]); runid=sys.argv[5]; selftest=len(sys.argv)>6 and sys.argv[6]=='selftest'
assert arm in ('A','U','S'); tag='%s_s%d'%(arm,seed); OUT=((EXP_ROOT+'/selftest_st11_') if selftest else (EXP_ROOT+'/lang11/runs/'))+runid; os.makedirs(OUT,exist_ok=False); memguard.check('start')
def guard():
    if time.perf_counter()-T0>85: raise RuntimeError('internal wall guard')
pool=np.load((EXP_ROOT+'/lang11/part/pool.npy')).astype(np.int64); N=len(pool); AI=D.assoc_index(pool)
Z=np.load((EXP_ROOT+'/lang11/rev/val_prompts11.npz')); Wv,Cv,qav,qbv,gv=[Z[k].astype(np.int64) for k in ('W','C','qa','qb','gold')]
xv=D.prompt(Wv,Cv,qav,qbv); tv=D.target(gv); SP=json.load(open((EXP_ROOT+'/lang11/rev/S_perms11.json')))[str(seed)]
led=dict(t_train=0.0,t_val=0.0,t_weights=0.0,val_eval_prompts=0,weight_computations=0,train_prompts=0,updates=0)
def val_eval(m):
    t=time.perf_counter(); fr=evalb.greedy_batched(m,xv,128); led['t_val']+=time.perf_counter()-t; led['val_eval_prompts']+=len(xv); return fr
def assoc_err(fr):
    err=np.zeros(36); cnt=np.zeros(36)
    for j in (0,1):
        a=(qav if j==0 else qbv)*D.NC+gv[:,j]; np.add.at(cnt,a,1); np.add.at(err,a,(fr[:,j]!=tv[:,j]).astype(float))
    return err/np.maximum(cnt,1),cnt
m=gi.make(seed); torch.save(m.state_dict(),OUT+'/init_'+tag+'.pt')
opt=torch.optim.AdamW(m.parameters(),lr=1e-3,betas=(0.9,0.999),weight_decay=0.0)
ctx_rng=np.random.RandomState(1000+seed); sup_rng=np.random.RandomState(2000+seed); q_rng=np.random.RandomState(3000+seed)
p=np.full(N,1.0/N); cdf=np.cumsum(p); acc=0.0; curve=[]; first=None; rounds=[]; PW=[p.copy()]; PC=[np.zeros(N,dtype=np.int32)]; ESS=[D.ess_ratio(p)]
for step in range(STEPS):
    t=time.perf_counter()
    idx=np.minimum(np.searchsorted(cdf,ctx_rng.random_sample(32),side='right'),N-1); PC[-1]+=np.bincount(idx,minlength=N).astype(np.int32)
    pr=np.stack([sup_rng.permutation(4) for _ in range(32)]); qp=np.array([D.PAIRS12[i] for i in q_rng.randint(12,size=32)]); ar=np.arange(32)
    W=pool[idx][ar[:,None],pr,0]; C=pool[idx][ar[:,None],pr,1]; qa=W[ar,qp[:,0]]; qb=W[ar,qp[:,1]]; gold=np.stack([C[ar,qp[:,0]],C[ar,qp[:,1]]],1)
    x=D.prompt(W,C,qa,qb)
    if not (D.oracle(x)==gold).all(): raise RuntimeError('train gold mismatch')
    assert x.shape==(32,25) and int(x.max())<G.NV
    xt=torch.from_numpy(x); tg=torch.from_numpy(D.target(gold))
    for gp in opt.param_groups: gp['lr']=1e-3*min(1,(step+1)/200)
    lo=G.loss_tf(m,xt,tg)
    if not torch.isfinite(lo): raise RuntimeError('nonfinite')
    opt.zero_grad(set_to_none=True); lo.backward(); torch.nn.utils.clip_grad_norm_(m.parameters(),1.0); opt.step(); memguard.check('train step %d'%step)
    v=float(lo.detach()); acc+=v; first=v if first is None else first; led['t_train']+=time.perf_counter()-t; led['updates']+=1; led['train_prompts']+=32
    if (step+1)%500==0 or (selftest and (step+1)%ROUND==0): curve.append(acc/(500 if (step+1)%500==0 else ROUND)); acc=0.0
    if (step+1)%ROUND==0 and (step+1)<STEPS:
        r=(step+1)//ROUND; m.eval(); fr=val_eval(m); m.train(); e,cn=assoc_err(fr); tw=time.perf_counter()
        if arm=='U': p=np.full(N,1.0/N); _=D.weights(AI,e)
        elif arm=='A': p=D.weights(AI,e)
        else: p=D.weights(AI,e[np.array(SP[r-1])])
        cdf=np.cumsum(p); led['t_weights']+=time.perf_counter()-tw; led['weight_computations']+=1
        rounds.append(dict(round=r,update=step+1,val_exact=float((fr==tv).all(1).mean()),val_tok_acc=[float((fr[:,j]==tv[:,j]).mean()) for j in range(3)],e_assoc=e.tolist(),n_assoc=cn.tolist(),ess_ratio=D.ess_ratio(p),wmax_over_wmin=float(p.max()/p.min())))
        PW.append(p.copy()); PC.append(np.zeros(N,dtype=np.int32)); ESS.append(D.ess_ratio(p)); guard()
torch.save(m.state_dict(),OUT+'/final_'+tag+'.pt'); torch.save(opt.state_dict(),OUT+'/adam_'+tag+'.pt')
pickle.dump(dict(ctx=ctx_rng.get_state(),sup=sup_rng.get_state(),q=q_rng.get_state(),torch_rng=torch.get_rng_state().numpy()),open(OUT+'/rng_'+tag+'.pkl','wb'))
m.eval(); fr=val_eval(m); e,cn=assoc_err(fr); np.savez(OUT+'/val_raw_'+tag+'.npz',pred=fr,gold=tv,weights=np.stack(PW),practice_counts=np.stack(PC),ess=np.array(ESS),e_final=e)
m2=gi.make(0); m2.load_state_dict(torch.load(OUT+'/final_'+tag+'.pt')); m2.eval(); fr2=val_eval(m2); parity=bool((fr==fr2).all())
memguard.check('after last eval/parity batch'); _rss,_hwm=memguard.usage()
sha=lambda f:hashlib.sha256(open(f,'rb').read()).hexdigest()
r=dict(tag=tag,runid=runid,arm=arm,seed=seed,steps=STEPS,round_every=ROUND,selftest=selftest,val_final=dict(exact=float((fr==tv).all(1).mean()),tok_acc=[float((fr[:,j]==tv[:,j]).mean()) for j in range(3)],note='selection set only, not primary evidence'),
  rounds=rounds,loss_first=first,loss_curve=curve,parity=parity,ledger=led,wall=time.perf_counter()-T0,rss_mib=_hwm,rss_now_mib=_rss,eval_batch=128,
  sha256={f:sha(f) for f in [(EXP_ROOT+'/lang8/gt.py'),(EXP_ROOT+'/lang8/gi.py'),(EXP_ROOT+'/lang9/memguard.py'),(EXP_ROOT+'/lang9/evalb.py'),(EXP_ROOT+'/lang11/data11.py'),(EXP_ROOT+'/lang11/part/pool.npy'),(EXP_ROOT+'/lang11/rev/val_prompts11.npz'),(EXP_ROOT+'/lang11/rev/S_perms11.json'),os.path.abspath(__file__)]},
  init_sha=sha(OUT+'/init_'+tag+'.pt'),final_sha=sha(OUT+'/final_'+tag+'.pt'))
json.dump(r,open(OUT+'/res.json','w'),indent=1); print('done',tag,runid,round(r['val_final']['exact'],3),round(curve[-1],4))
