import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import sys,time,json,os,pickle,hashlib,numpy as np,torch
torch.set_num_threads(1)
sys.path.insert(0,(EXP_ROOT+'/lang11')); sys.path.insert(0,(EXP_ROOT+'/lang9')); sys.path.insert(0,(EXP_ROOT+'/lang8'))
import gt as G, gi, memguard, evalb, data11 as D
T0=time.perf_counter(); seed=int(sys.argv[1]); STEPS=1000; tag='s%d'%seed; OUT=(EXP_ROOT+'/lang12/runs/o12_')+tag; os.makedirs(OUT,exist_ok=False); memguard.check('start')
pool=np.load((EXP_ROOT+'/lang11/part/pool.npy')).astype(np.int64); N=len(pool)
r0=np.random.RandomState(12012); idx=r0.randint(N,size=32); ar=np.arange(32)
pr=np.stack([r0.permutation(4) for _ in range(32)]); qp=np.array([D.PAIRS12[i] for i in r0.randint(12,size=32)])
W=pool[idx][ar[:,None],pr,0]; C=pool[idx][ar[:,None],pr,1]; qa=W[ar,qp[:,0]]; qb=W[ar,qp[:,1]]; gold=np.stack([C[ar,qp[:,0]],C[ar,qp[:,1]]],1)
x=D.prompt(W,C,qa,qb)
if not (D.oracle(x)==gold).all(): raise RuntimeError('gold mismatch')
tg_np=D.target(gold); xt=torch.from_numpy(x); tg=torch.from_numpy(tg_np); bsha=hashlib.sha256(x.tobytes()+tg_np.tobytes()).hexdigest()
np.savez(OUT+'/batch.npz',x=x,t=tg_np,idx=idx)
def ex(m):
    m.eval(); fr=evalb.greedy_batched(m,x,128); return fr,float((fr==tg_np).all(1).mean())
m=gi.make(seed); torch.save(m.state_dict(),OUT+'/init_'+tag+'.pt')
fr0,e0=ex(m); m.train()
opt=torch.optim.AdamW(m.parameters(),lr=1e-3,betas=(0.9,0.999),weight_decay=0.0)
losses=[];gn=[];un=[];fin=True;t=time.perf_counter()
for step in range(STEPS):
    for gp in opt.param_groups: gp['lr']=1e-3*min(1,(step+1)/200)
    lo=G.loss_tf(m,xt,tg)
    if not torch.isfinite(lo): raise RuntimeError('nonfinite loss')
    opt.zero_grad(set_to_none=True); lo.backward(); g=float(torch.nn.utils.clip_grad_norm_(m.parameters(),1.0))
    if not np.isfinite(g): raise RuntimeError('nonfinite grad')
    before=[q.detach().clone() for q in m.parameters()]; opt.step()
    un.append(float(torch.sqrt(sum(((q.detach()-b)**2).sum() for q,b in zip(m.parameters(),before)))))
    losses.append(float(lo.detach())); gn.append(g); memguard.check('step %d'%step)
ttrain=time.perf_counter()-t
torch.save(m.state_dict(),OUT+'/final_'+tag+'.pt'); torch.save(opt.state_dict(),OUT+'/adam_'+tag+'.pt')
pickle.dump(dict(torch_rng=torch.get_rng_state().numpy(),np_rng=r0.get_state()),open(OUT+'/rng_'+tag+'.pkl','wb'))
fr1,e1=ex(m)
with torch.no_grad(): fl=float(G.loss_tf(m,xt,tg))
m2=gi.make(0); m2.load_state_dict(torch.load(OUT+'/final_'+tag+'.pt')); fr2,e2=ex(m2); parity=bool((fr1==fr2).all())
memguard.check('after reload'); _rss,_hwm=memguard.usage()
sha=lambda f:hashlib.sha256(open(f,'rb').read()).hexdigest()
r=dict(tag=tag,seed=seed,steps=STEPS,batch_sha=bsha,exact_before=e0,exact_after=e1,exact_after_reload=e2,parity=parity,final_loss_eval=fl,loss_first=losses[0],loss_last=losses[-1],loss_mean_last50=float(np.mean(losses[-50:])),
 loss_every50=[losses[i] for i in range(49,STEPS,50)],gradnorm_first=gn[0],gradnorm_max=max(gn),gradnorm_last=gn[-1],updnorm_first=un[0],updnorm_max=max(un),updnorm_last=un[-1],all_finite=True,
 t_train=ttrain,wall=time.perf_counter()-T0,rss_mib=_hwm,rss_now_mib=_rss,eval_batch=128,
 sha256={f:sha(f) for f in [(EXP_ROOT+'/lang8/gt.py'),(EXP_ROOT+'/lang8/gi.py'),(EXP_ROOT+'/lang11/data11.py'),(EXP_ROOT+'/lang11/part/pool.npy'),os.path.abspath(__file__)]},init_sha=sha(OUT+'/init_'+tag+'.pt'),final_sha=sha(OUT+'/final_'+tag+'.pt'))
np.savez(OUT+'/curves_'+tag+'.npz',loss=np.array(losses),gn=np.array(gn),un=np.array(un))
json.dump(r,open(OUT+'/res.json','w'),indent=1); print('done',tag,e0,e1,fl)
