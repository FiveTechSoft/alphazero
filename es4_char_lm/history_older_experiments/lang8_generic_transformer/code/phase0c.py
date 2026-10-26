import time,json,hashlib,numpy as np,torch
torch.set_num_threads(1)
import lang7 as L, gt as G, gi
T0=time.perf_counter(); res={}
res['code_sha']={f:hashlib.sha256(open(f,'rb').read()).hexdigest() for f in ['lang7.py','gt.py','gi.py','phase0c.py']}
EV=L.build_eval(3101)
res['baselines']={k:L.baselines(v) for k,v in EV.items()}
res['noctx']={k:L.noctx_bound((L.SPECS[k][0],L.SPECS[k][1],L.SPECS[k][2],L.SPECS[k][5])) for k in L.SPECS}
res['eval_ser_sha']=hashlib.sha256(open('out/eval_ser.npz','rb').read()).hexdigest()
ES=np.load('out/eval_ser.npz'); names=list(L.SPECS)
m=gi.make(0); lg=m(torch.from_numpy(ES['TR_x'][:32])); res['init_check']=dict(init_loss=float(G.loss_tf(m,torch.from_numpy(ES['TR_x'][:32]),torch.from_numpy(G.target(ES['TR_gold'][:32])))),ln20=float(np.log(20)))
rng=np.random.RandomState(555); prng=np.random.RandomState(556)
opt=torch.optim.AdamW(m.parameters(),lr=1e-3,betas=(0.9,0.999),weight_decay=0.0); tm=[]
for s in range(120):
    a=time.perf_counter(); eps=[L.gen(rng,L.TRB,L.TRS,L.TRL,'train',prng) for _ in range(32)]
    for e in eps:
        if L.validate(e,L.TRB,L.TRS,L.TRL): raise RuntimeError('train validation')
    T,sw,sc,lay,gold=L.pack(eps,'full'); x=torch.from_numpy(gi.prep(G.ser(T,sw,sc,lay),'full')); tg=torch.from_numpy(G.target(gold))
    for g in opt.param_groups: g['lr']=1e-3*min(1,(s+1)/200)
    lo=G.loss_tf(m,x,tg)
    if not torch.isfinite(lo): raise RuntimeError('nonfinite')
    opt.zero_grad(set_to_none=True); lo.backward(); torch.nn.utils.clip_grad_norm_(m.parameters(),1.0); opt.step(); lo.item(); tm.append(time.perf_counter()-a)
res['step_full_loop_s']=dict(median=float(np.median(tm[10:])),mean=float(np.mean(tm[10:])),max=float(max(tm[10:])),n=len(tm)-10)
t1=time.perf_counter()
for k in names: G.greedy(m,torch.from_numpy(ES[k+'_x']))
res['eval_pass_all_splits_s']=time.perf_counter()-t1
res['wall']=time.perf_counter()-T0
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): res['rss_mib']=int(l.split()[1])/1024
json.dump(res,open('cost8.json','w'),indent=1); print('done',res['step_full_loop_s'],res['eval_pass_all_splits_s'])
