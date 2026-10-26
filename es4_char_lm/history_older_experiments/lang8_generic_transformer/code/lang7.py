
import numpy as np, itertools
from collections import Counter
PAIRS=[(x,y) for x in range(4) for y in range(4) if x!=y]
PERMS=list(itertools.permutations(range(4)))
perm0=np.random.RandomState(0).permutation(24); NEWB=[int(i) for i in perm0[:6]]; TRB=[int(i) for i in perm0[6:]]
perm5=np.random.RandomState(5).permutation(24); TRL=[int(i) for i in perm5[:18]]; HL=[int(i) for i in perm5[18:]]
DE,A,PAD=4,5,6; CPAD=4
ORDER0=[frozenset(x) for x in [(0,1),(2,3),(0,2)]]
TRS=[(s,d,0 if frozenset((s,d)) in ORDER0 else 1) for (s,d) in PAIRS]
RSS=[(s,d,1-o) for (s,d,o) in TRS]
def answer(b,p,l):
    L=PERMS[l]; return (L.index(PERMS[b][p[0]]),L.index(PERMS[b][p[1]]))
def sid(s,d,o): return PAIRS.index((s,d))*2+o
def mkT(s,d,o,tmpl):
    core=[DE,s,A,d] if o==0 else [A,d,DE,s]
    return np.array([PAD]+core if tmpl=='train' else core+[PAD],dtype=np.int64)
def gen(rng,bset,strings,layouts,tmpl,prng=None):
    b=int(bset[rng.randint(len(bset))]); s,d,o=strings[rng.randint(len(strings))]; l=int(layouts[rng.randint(len(layouts))])
    sw=rng.permutation(4).astype(np.int64); sc=np.array([PERMS[b][w] for w in sw],dtype=np.int64); sc2=None
    if prng is not None:
        b2=int(prng.choice([x for x in range(24) if x!=b])); sc2=np.array([PERMS[b2][w] for w in sw],dtype=np.int64)
    return dict(b=b,s=s,d=d,o=o,l=l,sw=sw,sc=sc,sc2=sc2,lay=np.array(PERMS[l],dtype=np.int64),T=mkT(s,d,o,tmpl),tmpl=tmpl,gold=answer(b,(s,d),l),sid=sid(s,d,o))
def swapcues(e):
    e=dict(e); T=e['T'].copy(); m4=T==DE; m5=T==A; T[m4]=A; T[m5]=DE; e['T']=T; e['gold']=answer(e['b'],(e['d'],e['s']),e['l']); e['swapped']=True; return e
def parse(T):
    src=dst=None
    for j,t in enumerate(T):
        if t==DE: src=int(T[j+1])
        if t==A: dst=int(T[j+1])
    return src,dst
def validate(ep,bset,strings,layouts):
    v=[]
    if ep['b'] not in bset: v.append('binding_split')
    if strings is not None and (ep['s'],ep['d'],ep['o']) not in strings: v.append('string_split')
    if ep['l'] not in layouts: v.append('layout_split')
    if sorted(ep['sw'])!=[0,1,2,3] or sorted(ep['sc'])!=[0,1,2,3]: v.append('support')
    if len(ep['T'])!=5 or ep['T'].max()>6 or ep['T'].min()<0: v.append('tokens')
    src,dst=parse(ep['T']); col={int(w):int(c) for w,c in zip(ep['sw'],ep['sc'])}; L=list(ep['lay'])
    if src is None or dst is None or (L.index(col[src]),L.index(col[dst]))!=ep['gold']: v.append('oracle_wrong')
    return v
def pack(eps,arm):
    T=np.stack([e['T'] for e in eps]); sw=np.stack([e['sw'] for e in eps]); lay=np.stack([e['lay'] for e in eps])
    sc=np.stack([e['sc2'] if arm=='perm' else e['sc'] for e in eps])
    if arm=='nocue': T=np.where((T==DE)|(T==A),PAD,T)
    if arm=='empty': sw=np.full_like(sw,PAD); sc=np.full_like(sc,CPAD)
    gold=np.array([e['gold'] for e in eps]); return T,sw,sc,lay,gold
def init(seed,d=16):
    r=np.random.RandomState(seed); f=lambda *s,sc=1.0:(r.randn(*s)*sc).astype(np.float32); w=1/np.sqrt(d)
    return dict(E=f(7,d),Ec=f(5,d),P=f(5,d,sc=0.3),Wm=f(d,d,sc=w),Wa=f(d,d,sc=w),rho=f(2,d),Wq=f(d,d,sc=w),Wk=f(d,d,sc=w),Wv=f(d,d,sc=w),Wp=f(d,d,sc=w))
def sm(x):
    x=x-x.max(1,keepdims=True); e=np.exp(x); return e/e.sum(1,keepdims=True)
def fwd(P,T,sw,sc,lay,gold):
    B=len(T); d=P['E'].shape[1]; scl=1.0/np.sqrt(d); N=2*B; role=np.concatenate([np.zeros(B,int),np.ones(B,int)])
    TN=np.concatenate([T,T]); swN=np.concatenate([sw,sw]); scN=np.concatenate([sc,sc]); layN=np.concatenate([lay,lay]); tg=np.concatenate([gold[:,0],gold[:,1]])
    tp=np.concatenate([np.full((N,1),PAD),TN[:,:-1]],1)
    Et=P['E'][TN]; Ep=P['E'][tp]; h=Et+P['P'][None]+Ep@P['Wm']; g=h@P['Wa']; rho=P['rho'][role]
    sr=np.einsum('nld,nd->nl',g,rho)*scl; al=sm(sr); u=np.einsum('nl,nld->nd',al,Et)
    q=u@P['Wq']; ke=P['E'][swN]; K=ke@P['Wk']; ve=P['Ec'][scN]; V=ve@P['Wv']
    s=np.einsum('nd,nid->ni',q,K)*scl; a=sm(s); r=np.einsum('ni,nid->nd',a,V); ce=P['Ec'][layN]; Pc=ce@P['Wp']
    p=sm(np.einsum('nd,nkd->nk',r,Pc)*scl); n=np.arange(N); loss=-np.log(p[n,tg]+1e-30).sum()/B
    return loss,(B,N,role,TN,swN,scN,layN,tg,tp,Et,Ep,h,g,rho,al,u,q,ke,K,ve,V,a,r,ce,Pc,p,scl)
def bwd(P,c):
    B,N,role,TN,swN,scN,layN,tg,tp,Et,Ep,h,g,rho,al,u,q,ke,K,ve,V,a,r,ce,Pc,p,scl=c
    G={k:np.zeros_like(v) for k,v in P.items()}; n=np.arange(N)
    dl=p.copy(); dl[n,tg]-=1; dl/=B
    dr=scl*np.einsum('nk,nkd->nd',dl,Pc); dPc=scl*np.einsum('nk,nd->nkd',dl,r)
    G['Wp']=np.einsum('nkd,nke->de',ce,dPc); np.add.at(G['Ec'],layN,dPc@P['Wp'].T)
    da=np.einsum('nd,nid->ni',dr,V); dV=a[:,:,None]*dr[:,None,:]; ds=a*(da-(a*da).sum(1,keepdims=True))
    dq=scl*np.einsum('ni,nid->nd',ds,K); dK=scl*ds[:,:,None]*q[:,None,:]
    G['Wv']=np.einsum('nid,nie->de',ve,dV); np.add.at(G['Ec'],scN,dV@P['Wv'].T)
    G['Wk']=np.einsum('nid,nie->de',ke,dK); np.add.at(G['E'],swN,dK@P['Wk'].T)
    G['Wq']=u.T@dq; du=dq@P['Wq'].T
    dal=np.einsum('nd,nld->nl',du,Et); dEt=al[:,:,None]*du[:,None,:]
    dsr=al*(dal-(al*dal).sum(1,keepdims=True))
    dg=scl*dsr[:,:,None]*rho[:,None,:]; drho=scl*np.einsum('nl,nld->nd',dsr,g); np.add.at(G['rho'],role,drho)
    G['Wa']=np.einsum('nld,nle->de',h,dg); dh=dg@P['Wa'].T
    np.add.at(G['E'],TN,dEt+dh); G['P']=dh.sum(0); np.add.at(G['E'],tp,dh@P['Wm'].T); G['Wm']=np.einsum('nld,nle->de',Ep,dh)
    return G
def predict(P,T,sw,sc,lay):
    _,c=fwd(P,T,sw,sc,lay,np.zeros((len(T),2),dtype=np.int64)); B=len(T); p=c[25]; return np.stack([p[:B].argmax(1),p[B:].argmax(1)],1)
def gradcheck():
    P={k:v.astype(np.float64) for k,v in init(5,4).items()}; rng=np.random.RandomState(6)
    eps=[gen(rng,TRB,TRS,TRL,'train') for _ in range(3)]; T,sw,sc,lay,gold=pack(eps,'full')
    loss,c=fwd(P,T,sw,sc,lay,gold); G=bwd(P,c); worst=0.0; n=0; rr=np.random.RandomState(9)
    for k in P:
        flat=P[k].reshape(-1)
        for i in rr.choice(flat.size,min(8,flat.size),replace=False):
            o=flat[i]; e=1e-5; flat[i]=o+e; lp=fwd(P,T,sw,sc,lay,gold)[0]; flat[i]=o-e; lm=fwd(P,T,sw,sc,lay,gold)[0]; flat[i]=o
            num=(lp-lm)/(2*e); an=G[k].reshape(-1)[i]; worst=max(worst,abs(num-an)/(abs(num)+abs(an)+1e-8)); n+=1
    return worst,n

SPECS={'TR':(TRB,TRS,TRL,3,'train',False),'E1':(NEWB,TRS,HL,6,'train',False),'E2':(TRB,RSS,HL,2,'train',False),'E3':(NEWB,RSS,HL,10,'train',False),
       'E4':(NEWB,TRS,HL,6,'shift',False),'E5':(NEWB,RSS,HL,10,'shift',False),'SW':(NEWB,TRS,HL,6,'train',True)}
def build_eval(seed):
    rng=np.random.RandomState(seed); prng=np.random.RandomState(seed+1); out={}
    for name,(bs,ss,ls,K,tm,sw_) in SPECS.items():
        eps=[]
        for b in bs:
            for st in ss:
                for _ in range(K):
                    e=gen(rng,[b],[st],ls,tm,prng)
                    if validate(e,[b],[st],ls): raise RuntimeError('eval validation '+name)
                    if sw_:
                        e=swapcues(e)
                        if validate(e,[b],None,ls): raise RuntimeError('eval validation swap')
                    e['grp']=b*1000+e['sid']; eps.append(e)
        out[name]=eps
    return out
def noctx_bound(eps_or_split):

    bs,ss,ls,sw_=eps_or_split; tot=0;m=0
    for (s,d,o) in ss:
        gp=(d,s) if sw_ else (s,d)
        for l in ls:
            tot+=Counter(answer(b,gp,l) for b in bs).most_common(1)[0][1]/len(bs); m+=1
    return tot/m
def baselines(eps):

    out={}
    def acc(rule):
        ok=0
        for e in eps:
            T=list(e['T']); words=[t for t in T if t<4]; first,second=words[0],words[1]
            src,dst=rule(e,first,second); ok+= (answer(e['b'],(src,dst),e['l'])==e['gold'])
        return ok/len(eps)
    out['first_is_src']=acc(lambda e,f,s:(f,s)); out['second_is_src']=acc(lambda e,f,s:(s,f))

    out['train_pair_template']=acc(lambda e,f,s:(f,s) if frozenset((f,s)) in ORDER0 else (s,f))
    return out
