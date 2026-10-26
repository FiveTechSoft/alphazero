import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import sys,time,json,os,pickle,hashlib,numpy as np,torch
torch.set_num_threads(1)
import lang7 as L, gt as G, gi
T0=time.perf_counter(); seed=int(sys.argv[1]); arm=sys.argv[2]; STEPS=int(open('steps.txt').read().strip()); OUT=(EXP_ROOT+'/lang8/out'); tag='%s_s%d'%(arm,seed)
assert arm in ('full','nocue')
def guard():
    if time.perf_counter()-T0>88: raise RuntimeError('internal wall guard')
ES=np.load(OUT+'/eval_ser.npz'); names=list(L.SPECS)
m=gi.make(seed); torch.save(m.state_dict(),OUT+'/init_'+tag+'.pt')
opt=torch.optim.AdamW(m.parameters(),lr=1e-3,betas=(0.9,0.999),weight_decay=0.0)
rng=np.random.RandomState(1000+seed); prng=np.random.RandomState(2000+seed); acc=0.0; curve=[]; first=None
for step in range(STEPS):
    eps=[L.gen(rng,L.TRB,L.TRS,L.TRL,'train',prng) for _ in range(32)]
    for e in eps:
        if L.validate(e,L.TRB,L.TRS,L.TRL): raise RuntimeError('train validation')
    T,sw,sc,lay,gold=L.pack(eps,'full'); x=torch.from_numpy(gi.prep(G.ser(T,sw,sc,lay),arm)); tg=torch.from_numpy(G.target(gold))
    assert tuple(x.shape)==(32,25) and int(x.max())<G.NV
    for g in opt.param_groups: g['lr']=1e-3*min(1,(step+1)/200)
    lo=G.loss_tf(m,x,tg)
    if not torch.isfinite(lo): raise RuntimeError('nonfinite')
    opt.zero_grad(set_to_none=True); lo.backward(); torch.nn.utils.clip_grad_norm_(m.parameters(),1.0); opt.step()
    v=float(lo.detach()); acc+=v; first=v if first is None else first
    if (step+1)%500==0: curve.append(acc/500); acc=0.0; guard()
torch.save(m.state_dict(),OUT+'/final_'+tag+'.pt'); torch.save(opt.state_dict(),OUT+'/adam_'+tag+'.pt')
pickle.dump(dict(rng=rng.get_state(),prng=prng.get_state(),torch_rng=torch.get_rng_state().numpy()),open(OUT+'/rng_'+tag+'.pkl','wb'))
def evaluate(model):
    out={};preds={}
    for k in names:
        x0=ES[k+'_x']; gold=ES[k+'_gold']; grp=ES[k+'_grp']; fr=G.greedy(model,torch.from_numpy(gi.prep(x0,arm))).numpy()
        tgt=G.target(gold); ok=(fr==tgt).all(1); inv=~((fr[:,0]>=G.K0)&(fr[:,0]<G.K0+4)&(fr[:,1]>=G.K0)&(fr[:,1]<G.K0+4)&(fr[:,2]==G.STOP))
        ga=[float(ok[grp==u].mean()) for u in np.unique(grp)]
        out[k]=dict(free_exact=float(ok.mean()),invalid=float(inv.mean()),step_acc=[float((fr[:,i]==tgt[:,i]).mean()) for i in range(3)],groups=len(ga),min_group=min(ga),groups_ge80=int(sum(a>=.8 for a in ga)),n=int(len(ok)))
        preds[k]=fr
    return out,preds
def agree(preds,k):
    x=ES[k+'_x']; fr=preds[k]; res={'first_is_src':0,'second_is_src':0,'train_pair_template':0}
    for xi,p in zip(x,fr):
        col={int(xi[1+2*i]):int(xi[2+2*i])-G.C0 for i in range(4)}; lc=[int(xi[11+2*i])-G.C0 for i in range(4)]
        ws=[int(t) for t in xi[19:24] if t<4]; f,s=ws[0],ws[1]
        pol={'first_is_src':(f,s),'second_is_src':(s,f),'train_pair_template':(f,s) if frozenset((f,s)) in L.ORDER0 else (s,f)}
        for n,(a,b) in pol.items(): res[n]+=(G.K0+lc.index(col[a]),G.K0+lc.index(col[b]))==(int(p[0]),int(p[1]))
    return {n:v/len(x) for n,v in res.items()}
m.eval(); res,preds=evaluate(m); np.savez(OUT+'/pred_'+tag+'.npz',**preds)
m2=gi.make(0); m2.load_state_dict(torch.load(OUT+'/final_'+tag+'.pt')); _,p2=evaluate(m2); parity=bool(all((preds[k]==p2[k]).all() for k in preds))
r=dict(tag=tag,steps=STEPS,final=res,loss_first=first,loss_curve=curve,parity=parity,agree_E3=agree(preds,'E3'),agree_E5=agree(preds,'E5'),wall=time.perf_counter()-T0)
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): r['rss_mib']=int(l.split()[1])/1024
json.dump(r,open(OUT+'/res_'+tag+'.json','w'),indent=1); print('done',tag,{k:round(v['free_exact'],3) for k,v in res.items()},curve[-1])
