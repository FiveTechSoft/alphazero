# Independent numpy check of weights_elig.json (run from the package directory). Not part of the TEST. Zero expansions.
import json,numpy as np
d=json.load(open('train.json'));X=np.array([s['X'] for s in d]);Y=np.array([s['Y'] for s in d])
m=(X[:,2]==0)&(X[:,5]==0)
print('total',len(d),'eligible',int(m.sum()),'remove',int((X[:,2]==1).sum()),'redundant non-remove',int(((X[:,5]==1)&(X[:,2]==0)).sum()))
Xe=np.c_[np.ones(m.sum()),X[m]];w=np.linalg.solve(Xe.T@Xe+np.eye(12),Xe.T@Y[m])
g=json.load(open('weights_elig.json'))
print('max abs diff vs weights_elig.json learned_elig: %.3g'%np.abs(np.array(g['learned_elig'])-w).max())
print('w[5] (coefficient on x[4], corpus frequency): %.3f; w[10], w[11] (interactions): %.3f, %.3f'%(w[5],w[10],w[11]))
