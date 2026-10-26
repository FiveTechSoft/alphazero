

import numpy as np
import lang7 as L
CELL=[0,1,2,3]; STOP=4; BOS=5
NV=16
NAMES='B N T sw sc lay prev step tg tp z rho Et Ep h g al u q ke K ve V a r ce Pc p scl'.split()
def init(seed,d=16):
    r=np.random.RandomState(seed); f=lambda *s,sc=1.0:(r.randn(*s)*sc).astype(np.float32); w=1/np.sqrt(d)
    return dict(E=f(7,d),Ec=f(5,d),P=f(5,d,sc=0.3),Wm=f(d,d,sc=w),Wa=f(d,d,sc=w),Wq=f(d,d,sc=w),Wk=f(d,d,sc=w),Wv=f(d,d,sc=w),Wp=f(d,d,sc=w),
                Ed=f(NV,d),Pd=f(3,d,sc=0.3),Wr=f(d,d,sc=w),Eo=f(NV,d))
def fwd(P,T,sw,sc,lay,prev,step,tg,B,useprev=True):
    d=P['E'].shape[1]; scl=1.0/np.sqrt(d); N=len(T)
    if not useprev: prev=np.full(N,BOS,dtype=np.int64)
    tp=np.concatenate([np.full((N,1),L.PAD),T[:,:-1]],1)
    z=P['Ed'][prev]+P['Pd'][step]; rho=z@P['Wr']
    Et=P['E'][T]; Ep=P['E'][tp]; h=Et+P['P'][None]+Ep@P['Wm']; g=h@P['Wa']
    sr=np.einsum('nld,nd->nl',g,rho)*scl; al=L.sm(sr); u=np.einsum('nl,nld->nd',al,Et)
    q=u@P['Wq']; ke=P['E'][sw]; K=ke@P['Wk']; ve=P['Ec'][sc]; V=ve@P['Wv']
    s=np.einsum('nd,nid->ni',q,K)*scl; a=L.sm(s); r=np.einsum('ni,nid->nd',a,V); ce=P['Ec'][lay]; Pc=ce@P['Wp']
    lc=np.einsum('nd,nkd->nk',r,Pc)*scl; lo=z@P['Eo'][4:].T
    p=L.sm(np.concatenate([lc,lo],1)); n=np.arange(N); loss=-np.log(p[n,tg]+1e-30).sum()/B
    vals=dict(B=B,N=N,T=T,sw=sw,sc=sc,lay=lay,prev=prev,step=step,tg=tg,tp=tp,z=z,rho=rho,Et=Et,Ep=Ep,h=h,g=g,al=al,u=u,q=q,ke=ke,K=K,ve=ve,V=V,a=a,r=r,ce=ce,Pc=Pc,p=p,scl=scl)
    return loss,vals
def bwd(P,c):
    B,N,T,sw,sc,lay,prev,step,tg,tp,z,rho,Et,Ep,h,g,al,u,q,ke,K,ve,V,a,r,ce,Pc,p,scl=[c[k] for k in NAMES]
    G={k:np.zeros_like(v) for k,v in P.items()}; n=np.arange(N)
    dl=p.copy(); dl[n,tg]-=1; dl/=B; dlc=dl[:,:4]; dlo=dl[:,4:]
    dr=scl*np.einsum('nk,nkd->nd',dlc,Pc); dPc=scl*np.einsum('nk,nd->nkd',dlc,r)
    G['Wp']=np.einsum('nkd,nke->de',ce,dPc); np.add.at(G['Ec'],lay,dPc@P['Wp'].T)
    dz=dlo@P['Eo'][4:]; G['Eo'][4:]=dlo.T@z
    da=np.einsum('nd,nid->ni',dr,V); dV=a[:,:,None]*dr[:,None,:]; ds=a*(da-(a*da).sum(1,keepdims=True))
    dq=scl*np.einsum('ni,nid->nd',ds,K); dK=scl*ds[:,:,None]*q[:,None,:]
    G['Wv']=np.einsum('nid,nie->de',ve,dV); np.add.at(G['Ec'],sc,dV@P['Wv'].T)
    G['Wk']=np.einsum('nid,nie->de',ke,dK); np.add.at(G['E'],sw,dK@P['Wk'].T)
    G['Wq']=u.T@dq; du=dq@P['Wq'].T
    dal=np.einsum('nd,nld->nl',du,Et); dEt=al[:,:,None]*du[:,None,:]
    dsr=al*(dal-(al*dal).sum(1,keepdims=True))
    dg=scl*dsr[:,:,None]*rho[:,None,:]; drho=scl*np.einsum('nl,nld->nd',dsr,g)
    G['Wr']=z.T@drho; dz=dz+drho@P['Wr'].T
    np.add.at(G['Ed'],prev,dz); np.add.at(G['Pd'],step,dz)
    G['Wa']=np.einsum('nld,nle->de',h,dg); dh=dg@P['Wa'].T
    np.add.at(G['E'],T,dEt+dh); G['P']=dh.sum(0); np.add.at(G['E'],tp,dh@P['Wm'].T); G['Wm']=np.einsum('nld,nle->de',Ep,dh)
    return G
def tf_batch(T,sw,sc,lay,gold):
    B=len(T); bos=np.full(B,BOS,dtype=np.int64)
    X=[np.concatenate([a]*3) for a in (T,sw,sc,lay)]
    prev=np.concatenate([bos,gold[:,0],gold[:,1]]).astype(np.int64); step=np.repeat(np.arange(3),B)
    tg=np.concatenate([gold[:,0],gold[:,1],np.full(B,STOP)]).astype(np.int64)
    return X[0],X[1],X[2],X[3],prev,step,tg
def train_loss(P,T,sw,sc,lay,gold,useprev):
    T3,sw3,sc3,lay3,prev,step,tg=tf_batch(T,sw,sc,lay,gold); return fwd(P,T3,sw3,sc3,lay3,prev,step,tg,len(T),useprev)
def tf_pred(P,T,sw,sc,lay,gold,useprev):
    T3,sw3,sc3,lay3,prev,step,tg=tf_batch(T,sw,sc,lay,gold); _,c=fwd(P,T3,sw3,sc3,lay3,prev,step,tg,len(T),useprev)
    return c['p'].argmax(1).reshape(3,len(T)).T
def free_pred(P,T,sw,sc,lay,useprev):

    N=len(T); out=np.zeros((N,3),dtype=np.int64); prev=np.full(N,BOS,dtype=np.int64); z0=np.zeros(N,dtype=np.int64)
    for t in range(3):
        _,c=fwd(P,T,sw,sc,lay,prev,np.full(N,t,dtype=np.int64),z0,1,useprev); out[:,t]=c['p'].argmax(1); prev=out[:,t].copy()
    return out
def gradcheck(useprev):
    P={k:v.astype(np.float64) for k,v in init(5,4).items()}; rng=np.random.RandomState(6)
    eps=[L.gen(rng,L.TRB,L.TRS,L.TRL,'train') for _ in range(3)]; T,sw,sc,lay,gold=L.pack(eps,'full')
    loss,c=train_loss(P,T,sw,sc,lay,gold,useprev); G=bwd(P,c); worst=0.0; n=0; rr=np.random.RandomState(9)
    for k in P:
        flat=P[k].reshape(-1)
        for i in rr.choice(flat.size,min(8,flat.size),replace=False):
            o=flat[i]; e=1e-5; flat[i]=o+e; lp=train_loss(P,T,sw,sc,lay,gold,useprev)[0]; flat[i]=o-e; lm=train_loss(P,T,sw,sc,lay,gold,useprev)[0]; flat[i]=o
            num=(lp-lm)/(2*e); an=G[k].reshape(-1)[i]; worst=max(worst,abs(num-an)/(abs(num)+abs(an)+1e-8)); n+=1
    return worst,n
