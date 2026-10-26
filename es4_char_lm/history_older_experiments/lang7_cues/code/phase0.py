import numpy as np,time,json
import lang7 as M
T0=time.perf_counter(); res={}
w,n=M.gradcheck(); res['gradcheck']=dict(worst=w,n=n)
if w>1e-6: res['STOP']='gradcheck'; json.dump(res,open('cost.json','w')); raise SystemExit
ev=M.build_eval(3101)
res['counts']={k:len(v) for k,v in ev.items()}; res['groups']={k:len(set(e['grp'] for e in v)) for k,v in ev.items()}
from collections import Counter
res['train_orders']=dict(Counter(o for _,_,o in M.TRS)); res['reserved_orders']=dict(Counter(o for _,_,o in M.RSS))
res['noctx_bound']={k:M.noctx_bound((M.SPECS[k][0],M.SPECS[k][1],M.SPECS[k][2],M.SPECS[k][5])) for k in M.SPECS}
res['baselines_perfect_binding']={k:M.baselines(v) for k,v in ev.items()}

P=M.init(7); rng=np.random.RandomState(99); prng=np.random.RandomState(98); tg=[];tf=[];ta=[]
m={k:np.zeros_like(v) for k,v in P.items()}; v2={k:np.zeros_like(v) for k,v in P.items()}; t=0
for i in range(300):
    a=time.perf_counter(); eps=[M.gen(rng,M.TRB,M.TRS,M.TRL,'train',prng) for _ in range(32)]
    for e in eps:
        if M.validate(e,M.TRB,M.TRS,M.TRL): raise RuntimeError('train validation')
    T,sw,sc,lay,gold=M.pack(eps,'full'); b=time.perf_counter()
    loss,c=M.fwd(P,T,sw,sc,lay,gold); g=M.bwd(P,c); e_=time.perf_counter(); t+=1
    for k in P:
        m[k]=0.9*m[k]+0.1*g[k]; v2[k]=0.999*v2[k]+0.001*g[k]**2
        P[k]-=(3e-3*(m[k]/(1-0.9**t))/(np.sqrt(v2[k]/(1-0.999**t))+1e-8)).astype(np.float32)
    f=time.perf_counter(); tg.append(b-a); tf.append(e_-b); ta.append(f-e_)
tot=float(np.median(tg)+np.median(tf)+np.median(ta))
res['cost']=dict(gen_validate=float(np.median(tg)),fwd_bwd=float(np.median(tf)),adam=float(np.median(ta)),step_total=tot,params=int(sum(v.size for v in P.values())))
res['wall']=time.perf_counter()-T0
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): res['rss_mib']=int(l.split()[1])/1024
json.dump(res,open('cost.json','w'),indent=1); print('done',tot)
