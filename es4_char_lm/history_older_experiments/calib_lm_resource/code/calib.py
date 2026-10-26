import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import time,json,sys,os,numpy as np
T0=time.perf_counter()
import torch,torch.nn as nn,torch.nn.functional as F
torch.set_num_threads(1)
sys.path.insert(0,(EXP_ROOT+'/lang9')); import memguard
memguard.check('start'); t_imp=time.perf_counter()-T0
V,CTX,D,L,H,B=8192,256,256,6,4,8
class Blk(nn.Module):
    def __init__(s):
        super().__init__(); s.l1=nn.LayerNorm(D); s.qkv=nn.Linear(D,3*D); s.o=nn.Linear(D,D); s.l2=nn.LayerNorm(D); s.f1=nn.Linear(D,4*D); s.f2=nn.Linear(4*D,D)
    def forward(s,x):
        b,t,_=x.shape; q,k,v=s.qkv(s.l1(x)).view(b,t,3,H,D//H).permute(2,0,3,1,4)
        a=F.scaled_dot_product_attention(q,k,v,is_causal=True).transpose(1,2).reshape(b,t,D); x=x+s.o(a)
        return x+s.f2(F.gelu(s.f1(s.l2(x))))
class LM(nn.Module):
    def __init__(s):
        super().__init__(); s.te=nn.Embedding(V,D); s.pe=nn.Embedding(CTX,D); s.bl=nn.ModuleList([Blk() for _ in range(L)]); s.ln=nn.LayerNorm(D)
    def forward(s,x):
        h=s.te(x)+s.pe(torch.arange(x.shape[1])); 
        for b in s.bl: h=b(h)
        return s.ln(h)@s.te.weight.T
t=time.perf_counter(); torch.manual_seed(13013); m=LM(); opt=torch.optim.AdamW(m.parameters(),lr=3e-4,weight_decay=0.0); t_init=time.perf_counter()-t
npar=sum(p.numel() for p in m.parameters()); rng=np.random.RandomState(13013)
def step():
    x=torch.from_numpy(rng.randint(V,size=(B,CTX+1)).astype(np.int64)); lg=m(x[:,:-1]); lo=F.cross_entropy(lg.reshape(-1,V),x[:,1:].reshape(-1))
    if not torch.isfinite(lo): raise RuntimeError('nonfinite')
    opt.zero_grad(set_to_none=True); lo.backward(); torch.nn.utils.clip_grad_norm_(m.parameters(),1.0); opt.step(); memguard.check('step'); return float(lo)
tw=[]; 
for _ in range(3): t=time.perf_counter(); lw=step(); tw.append(time.perf_counter()-t)
ts=[]; T1=time.perf_counter()
while time.perf_counter()-T1<30.0 and len(ts)<40:
    t=time.perf_counter(); last=step(); ts.append(time.perf_counter()-t)
tot=time.perf_counter()-T1; a=np.array(ts); _rss,_hwm=memguard.usage()
r=dict(config=dict(vocab=V,ctx=CTX,d=D,layers=L,heads=H,batch=B,tokens_per_step=B*CTX,params=npar,tied=True,dtype='fp32',threads=1),t_import_s=t_imp,t_build_opt_init_s=t_init,warmup_step_s=tw,
 n_timed=len(ts),step_median_s=float(np.median(a)),step_min_s=float(a.min()),step_max_s=float(a.max()),step_p90_s=float(np.percentile(a,90)),timed_total_s=tot,
 tok_per_s_median=float(B*CTX/np.median(a)),tok_per_s_total=float(len(ts)*B*CTX/tot),loss_first_warm=lw,loss_last=last,rss_mib=_hwm,rss_now_mib=_rss,wall=time.perf_counter()-T0,torch=torch.__version__)
os.makedirs((EXP_ROOT+'/calib_lm/out'),exist_ok=True); json.dump(r,open((EXP_ROOT+'/calib_lm/out/res.json'),'w'),indent=1); print(json.dumps(r))
