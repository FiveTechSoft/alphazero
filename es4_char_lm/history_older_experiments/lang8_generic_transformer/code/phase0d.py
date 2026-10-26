import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import json,hashlib,time,numpy as np
import lang7 as L, gt as G, dtasks as D
T0=time.perf_counter(); res={}; OUT=(EXP_ROOT+'/lang8/out')
res['code_sha']={f:hashlib.sha256(open(f,'rb').read()).hexdigest() for f in ['lang7.py','gt.py','gi.py','dtasks.py','phase0d.py']}
SPECS={'TR':(L.TRB,L.TRS,L.TRL,3),'E_b':(L.NEWB,L.TRS,L.TRL,6),'E_l':(L.TRB,L.TRS,L.HL,6),'E_bl':(L.NEWB,L.TRS,L.HL,6)}
rng=np.random.RandomState(3201); AR={}
for name,(bs,ss,ls,K) in SPECS.items():
    eps=[]
    for b in bs:
        for st in ss:
            for _ in range(K):
                e=L.gen(rng,[b],[st],ls,'train',None)
                if L.validate(e,[b],[st],ls): raise RuntimeError('eval validation '+name)
                eps.append(e)
    AR[name]=D.arrays(eps)
res['counts']={k:len(v['gold']) for k,v in AR.items()}; res['groups']={k:int(len(set(v['grp']))) for k,v in AR.items()}
flat={}; chk={}; crng=np.random.RandomState(3301)
for task in ('d1','d2'):
    for name,A in AR.items():
        x=D.prompts(task,A); og=D.oracle(x,task); ok=bool((og==A['gold']).all())
        rt=bool(x.shape[1]==G.PL and x.min()>=0 and x.max()<G.NV and (x[:,[10,12,14,16]]==G.K0+np.arange(4)).all() and (x[:,[0,9,18,24]]==np.array([G.SUP,G.LAY,G.Q,G.OUT])).all() and (x[:,[19,22,23]]==G.PAD).all())
        flat[task+'_'+name+'_x']=x; chk[task+'_'+name]=dict(oracle_equals_gold=ok,format=rt)
        if not (ok and rt): res['checks']=chk; res['STOP']='serialization'; json.dump(res,open('base_d.json','w'),indent=1); raise SystemExit
for name,A in AR.items(): flat['gold_'+name]=A['gold']; flat['grp_'+name]=A['grp']

for name in ('TR','E_bl'):
    A=AR[name]
    for task in ('d1','d2'):
        x=flat[task+'_'+name+'_x']
        xl,l2=D.cf_layout(x,A,crng); gl=D.oracle(xl,task); gl2=np.array([L.answer(int(b),(int(s),int(d)),int(l)) for b,s,d,l in zip(A['b'],A['s'],A['d'],l2)])
        okl=bool((gl==gl2).all() and (gl!=A['gold']).any(1).all()); flat['cfL_'+task+'_'+name+'_x']=xl; flat['cfL_'+task+'_'+name+'_gold']=gl
        xs,b2=D.cf_support(x,A,crng,change=(task=='d2')); gs=D.oracle(xs,task)
        if task=='d2':
            gs2=np.array([L.answer(int(b),(int(s),int(d)),int(l)) for b,s,d,l in zip(b2,A['s'],A['d'],A['l'])]); oks=bool((gs==gs2).all() and (gs!=A['gold']).any(1).all())
        else: oks=bool((gs==A['gold']).all())
        flat['cfS_'+task+'_'+name+'_x']=xs; flat['cfS_'+task+'_'+name+'_gold']=gs; chk['cf_'+task+'_'+name]=dict(layout_consistent_and_changed=okl,support_ok=oks)
        if not (okl and oks): res['checks']=chk; res['STOP']='counterfactual'; json.dump(res,open('base_d.json','w'),indent=1); raise SystemExit
res['checks']=chk
np.savez(OUT+'/eval_d.npz',**flat); res['eval_sha']=hashlib.sha256(open(OUT+'/eval_d.npz','rb').read()).hexdigest()

prng=np.random.RandomState(555); gs=[]
for _ in range(20000): gs.append(L.gen(prng,L.TRB,L.TRS,L.TRL,'train',None)['gold'])
gs=np.array(gs); J={}
for a,b in gs: J[(int(a),int(b))]=J.get((int(a),int(b)),0)+1
n=len(gs); P1=np.bincount(gs[:,0],minlength=4)/n; H1=-sum(p*np.log(p) for p in P1 if p>0)
H2=0.0
for a in range(4):
    row=np.array([J.get((a,b),0) for b in range(4)],dtype=float); pa=row.sum()/n
    if row.sum()>0: q=row/row.sum(); H2+=pa*-sum(p*np.log(p) for p in q if p>0)
mode=max(J,key=J.get)
res['prior']=dict(n=n,marginal_c1=[float(p) for p in P1],joint_top3=sorted([(k,v/n) for k,v in J.items()],key=lambda t:-t[1])[:3],H_c1=float(H1),H_c2_given_c1=float(H2),loss_floor_context_free=float((H1+H2)/3),loss_uniform_two_heads=float(2*np.log(4)/3),
    const_mode_pair=list(mode),const_mode_exact_by_split={k:float((v['gold']==np.array(mode)).all(1).mean()) for k,v in AR.items()},
    const_cell_per_step_acc_by_split={k:[float((v['gold'][:,0]==int(np.argmax(np.bincount(gs[:,0],minlength=4)))).mean()),float((v['gold'][:,1]==int(np.argmax(np.bincount(gs[:,1],minlength=4)))).mean())] for k,v in AR.items()})
res['wall']=time.perf_counter()-T0
json.dump(res,open('base_d.json','w'),indent=1); print('done',res['prior']['loss_floor_context_free'])
