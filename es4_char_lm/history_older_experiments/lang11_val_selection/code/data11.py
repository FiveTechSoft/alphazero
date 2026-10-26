import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import sys,numpy as np
sys.path.insert(0,(EXP_ROOT+'/lang8'))
import gt as G
C0=6; NW=6; NC=6; PAIRS12=[(p,q) for p in range(4) for q in range(4) if p!=q]
def prompt(W,C,qa,qb):
    N=len(W); x=np.full((N,G.PL),G.PAD,dtype=np.int64); x[:,0]=G.SUP; x[:,1:9:2]=W; x[:,2:9:2]=C0+C; x[:,18]=G.Q; x[:,20]=qa; x[:,21]=qb; x[:,24]=G.OUT; return x
def oracle(x):
    sw=x[:,1:9:2]; sc=x[:,2:9:2]-C0; ar=np.arange(len(x)); return np.stack([sc[ar,(sw==x[:,20:21]).argmax(1)],sc[ar,(sw==x[:,21:22]).argmax(1)]],1)
def target(g): return np.concatenate([C0+g,np.full((len(g),1),G.STOP)],1).astype(np.int64)
def assoc_index(ctx): return (ctx[:,:,0].astype(np.int64)*NC+ctx[:,:,1].astype(np.int64))
def weights(AI,e,eps=0.05):
    w=eps+e[AI].mean(1); w=np.minimum(w,10*w.min()); return w/w.sum()
def ess_ratio(p): return float(1.0/(len(p)*np.square(p).sum()))
