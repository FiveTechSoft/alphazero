import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')

import sys,itertools,json,hashlib,numpy as np
sys.path.insert(0,(EXP_ROOT+'/lang10')); sys.path.insert(0,(EXP_ROOT+'/lang8'))
import lang7 as L, gt as G, data10 as D
PM=np.array(L.PERMS); TR=set(L.TRB)
rows=[]
SW=list(itertools.permutations(range(4)))
for q0 in itertools.combinations(range(4),2):
    for b1 in range(24):
        for b2 in range(b1+1,24):
            if all(PM[b1][w]==PM[b2][w] for w in q0) and any(PM[b1][w]!=PM[b2][w] for w in range(4)):
                for q in (q0,q0[::-1]):
                    for si,sw in enumerate(SW): rows.append((b1,b2,q[0],q[1],si))
R=np.array(rows); n=len(R); sw=np.array([SW[i] for i in R[:,4]])
def mk(bcol):
    A=dict(sw=sw,b=R[:,bcol],s=R[:,2],d=R[:,3]); sc=D.true_sc(A); return D.prompt(A,sc),D.gold(A),sc
x1,g1,sc1=mk(0); x2,g2,sc2=mk(1)
assert (g1==g2).all() and (D.oracle(x1,{})==g1).all() and (D.oracle(x2,{})==g2).all()
others=np.array([[w not in (s,d) for w in row] for row,s,d in zip(sw,R[:,2],R[:,3])])
assert ((x1[:,2:9:2]-4!=x2[:,2:9:2]-4)==others).all()
assert (x1[:,1:9:2]==x2[:,1:9:2]).all() and (np.delete(x1,[2,4,6,8],1)==np.delete(x2,[2,4,6,8],1)).all()
assert (x1[:,20:22]==x2[:,20:22]).all()
kind=np.array([''.join(sorted(('S' if a in TR else 'H')+('S' if b in TR else 'H'))) for a,b in R[:,:2]])
strat={k:int(len({(a,b) for (a,b),kk in zip(R[:,:2].tolist(),kind) if kk==k})) for k in ('SS','HS','HH')}
unit=np.array([a*24+b for a,b in R[:,:2]]); heldfirst=np.array([a not in TR for a in R[:,0]])
np.savez((EXP_ROOT+'/lang10p/probe10.npz'),x1=x1,x2=x2,gold=g1,R=R,kind=kind,unit=unit,s_in_TRS=np.array([(s,d) in {(a,b) for a,b,o in L.TRS} for s,d in R[:,2:4]]))
print('n prompts per context',n,'strata (binding pairs)',strat,'ordered queries per pair 2, support orders 24')
print(hashlib.sha256(open((EXP_ROOT+'/lang10p/probe10.npz'),'rb').read()).hexdigest()[:16])
