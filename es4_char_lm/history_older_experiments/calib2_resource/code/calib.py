import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import time,json,sys,os,numpy as np
T0=time.perf_counter()
def mk(tag):
    r=h=None
    for l in open('/proc/self/status'):
        if l.startswith('VmRSS'): r=int(l.split()[1])/1024
        elif l.startswith('VmHWM'): h=int(l.split()[1])/1024
    print('MARK %s t=%.3f rss=%.1f hwm=%.1f'%(tag,time.perf_counter()-T0,r,h),flush=True); 
    if h>640.0: raise RuntimeError('HWM %.1f > 640 at %s'%(h,tag))
    return r,h
mk('start')
import torch,torch.nn as nn,torch.nn.functional as F
torch.set_num_threads(1); mk('after_import_torch')
V,CTX,D,L,H,B=8192,256,256,6,4,2
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
        h=s.te(x)+s.pe(torch.arange(x.shape[1]))
        for b in s.bl: h=b(h)
        return s.ln(h)@s.te.weight.T
t=time.perf_counter(); torch.manual_seed(13013); m=LM(); opt=torch.optim.AdamW(m.parameters(),lr=3e-4,weight_decay=0.0); t_init=time.perf_counter()-t
npar=sum(p.numel() for p in m.parameters()); print('PARAMS',npar,'init_s',t_init,flush=True); mk('after_build_opt')
rng=np.random.RandomState(13013); ts=[]; hw=[]; phases=[]; T1=time.perf_counter(); n=0
while time.perf_counter()-T1<30.0 and n<60:
    n+=1; ins=n<=3; t=time.perf_counter(); ph={}
    x=torch.from_numpy(rng.randint(V,size=(B,CTX+1)).astype(np.int64)); lg=m(x[:,:-1])
    if ins: ph['fwd']=mk('step%d_forward'%n)[1]; ph['t_fwd']=time.perf_counter()-t
    lo=F.cross_entropy(lg.reshape(-1,V),x[:,1:].reshape(-1))
    if not torch.isfinite(lo): raise RuntimeError('nonfinite')
    if ins: ph['loss']=mk('step%d_loss'%n)[1]
    opt.zero_grad(set_to_none=True); lo.backward()
    if ins: ph['bwd']=mk('step%d_backward'%n)[1]; ph['t_bwd_end']=time.perf_counter()-t
    torch.nn.utils.clip_grad_norm_(m.parameters(),1.0)
    if ins: ph['clip']=mk('step%d_clip'%n)[1]
    opt.step()
    ts.append(time.perf_counter()-t); r,h=mk('step%d_done t_step=%.3f'%(n,ts[-1])); hw.append(h); phases.append(ph)
tot=time.perf_counter()-T1; a=np.array(ts); b=a[3:] if len(a)>3 else a; _r,_h=mk('end')
res=dict(config=dict(vocab=V,ctx=CTX,d=D,layers=L,heads=H,batch=B,tokens_per_step=B*CTX,params=npar,tied=True,dtype='fp32',threads=1),t_build_opt_init_s=t_init,n_steps=n,step_times_s=[float(v) for v in a],hwm_after_each_step=hw,phase_marks_steps_1_3=phases,
 all_steps=dict(median=float(np.median(a)),mean=float(a.mean()),min=float(a.min()),max=float(a.max()),tok_per_s_median=float(B*CTX/np.median(a)),tok_per_s_total=float(n*B*CTX/tot)),
 steps_ge4=dict(median=float(np.median(b)),mean=float(b.mean()),min=float(b.min()),max=float(b.max()),p90=float(np.percentile(b,90)),tok_per_s_median=float(B*CTX/np.median(b))),
 timed_total_s=tot,rss_mib=_h,rss_now_mib=_r,wall=time.perf_counter()-T0,torch=torch.__version__)
os.makedirs((EXP_ROOT+'/calib2/out'),exist_ok=True); json.dump(res,open((EXP_ROOT+'/calib2/out/res.json'),'w'),indent=1); print('DONE',json.dumps(res['all_steps']),json.dumps(res['steps_ge4']),flush=True)
