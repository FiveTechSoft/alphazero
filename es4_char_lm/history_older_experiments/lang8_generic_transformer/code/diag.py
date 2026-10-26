import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import time,json,sys,hashlib,platform
import numpy as np
T0=time.perf_counter(); res={}
import torch
torch.set_num_threads(1); torch.manual_seed(0)
import lang7 as L, gt as G
res['versions']=dict(python=platform.python_version(),torch=torch.__version__,numpy=np.__version__,threads=torch.get_num_threads(),interop=torch.get_num_interop_threads())
res['code_sha']={f:hashlib.sha256(open(f,'rb').read()).hexdigest() for f in ['lang7.py','gt.py','diag.py']}
EV=L.build_eval(3101); fz=np.load((EXP_ROOT+'/lang7e/out/eval_frozen.npz')); same=True; out={}; chk={}
for k,eps in EV.items():
    T,sw,sc,lay,gold=L.pack(eps,'full'); same&=bool((fz[k+'_gold']==gold).all() and (fz[k+'_T']==T).all() and (fz[k+'_sw']==sw).all() and (fz[k+'_sc']==sc).all() and (fz[k+'_lay']==lay).all())
    x=G.ser(T,sw,sc,lay); t2,sw2,sc2,lay2=G.unser(x)
    rt=bool((t2==T).all() and (sw2==sw).all() and (sc2==sc).all() and (lay2==lay).all())
    cells=bool((x[:,[10,12,14,16]]==G.K0+np.arange(4)).all()); rng_ok=bool(x.min()>=0 and x.max()<G.NV and x.shape[1]==G.PL)

    ok=0
    for xi,gi in zip(x,gold):
        col={int(xi[1+2*i]):int(xi[2+2*i])-G.C0 for i in range(4)}; q=list(xi[19:24]); src=dst=None
        for j,t in enumerate(q):
            if t==G.DE: src=int(q[j+1])
            if t==G.A: dst=int(q[j+1])
        lc=[int(xi[11+2*i])-G.C0 for i in range(4)]; ok+=(lc.index(col[src]),lc.index(col[dst]))==tuple(gi)
    chk[k]=dict(roundtrip=rt,cells=cells,range=rng_ok,oracle_from_tokens=ok/len(x),n=len(x)); out[k]=(x,gold,np.array([e['grp'] for e in eps]))
    if not (rt and cells and rng_ok and ok==len(x)): res['serialization_checks']=chk; res['STOP']='serialization'; json.dump(res,open('diag.json','w'),indent=1); raise SystemExit
res['eval_equals_lang7e_frozen']=same; res['serialization_checks']=chk
if not same: res['STOP']='eval_mismatch'; json.dump(res,open('diag.json','w'),indent=1); raise SystemExit
np.savez('out/eval_ser.npz',**{k+'_x':v[0] for k,v in out.items()},**{k+'_gold':v[1] for k,v in out.items()},**{k+'_grp':v[2] for k,v in out.items()})

rng=np.random.RandomState(99); prng=np.random.RandomState(98); batches=[]
for i in range(60):
    eps=[L.gen(rng,L.TRB,L.TRS,L.TRL,'train',prng) for _ in range(32)]
    for e in eps:
        if L.validate(e,L.TRB,L.TRS,L.TRL): raise RuntimeError('train validation')
    T,sw,sc,lay,gold=L.pack(eps,'full'); batches.append((torch.from_numpy(G.ser(T,sw,sc,lay)),torch.from_numpy(G.target(gold))))
res['lengths']=dict(prompt=G.PL,train_input_positions=G.PL+2,generated=3,max_positions=G.ML,vocab=G.NV)
m=G.GT(64,4,2); res['model']=dict(d=64,heads=4,layers=2,mlp=256,params=int(sum(p.numel() for p in m.parameters())),tied_head=True,causal=True,positional='learned')

with torch.no_grad():
    x=torch.cat([batches[0][0],batches[0][1][:,:2]],1); l1=m(x); x2=x.clone(); x2[:,-1]=(x2[:,-1]+1)%G.NV; l2=m(x2)
res['contract']=dict(logits_shape=list(l1.shape),causal_max_diff_earlier=float((l1[:,:-1]-l2[:,:-1]).abs().max()),last_changes=bool((l1[:,-1]-l2[:,-1]).abs().max()>0),finite=bool(torch.isfinite(l1).all()))
if res['contract']['causal_max_diff_earlier']>1e-6 or not res['contract']['finite'] or list(l1.shape)!=[32,27,G.NV]: res['STOP']='contract'; json.dump(res,open('diag.json','w'),indent=1); raise SystemExit
tm=[]
for i,(p,t) in enumerate(batches):
    m.zero_grad(set_to_none=True); a=time.perf_counter(); lo=G.loss_tf(m,p,t); lo.backward(); b=time.perf_counter(); tm.append(b-a)
    if not torch.isfinite(lo): raise RuntimeError('nonfinite')
res['fwd_bwd_batch32']=dict(median_s=float(np.median(tm[5:])),min_s=float(min(tm[5:])),max_s=float(max(tm[5:])),n=len(tm)-5,init_loss=float(lo),note='random weights, no optimizer update')
t1=time.perf_counter(); fp={k:G.greedy(m,torch.from_numpy(v[0])) for k,v in out.items()}; res['greedy_all_eval_splits_s']=time.perf_counter()-t1
res['greedy_shape']={k:list(v.shape) for k,v in fp.items()}; res['wall']=time.perf_counter()-T0
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): res['rss_mib']=int(l.split()[1])/1024
json.dump(res,open('diag.json','w'),indent=1); print('done',res['fwd_bwd_batch32']['median_s'])
