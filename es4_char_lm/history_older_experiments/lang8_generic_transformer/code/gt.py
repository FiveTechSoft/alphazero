
import numpy as np, torch, torch.nn as nn, torch.nn.functional as F
import lang7 as L
C0,K0=4,8
STOP,DE,A,PAD,SUP,LAY,Q,OUT=12,13,14,15,16,17,18,19; NV=20
PL=25; ML=27
def ser(T,sw,sc,lay):

    N=len(T); x=np.zeros((N,PL),dtype=np.int64); x[:,0]=SUP; x[:,1:9:2]=sw; x[:,2:9:2]=C0+sc; x[:,9]=LAY
    for i in range(4): x[:,10+2*i]=K0+i; x[:,11+2*i]=C0+lay[:,i]
    x[:,18]=Q; x[:,19:24]=np.where(T==L.DE,DE,np.where(T==L.A,A,np.where(T==L.PAD,PAD,T))); x[:,24]=OUT
    return x
def unser(x):
    T=x[:,19:24].copy(); T=np.where(T==DE,L.DE,np.where(T==A,L.A,np.where(T==PAD,L.PAD,T)))
    return T,x[:,1:9:2].copy(),x[:,2:9:2]-C0,x[:,[11,13,15,17]]-C0
def target(gold): return np.concatenate([K0+gold,np.full((len(gold),1),STOP)],1).astype(np.int64)
class Block(nn.Module):
    def __init__(s,d,h):
        super().__init__(); s.ln1=nn.LayerNorm(d); s.qkv=nn.Linear(d,3*d); s.o=nn.Linear(d,d); s.ln2=nn.LayerNorm(d); s.f1=nn.Linear(d,4*d); s.f2=nn.Linear(4*d,d); s.h=h
    def forward(s,x):
        B,T,D=x.shape; q,k,v=s.qkv(s.ln1(x)).split(D,2)
        q,k,v=[t.view(B,T,s.h,D//s.h).transpose(1,2) for t in (q,k,v)]
        a=F.scaled_dot_product_attention(q,k,v,is_causal=True).transpose(1,2).reshape(B,T,D)
        x=x+s.o(a); return x+s.f2(F.gelu(s.f1(s.ln2(x))))
class GT(nn.Module):
    def __init__(s,d=64,h=4,nl=2):
        super().__init__(); s.emb=nn.Embedding(NV,d); s.pos=nn.Embedding(ML,d); s.blocks=nn.ModuleList([Block(d,h) for _ in range(nl)]); s.lnf=nn.LayerNorm(d)
    def forward(s,x):
        z=s.emb(x)+s.pos(torch.arange(x.shape[1]))
        for b in s.blocks: z=b(z)
        return s.lnf(z)@s.emb.weight.T
def loss_tf(m,prompt,tgt):
    x=torch.cat([prompt,tgt[:,:2]],1); lg=m(x)[:,PL-1:PL+2]; return F.cross_entropy(lg.reshape(-1,NV),tgt.reshape(-1))
@torch.no_grad()
def greedy(m,prompt):
    x=prompt
    for _ in range(3): x=torch.cat([x,m(x)[:,-1].argmax(1,keepdim=True)],1)
    return x[:,PL:PL+3]
