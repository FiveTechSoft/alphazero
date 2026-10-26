import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')
import numpy as np,time,json,os,hashlib,pickle
from collections import Counter
import diag5 as D
T0=time.perf_counter(); CAP=170
def guard():
    if time.perf_counter()-T0>CAP: raise RuntimeError('internal wall guard')
OUT=(EXP_ROOT+'/lang5/inv/out'); os.makedirs(OUT,exist_ok=True)
PAD=12; BS=32; LR=3e-3
def mk_P(seed):
    rng=np.random.RandomState(seed); P=D.init(rng,16,64,np.float32)
    P['E']=np.vstack([P['E'],(rng.randn(1,16)*0.1).astype(np.float32)]); return P
def make_ep(rng,b,qp,ql=None,sl=None):
    ep=D.gen_episode(rng,b,qp)
    if sl is not None: ep['sup']=[(p,sl,D.answer(b,p,sl)) for p,l,a in ep['sup']]
    if ql is not None: ep['ql']=ql; ep['qans']=D.answer(b,qp,ql)
    return ep
def build_eval(seed,bs,ps,K,qls,sl,pt):
    rng=np.random.RandomState(seed); q=[];st=[];tg=[];grp=[]
    for b in bs:
        for pi,p in enumerate(ps):
            for ql in (qls if qls is not None else [None]):
                for _ in range(K):
                    ep=make_ep(rng,int(b),p,ql,sl)
                    if D.validate(ep,pt): raise RuntimeError('eval validation violation')
                    q.append(D.tok_query(ep)); st.append(D.tok_support(ep)); tg.append(ep['qans']); grp.append(int(b)*10000+pi*100+(ql if ql is not None else 0))
    s=dict(q=np.stack(q),sup=np.stack(st),tgt=np.array(tg),grp=np.array(grp)); s['sup_empty']=np.full_like(s['sup'],PAD); return s
def predict(P,sup,q):
    _,c=D.fwd(P,sup,q,np.zeros((len(q),2),dtype=np.int64)); return np.stack([c[6].argmax(1),c[7].argmax(1)],1)
def evaluate(P,sets,key='sup'):
    out={};preds={}
    for n,s in sets.items():
        pr=predict(P,s[key],s['q']); preds[n]=pr; ok=(pr==s['tgt']).all(1); g=s['grp']
        ga=[float(ok[g==u].mean()) for u in np.unique(g)]
        out[n]=dict(joint=float(ok.mean()),src=float((pr[:,0]==s['tgt'][:,0]).mean()),dst=float((pr[:,1]==s['tgt'][:,1]).mean()),n=int(len(ok)),groups=len(ga),min_group=min(ga),groups_ge80=int(sum(a>=.8 for a in ga)))
    return out,preds
def mk_batch(rng,arm,cfg):
    eps=[]
    for _ in range(BS):
        b=int(cfg['bset'][rng.randint(len(cfg['bset']))]); qp=cfg['pairs'][rng.randint(len(cfg['pairs']))]
        ql=cfg['qls'][rng.randint(len(cfg['qls']))] if cfg['qls'] is not None else None
        ep=make_ep(rng,b,qp,ql,cfg['sl'])
        if D.validate(ep,'seen'): raise RuntimeError('train validation violation')
        eps.append(ep)
    sup=np.stack([D.tok_support(e) for e in eps]); q=np.stack([D.tok_query(e) for e in eps]); tgt=np.array([e['qans'] for e in eps],dtype=np.int64)
    assert q.max()<8 and sup[:,:,:6].max()<8 and sup[:,:,6:].min()>=8 and sup[:,:,6:].max()<12
    if arm=='empty': sup=np.full_like(sup,PAD)
    return sup,q,tgt
def train(tag,arm,steps,rseed,cfg,sets,every):
    P=mk_P(21); rng=np.random.RandomState(rseed); np.savez(OUT+'/init_'+tag+'.npz',**P)
    m={k:np.zeros_like(v) for k,v in P.items()}; v2={k:np.zeros_like(v) for k,v in P.items()}; t=0;acc=0.0;curve=[]
    for step in range(steps):
        sup,q,tgt=mk_batch(rng,arm,cfg); loss,c=D.fwd(P,sup,q,tgt)
        if not np.isfinite(loss): raise RuntimeError('nonfinite loss')
        g=D.bwd(P,sup,q,tgt,c); t+=1
        for k in P:
            m[k]=0.9*m[k]+0.1*g[k]; v2[k]=0.999*v2[k]+0.001*g[k]**2
            P[k]-=(LR*(m[k]/(1-0.9**t))/(np.sqrt(v2[k]/(1-0.999**t))+1e-8)).astype(np.float32)
        acc+=float(loss)
        if (step+1)%every==0: curve.append(acc/every); acc=0.0; guard()
    np.savez(OUT+'/final_'+tag+'.npz',**P); np.savez(OUT+'/adam_'+tag+'.npz',t=t,**{'m_'+k:v for k,v in m.items()},**{'v_'+k:v for k,v in v2.items()})
    pickle.dump(rng.get_state(),open(OUT+'/rng_'+tag+'.pkl','wb'))
    key='sup_empty' if arm=='empty' else 'sup'
    res,preds=evaluate(P,sets,key)
    np.savez(OUT+'/pred_'+tag+'.npz',**preds,**{k+'_gold':sets[k]['tgt'] for k in sets})
    L=dict(np.load(OUT+'/final_'+tag+'.npz')); _,p2=evaluate(L,sets,key)
    r=dict(final=res,loss_curve=curve,parity=bool(all((preds[k]==p2[k]).all() for k in preds)),steps=t)
    if arm=='full': r['on_empty_support']=evaluate(P,sets,'sup_empty')[0]
    return r
def bound(bs,ps,layouts):
    s=0;n=0
    for p in ps:
        for L in layouts:
            c=Counter(D.answer(b,p,L) for b in bs); s+=c.most_common(1)[0][1]/len(bs); n+=1
    return s/n
R={}

HELD_ALL=np.random.RandomState(5).permutation(24); TRL=[int(x) for x in HELD_ALL[:18]]; HL=[int(x) for x in HELD_ALL[18:]]
b0=int(D.TRB[0])
EA={'A1':build_eval(4001,[b0],D.SEEN,10,HL,None,'seen'),'A2':build_eval(4002,[b0],D.RES,5,TRL,None,'res'),'A3':build_eval(4003,[b0],D.RES,20,HL,None,'res')}
EB={'E3':build_eval(4011,D.NEWB,D.RES,50,None,None,'res'),'E1':build_eval(4012,D.NEWB,D.SEEN,25,None,None,'seen'),'E2':build_eval(4013,D.TRB,D.RES,10,None,None,'res'),'TR':build_eval(4014,D.TRB,D.SEEN,5,None,None,'seen')}
EC={'E3c':build_eval(4021,D.NEWB,D.RES,50,[0],0,'res'),'E1c':build_eval(4022,D.NEWB,D.SEEN,25,[0],0,'seen'),'E2c':build_eval(4023,D.TRB,D.RES,10,[0],0,'res'),'TRc':build_eval(4024,D.TRB,D.SEEN,5,[0],0,'seen')}
np.savez(OUT+'/eval_frozen.npz',**{n+'_'+f:v for E in (EA,EB,EC) for n,s in E.items() for f,v in s.items()})
R['eval_sha']=hashlib.sha256(open(OUT+'/eval_frozen.npz','rb').read()).hexdigest()
R['bounds']=dict(B_E3=bound(D.NEWB,D.RES,range(24)),B_E1=bound(D.NEWB,D.SEEN,range(24)),B_E2=bound(D.TRB,D.RES,range(24)),B_TR=bound(D.TRB,D.SEEN,range(24)),
    C_E3=bound(D.NEWB,D.RES,[0]),C_E1=bound(D.NEWB,D.SEEN,[0]),C_E2=bound(D.TRB,D.RES,[0]),C_TR=bound(D.TRB,D.SEEN,[0]))
guard()

R['A']=train('A_full','full',3000,6001,dict(bset=[b0],pairs=D.SEEN,qls=TRL,sl=None),EA,250); guard()
a=R['A']['final']; R['A_pass']=bool(a['A1']['joint']>=.90); R['A_comp']=bool(a['A3']['joint']>=.90)

cb=dict(bset=D.TRB,pairs=D.SEEN,qls=None,sl=None)
R['B_full']=train('B_full','full',10000,6002,cb,EB,500); guard()
R['B_empty']=train('B_empty','empty',10000,6002,cb,EB,500); guard()
f=R['B_full']['final'];e=R['B_empty']['final'];fe=R['B_full']['on_empty_support']
R['B_ctx_use']=bool(f['TR']['joint']>=.25 and f['TR']['joint']-e['TR']['joint']>=.10 and f['TR']['joint']-fe['TR']['joint']>=.10)
R['B_transfer']=bool(f['E3']['joint']>=.90 and f['E3']['min_group']>=.80 and f['E3']['joint']-e['E3']['joint']>=.30)

cc=dict(bset=D.TRB,pairs=D.SEEN,qls=[0],sl=0)
R['C_full']=train('C_full','full',3000,6003,cc,EC,250); guard()
R['C_empty']=train('C_empty','empty',3000,6003,cc,EC,250); guard()
f=R['C_full']['final'];e=R['C_empty']['final']
R['C_pass']=bool(f['E3c']['joint']>=.90 and f['E3c']['min_group']>=.80 and f['E3c']['joint']-e['E3c']['joint']>=.30)
R['wall_sec']=time.perf_counter()-T0
for l in open('/proc/self/status'):
    if l.startswith('VmHWM'): R['rss_peak_mib']=int(l.split()[1])/1024
json.dump(R,open(OUT+'/result.json','w'),indent=1); print('done')
