
import numpy as np, torch
import gt as G, memguard
def greedy_batched(m,x,bs=128):
    assert 1<=bs<=128
    outs=[]
    for i in range(0,len(x),bs):
        memguard.check('eval batch %d'%(i//bs)); outs.append(G.greedy(m,torch.from_numpy(np.ascontiguousarray(x[i:i+bs]))).numpy())
    return np.concatenate(outs)
