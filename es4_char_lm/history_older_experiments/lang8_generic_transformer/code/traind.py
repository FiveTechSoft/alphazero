import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import sys,time,json,os,pickle,numpy as np,torch
torch.set_num_threads(1)
import lang7 as L, gt as G, gi, dtasks as D
T0=time.perf_counter(); task=sys.argv[1]; seed=int(sys.argv[2]); STEPS=int(open('steps.txt').read().strip()); OUT=(EXP_ROOT+'/lang8/out'); tag='%s_s%d'%(task,seed)
assert task in ('d1','d2')
def guard():
    if time.perf_counter()-T0>90: raise RuntimeError('internal wall guard')
ES=np.load(OUT+'/eval_d.npz'); SPL=['TR','E_b','E_l','E_bl']
m=gi.make(seed); torch.save(m.state_dict(),OUT+'/init_'+tag+'.pt')
opt=torch.optim.AdamW(m.parameters(),lr=1e-3,betas=(0.9,0.999),weight_decay=0.0)
rng=np.random.RandomState(1000+seed); acc=0.0; curve=[]; first=None
for step in range(STEPS):
    eps=[L.gen(rng,L.TRB,L.TRS,L.TRL,'train',None) for _ in range(32)]
    for e in eps:
        if L.validate(e,L.TRB,L.TRS,L.TRL): raise RuntimeError('train validation')
    A=D.arrays(eps); x=D.prompts(task,A)
    if not (D.oracle(x,task)==A['gold']).all(): raise RuntimeError('train gold mismatch')
    assert x.shape==(32,25) and int(x.max())<G.NV
    xt=torch.from_numpy(x); tg=torch.from_numpy(G.target(A['gold']))
    for g in opt.param_groups: g['lr']=1e-3*min(1,(step+1)/200)
    lo=G.loss_tf(m,xt,tg)
    if not torch.isfinite(lo): raise RuntimeError('nonfinite')
    opt.zero_grad(set_to_none=True); lo.backward(); torch.nn.utils.clip_grad_norm_(m.parameters(),1.0); opt.step()
    v=float(lo.detach()); acc+=v; first=v if first is None else first
    if (step+1)%500==0: curve.append(acc/500); acc=0.0; guard()
torch.save(m.state_dict(),OUT+'/final_'+tag+'.pt'); torch.save(opt.state_dict(),OUT+'/adam_'+tag+'.pt')
pickle.dump(dict(rng=rng.get_state(),torch_rng=torch.get_rng_state().numpy()),open(OUT+'/rng_'+tag+'.pkl','wb'))
def greedy(model,x): return G.greedy(model,torch.from_numpy(x)).numpy()
def evaluate(model):
    out={};preds={}
    for k in SPL:
        x=ES[task+'_'+k+'_x']; gold=ES['gold_'+k]; grp=ES['grp_'+k]; fr=greedy(model,x); tgt=G.target(gold); ok=(fr==tgt).all(1)
        inv=~((fr[:,0]>=G.K0)&(fr[:,0]<G.K0+4)&(fr[:,1]>=G.K0)&(fr[:,1]<G.K0+4)&(fr[:,2]==G.STOP)); ga=[float(ok[grp==u].mean()) for u in np.unique(grp)]
        out[k]=dict(free_exact=float(ok.mean()),invalid=float(inv.mean()),step_acc=[float((fr[:,i]==tgt[:,i]).mean()) for i in range(3)],groups=len(ga),min_group=min(ga),groups_ge80=int(sum(a>=.8 for a in ga)),n=int(len(ok)))
        preds[k]=fr
    return out,preds
def cf(model,preds):
    out={};cp={}
    for kind,lab in (('cfL','layout'),('cfS','support')):
        for k in ('TR','E_bl'):
            x2=ES[kind+'_'+task+'_'+k+'_x']; g2=ES[kind+'_'+task+'_'+k+'_gold']; g0=ES['gold_'+k]; p0=preds[k]; p2=greedy(model,x2)
            c0=G.K0+g0; c2=G.K0+g2; same_gold=bool((g2==g0).all())
            r=dict(n=int(len(x2)),exact_new=float((p2[:,:2]==c2).all(1).mean()),member_new=[float((p2[:,i]==c2[:,i]).mean()) for i in range(2)],pred_equals_old_gold=float((p2[:,:2]==c0).all(1).mean()),
                   orig_exact_old=float((p0[:,:2]==c0).all(1).mean()),pred_changed=float((p2!=p0).any(1).mean()),gold_unchanged_by_construction=same_gold)
            if not same_gold: r['correct_change']=float((((p2!=p0).any(1))&((p2[:,:2]==c2).all(1))).mean())
            out[lab+'_'+k]=r; cp[lab+'_'+k]=p2
    return out,cp
m.eval(); res,preds=evaluate(m); cres,cpreds=cf(m,preds); np.savez(OUT+'/pred_'+tag+'.npz',**preds,**{'cf_'+k:v for k,v in cpreds.items()})
m2=gi.make(0); m2.load_state_dict(torch.load(OUT+'/final_'+tag+'.pt')); _,p2=evaluate(m2); parity=bool(all((preds[k]==p2[k]).all() for k in preds))
r=dict(tag=tag,steps=STEPS,final=res,counterfactual=cres,loss_first=first,loss_curve=curve,parity=parity,wall=time.perf_counter()-T0)
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): r['rss_mib']=int(l.split()[1])/1024
json.dump(r,open(OUT+'/res_'+tag+'.json','w'),indent=1); print('done',tag,{k:round(v['free_exact'],3) for k,v in res.items()},round(curve[-1],4))
