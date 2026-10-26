import numpy as np,time,json,os
import lang6 as M
T0=time.perf_counter(); res={}
w,n=M.gradcheck(); res['gradcheck']=dict(worst=w,n=n)
if w>1e-6: res['STOP']='gradcheck'; json.dump(res,open('cost.json','w')); raise SystemExit
ev=M.build_eval(3001)
res['eval_counts']={k:len(v) for k,v in ev.items()}
res['groups']={k:len(set(e['grp'] for e in v)) for k,v in ev.items()}
res['bounds_noctx_exact']={k:M.bound(M.SPLITS[k][0],M.SPLITS[k][1],M.SPLITS[k][2]) for k in M.SPLITS}

ag=0;tot=0
for p in M.PAIRS:
    for b in range(24):
        for b2 in range(24):
            if b2!=b: tot+=1; ag+=(M.PERMS[b2][p[0]]==M.PERMS[b][p[0]] and M.PERMS[b2][p[1]]==M.PERMS[b][p[1]])
res['perm_agree_exact']=ag/tot
res['reserved']=dict(bindings_new=M.NEWB,pairs_reserved=M.RES,layouts_heldout=M.HL,train_bindings=len(M.TRB),train_layouts=len(M.TRL))

P=M.init(7); rng=np.random.RandomState(99); prng=np.random.RandomState(98); tg=[];tf=[];ta=[]
for i in range(300):
    a=time.perf_counter(); eps=[M.gen(rng,M.TRB,M.SEEN,M.TRL,prng) for _ in range(32)]
    for e in eps:
        if M.validate(e,M.TRB,M.SEEN,M.TRL): raise RuntimeError('train validation')
    sw,sc,lay,X,Y,gold=M.pack(eps,'full'); b=time.perf_counter()
    loss,c=M.fwd(P,sw,sc,lay,X,Y,gold); g=M.bwd(P,c); e_=time.perf_counter()
    m={k:np.zeros_like(v) for k,v in P.items()} if i==0 else m
    for k in P: m[k]=0.9*m[k]+0.1*g[k]; _x=(1e-3*m[k]/(np.sqrt(m[k]**2)+1e-8))
    f=time.perf_counter(); tg.append(b-a); tf.append(e_-b); ta.append(f-e_)
res['cost']=dict(gen_validate_median=float(np.median(tg)),fwd_bwd_median=float(np.median(tf)),adam_like_median=float(np.median(ta)),step_total_median=float(np.median(tg)+np.median(tf)+np.median(ta)),
  params=int(sum(v.size for v in P.values())),loss_random_init=float(loss))
res['wall']=time.perf_counter()-T0
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): res['rss_mib']=int(l.split()[1])/1024
json.dump(res,open('cost.json','w'),indent=1); print('done')
