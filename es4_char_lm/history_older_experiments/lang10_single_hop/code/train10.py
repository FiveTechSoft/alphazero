import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')

import sys,time,json,os,pickle,hashlib,numpy as np,torch
torch.set_num_threads(1)
sys.path.insert(0,(EXP_ROOT+'/lang10')); sys.path.insert(0,(EXP_ROOT+'/lang9')); sys.path.insert(0,(EXP_ROOT+'/lang8'))
import lang7 as L, gt as G, gi, memguard, evalb, data10 as D
T0=time.perf_counter(); arm=sys.argv[1]; seed=int(sys.argv[2]); STEPS=int(sys.argv[3]); runid=sys.argv[4]; selftest=len(sys.argv)>5 and sys.argv[5]=='selftest'
assert arm in ('ctx','ctl'); tag='%s_s%d'%(arm,seed); OUT=((EXP_ROOT+'/selftest_st10_') if selftest else (EXP_ROOT+'/lang10/runs/'))+runid; os.makedirs(OUT,exist_ok=False); memguard.check('start')
def guard():
    if time.perf_counter()-T0>85: raise RuntimeError('internal wall guard')
ES=np.load((EXP_ROOT+'/lang10/eval10.npz')); SPL=['TR','E_b','E_s','E_bs']
def greedy(m,x): return evalb.greedy_batched(m,x,128)
def evaluate(m):
    out={};P={}
    for k in SPL:
        g=ES[k+'_gold']; grp=ES[k+'_grp']; tgt=D.target(g); r={}
        for nm,xk in (('ctx',k+'_x'),('nocontext',k+'_xnc')):
            fr=greedy(m,ES[xk]); ok=(fr==tgt).all(1); P[k+'_'+nm]=fr
            inv=~((fr[:,0]>=G.C0)&(fr[:,0]<G.C0+4)&(fr[:,1]>=G.C0)&(fr[:,1]<G.C0+4)&(fr[:,2]==G.STOP)); ga=[float(ok[grp==u].mean()) for u in np.unique(grp)]
            r[nm]=dict(exact=float(ok.mean()),invalid=float(inv.mean()),first_tok=float((fr[:,0]==tgt[:,0]).mean()),groups=len(ga),min_group=min(ga),groups_ge80=int(sum(a>=.8 for a in ga)))
        p0=P[k+'_ctx']; p2=greedy(m,ES[k+'_xcf']); gc=D.target(ES[k+'_goldcf']); P[k+'_cf']=p2
        r['cf_support']=dict(exact_new=float((p2==gc).all(1).mean()),pred_equals_old_gold=float((p2[:,:2]==D.target(ES[k+'_gold'])[:,:2]).all(1).mean()),pred_changed=float((p2!=p0).any(1).mean()),correct_change=float((((p2!=p0).any(1))&((p2==gc).all(1))).mean()))
        out[k]=r
    return out,P
m=gi.make(seed); torch.save(m.state_dict(),OUT+'/init_'+tag+'.pt')
opt=torch.optim.AdamW(m.parameters(),lr=1e-3,betas=(0.9,0.999),weight_decay=0.0)
rng=np.random.RandomState(1000+seed); prng=np.random.RandomState(5000+seed); acc=0.0; curve=[]; first=None; hs=hashlib.sha256()
for step in range(STEPS):
    A=D.eps(rng,L.TRB,L.TRS,32)
    for key in ('sw','b','s','d'): hs.update(np.ascontiguousarray(A[key]).tobytes())
    sc=D.true_sc(A); g=D.gold(A)
    scu=D.shuffled_sc(A,prng,L.TRB) if arm=='ctl' else sc
    x=D.prompt(A,scu)
    if arm=='ctx' and not (D.oracle(x,A)==g).all(): raise RuntimeError('train gold mismatch')
    assert x.shape==(32,25) and int(x.max())<G.NV
    xt=torch.from_numpy(x); tg=torch.from_numpy(D.target(g))
    for gp in opt.param_groups: gp['lr']=1e-3*min(1,(step+1)/200)
    lo=G.loss_tf(m,xt,tg)
    if not torch.isfinite(lo): raise RuntimeError('nonfinite')
    opt.zero_grad(set_to_none=True); lo.backward(); torch.nn.utils.clip_grad_norm_(m.parameters(),1.0); opt.step(); memguard.check('train step %d'%step)
    v=float(lo.detach()); acc+=v; first=v if first is None else first
    if (step+1)%500==0 or (selftest and step+1==STEPS): curve.append(acc/(500 if (step+1)%500==0 else STEPS)); acc=0.0; guard()
torch.save(m.state_dict(),OUT+'/final_'+tag+'.pt'); torch.save(opt.state_dict(),OUT+'/adam_'+tag+'.pt')
pickle.dump(dict(rng=rng.get_state(),prng=prng.get_state(),torch_rng=torch.get_rng_state().numpy()),open(OUT+'/rng_'+tag+'.pkl','wb'))
m.eval(); res,P=evaluate(m); np.savez(OUT+'/pred_'+tag+'.npz',**P)
m2=gi.make(0); m2.load_state_dict(torch.load(OUT+'/final_'+tag+'.pt')); m2.eval(); _,Q=evaluate(m2); parity=bool(all((P[k]==Q[k]).all() for k in P))
memguard.check('after last eval/parity batch'); _rss,_hwm=memguard.usage()
sha=lambda f:hashlib.sha256(open(f,'rb').read()).hexdigest()
r=dict(tag=tag,runid=runid,arm=arm,steps=STEPS,selftest=selftest,stream_hash=hs.hexdigest(),final=res,loss_first=first,loss_curve=curve,parity=parity,wall=time.perf_counter()-T0,rss_mib=_hwm,rss_now_mib=_rss,eval_batch=128,
       sha256={f:sha(f) for f in [(EXP_ROOT+'/lang8/lang7.py'),(EXP_ROOT+'/lang8/gt.py'),(EXP_ROOT+'/lang8/gi.py'),(EXP_ROOT+'/lang9/memguard.py'),(EXP_ROOT+'/lang9/evalb.py'),(EXP_ROOT+'/lang10/data10.py'),(EXP_ROOT+'/lang10/eval10.npz'),os.path.abspath(__file__)]},
       init_sha=sha(OUT+'/init_'+tag+'.pt'),final_sha=sha(OUT+'/final_'+tag+'.pt'))
json.dump(r,open(OUT+'/res.json','w'),indent=1); print('done',tag,runid,{k:round(v['ctx']['exact'],3) for k,v in res.items()},round(curve[-1],4))
