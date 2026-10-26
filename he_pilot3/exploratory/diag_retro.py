# EXPLORATORY ONLY. Not a pilot 3 contrast, not confirmatory, no CIs. Caveats: (1) round grouping of train.json in blocks of 24 is INFERRED;
# (2) the data come from the collect regime (24 candidates evaluated per round, random improving edit accepted), not test trajectories;
# (3) linear R2 values are in-sample, no validation; (4) the Spearman average includes remove edits; (5) 'top1 by freqrule' takes argmax
# (ties go to the first index) and the 'random' reference is the mean over all 24 candidates; (6) nothing here identifies a cause of freqrule's advantage.
# Exploratory retrospective diagnostic over published he_pilot2 artifacts (commit 5a9cc9d). NOT confirmatory.
# Usage: python3 diag_retro.py /path/to/he_pilot2
# NOTE: grouping train.json into rounds of 24 is INFERRED (collect.log cumulative counts are multiples of 24 and main.go collect
# mode evaluates all 24 pool candidates per round); train.json has no seed/round columns.
import json,sys,collections,numpy as np
from scipy.stats import spearmanr
from numpy.linalg import lstsq
p=sys.argv[1]
d=json.load(open(p+'/train.json'));X=np.array([s['X'] for s in d]);Y=np.array([s['Y'] for s in d])
R=Y.reshape(-1,24);XR=X.reshape(-1,24,11);t=XR[:,:,:4].argmax(2)
print('n',len(Y),'rounds(inferred)',len(R),'frac neg %.2f median %.1f'%((Y<0).mean(),np.median(Y)))
for k in range(4):
    m=t==k;print('type',k,'n',m.sum(),'meanY %.2f sd %.2f frac>0 %.2f'%(R[m].mean(),R[m].std(),(R[m]>0).mean()))
Rc=R-R.mean(1,keepdims=True)
print('var total %.2f within-round %.2f between-round-means %.2f'%(Y.var(),Rc.var(),R.mean(1).var()))
def r2(F,y):
    F=np.c_[F,np.ones(len(y))];w=lstsq(F,y,rcond=None)[0];return 1-((y-F@w)**2).sum()/((y-y.mean())**2).sum()
Xc=(XR-XR.mean(1,keepdims=True)).reshape(-1,11);yc=Rc.ravel()
print('in-sample linear R2 within-round: all %.3f'%r2(Xc,yc),{n:round(r2(Xc[:,c],yc),3) for n,c in [('type',[0,1,2,3]),('freq',[4]),('red',[5])]})
fr=XR[:,:,4]
cs=[spearmanr(fr[i],R[i])[0] for i in range(len(R)) if fr[i].std()>0 and R[i].std()>0]
print('mean within-round spearman freq~gain %.3f (%d rounds)'%(np.mean(cs),len(cs)))
gf=[];gr=[];go=[]
for i in range(len(R)):
    sc=fr[i].copy();sc[(t[i]==2)|(XR[i,:,5]==1)]=-1
    gf.append(R[i][np.argmax(sc)]);gr.append(R[i].mean());go.append(R[i].max())
print('top1 by freqrule %.2f, random %.2f, best-of-24 %.2f'%(np.mean(gf),np.mean(gr),np.mean(go)))
ls=X[:,7]*40;print('train lib size min/median/p90/max',np.percentile(ls,[0,50,90,100]))
f=json.load(open(p+'/final.json'));A=collections.defaultdict(list)
for r in f:A[r['Arm']].append(r)
for a,rs in sorted(A.items()):
    g=lambda k:np.mean([r[k] for r in rs]);print(a,len(rs),'audit %.1f rounds %.1f accepts %.1f lib %.1f max lib %d'%(g('AuditSolve'),g('Rounds'),g('Accepts'),g('LibSize'),max(r['LibSize'] for r in rs)))
