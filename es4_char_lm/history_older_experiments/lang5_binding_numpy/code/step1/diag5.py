
import numpy as np, time, json, itertools, sys, os, hashlib
T0=time.perf_counter()
def guard():
    if time.perf_counter()-T0>100: raise RuntimeError("internal wall guard 100s")
def rss_hwm():
    for l in open('/proc/self/status'):
        if l.startswith('VmHWM'): return int(l.split()[1])/1024.0
PAIRS=[(x,y) for x in range(4) for y in range(4) if x!=y]
RES=[(0,1),(1,2),(2,3),(3,0)]
SEEN=[p for p in PAIRS if p not in RES]
for w in range(4):
    assert sum(1 for p in SEEN if p[0]==w)==2 and sum(1 for p in SEEN if p[1]==w)==2
PERMS=list(itertools.permutations(range(4)))
perm0=np.random.RandomState(0).permutation(24)
NEWB=[int(i) for i in perm0[:6]]; TRB=[int(i) for i in perm0[6:]]
def cell_of(L,c): return L.index(c)
def answer(b,pair,Lidx):
    L=PERMS[Lidx]; return (cell_of(L,PERMS[b][pair[0]]), cell_of(L,PERMS[b][pair[1]]))
NS=4
def gen_episode(rng,bidx,qpair,ans_b=None):
    cands=[p for p in SEEN if p!=qpair]
    tries=0
    while True:
        tries+=1
        idx=rng.choice(len(cands),NS,replace=False)
        sp=[cands[i] for i in idx]
        words=set(w for p in sp for w in p)
        if qpair[0] in words and qpair[1] in words: break
        if tries>500: raise RuntimeError("coverage failure")
    ab=bidx if ans_b is None else ans_b
    sl=[int(rng.randint(24)) for _ in range(NS)]
    sup=[];
    for p,l in zip(sp,sl):
        a=answer(ab,p,l)
        sup.append((p,l,a))
    ql=int(rng.randint(24))
    return dict(b=bidx,qpair=qpair,ql=ql,qans=answer(bidx,qpair,ql),sup=sup,tries=tries)
def tok_support(ep):
    out=np.zeros((NS,8),dtype=np.int64)
    for i,(p,l,a) in enumerate(ep['sup']):
        L=PERMS[l]; out[i]=[p[0],p[1]]+[4+c for c in L]+[8+a[0],8+a[1]]
    return out
def tok_query(ep):
    L=PERMS[ep['ql']]; return np.array([ep['qpair'][0],ep['qpair'][1]]+[4+c for c in L],dtype=np.int64)
def batch(rng,B,bset,pairs):
    eps=[gen_episode(rng,int(bset[rng.randint(len(bset))]),pairs[rng.randint(len(pairs))]) for _ in range(B)]
    sup=np.stack([tok_support(e) for e in eps]); q=np.stack([tok_query(e) for e in eps])
    tgt=np.array([e['qans'] for e in eps],dtype=np.int64)
    return eps,sup,q,tgt
def validate(ep,split_pairs_ok):
    v=[]
    sp=[p for p,_,_ in ep['sup']]
    if ep['qpair'] in sp: v.append('query_pair_in_support')
    if any(p in RES for p in sp): v.append('reserved_in_support')

    col={}
    for p,l,a in ep['sup']:
        L=PERMS[l]; col[p[0]]=L[a[0]]; col[p[1]]=L[a[1]]
    X,Y=ep['qpair']
    if X not in col or Y not in col: v.append('uncovered'); return v
    QL=PERMS[ep['ql']]
    if (cell_of(QL,col[X]),cell_of(QL,col[Y]))!=ep['qans']: v.append('oracle_wrong')
    if split_pairs_ok=='seen' and ep['qpair'] not in SEEN: v.append('pair_split')
    if split_pairs_ok=='res' and ep['qpair'] not in RES: v.append('pair_split')
    return v

def init(rng,d,H,dt):
    f=lambda *s: (rng.randn(*s)*0.1).astype(dt)
    return dict(E=f(12,d),Wi=f(8*d,H),bi=np.zeros(H,dt),Wq=f(6*d,H),bq=np.zeros(H,dt),W1=f(2*H,4),b1=np.zeros(4,dt),W2=f(2*H,4),b2=np.zeros(4,dt))
def fwd(P,sup,q,tgt,keep=False):
    B,ns,_=sup.shape; d=P['E'].shape[1]; H=P['Wi'].shape[1]
    xs=P['E'][sup].reshape(B,ns,8*d); hs=np.tanh(xs@P['Wi']+P['bi'])
    xq=P['E'][q].reshape(B,6*d); qv=np.tanh(xq@P['Wq']+P['bq'])
    sc=1.0/np.sqrt(H)
    s=np.einsum('bh,bnh->bn',qv,hs)*sc; s=s-s.max(1,keepdims=True); a=np.exp(s); a/=a.sum(1,keepdims=True)
    c=np.einsum('bn,bnh->bh',a,hs); z=np.concatenate([qv,c],1)
    l1=z@P['W1']+P['b1']; l2=z@P['W2']+P['b2']
    def sm(l):
        l=l-l.max(1,keepdims=True); e=np.exp(l); return e/e.sum(1,keepdims=True)
    p1=sm(l1); p2=sm(l2); r=np.arange(B)
    loss=-(np.log(p1[r,tgt[:,0]]+1e-30)+np.log(p2[r,tgt[:,1]]+1e-30)).mean()
    return loss,(xs,hs,xq,qv,a,z,p1,p2,sc)
def bwd(P,sup,q,tgt,cache):
    xs,hs,xq,qv,a,z,p1,p2,sc=cache; B,ns,_=sup.shape; d=P['E'].shape[1]; H=P['Wi'].shape[1]; r=np.arange(B)
    g={k:np.zeros_like(v) for k,v in P.items()}
    d1=p1.copy(); d1[r,tgt[:,0]]-=1; d1/=B; d2=p2.copy(); d2[r,tgt[:,1]]-=1; d2/=B
    g['W1']=z.T@d1; g['b1']=d1.sum(0); g['W2']=z.T@d2; g['b2']=d2.sum(0)
    dz=d1@P['W1'].T+d2@P['W2'].T
    dq=dz[:,:H].copy(); dc=dz[:,H:]
    dh=a[:,:,None]*dc[:,None,:]
    da=np.einsum('bh,bnh->bn',dc,hs)
    ds=a*(da-(a*da).sum(1,keepdims=True))
    dq+=sc*np.einsum('bn,bnh->bh',ds,hs)
    dh+=sc*ds[:,:,None]*qv[:,None,:]
    dpi=dh*(1-hs**2); dpq=dq*(1-qv**2)
    g['Wi']=xs.reshape(B*ns,8*d).T@dpi.reshape(B*ns,H); g['bi']=dpi.sum((0,1))
    g['Wq']=xq.T@dpq; g['bq']=dpq.sum(0)
    dxs=(dpi.reshape(B*ns,H)@P['Wi'].T).reshape(B,ns,8,d); dxq=(dpq@P['Wq'].T).reshape(B,6,d)
    np.add.at(g['E'],sup,dxs); np.add.at(g['E'],q,dxq)
    return g
def gradcheck():
    rng=np.random.RandomState(5); P=init(rng,4,8,np.float64)
    r2=np.random.RandomState(6); eps,sup,q,tgt=batch(r2,3,TRB,SEEN)
    loss,c=fwd(P,sup,q,tgt); g=bwd(P,sup,q,tgt,c); worst=0.0; n=0
    for k in P:
        flat=P[k].reshape(-1)
        for i in rng.choice(flat.size,min(6,flat.size),replace=False):
            o=flat[i]; e=1e-5
            flat[i]=o+e; lp=fwd(P,sup,q,tgt)[0]; flat[i]=o-e; lm=fwd(P,sup,q,tgt)[0]; flat[i]=o
            num=(lp-lm)/(2*e); an=g[k].reshape(-1)[i]
            err=abs(num-an)/(abs(num)+abs(an)+1e-8); worst=max(worst,err); n+=1
    return worst,n

def enum_noctx():
    out={}
    best={}
    cnt_all=0.0; cnt_tr_on_tr=0.0; cnt_tr_on_new=0.0; cnt_bayes_all=0.0; n=0
    for pr_name,pairs in (('seen',SEEN),('res',RES)):
        sa=0;st=0;sn=0;sb=0;m=0
        for p in pairs:
            for L in range(24):
                from collections import Counter
                ctr=Counter(answer(b,p,L) for b in TRB); pred=ctr.most_common(1)[0][0]; best[(p,L)]=pred
                call=Counter(answer(b,p,L) for b in range(24))
                sb+=call.most_common(1)[0][1]/24.0
                st+=sum(1 for b in TRB if answer(b,p,L)==pred)/len(TRB)
                sn+=sum(1 for b in NEWB if answer(b,p,L)==pred)/len(NEWB); m+=1
        out[pr_name]=dict(bayes_uniform24=sb/m,trainprior_on_trainbind=st/m,trainprior_on_newbind=sn/m)

    out['distinct_answers_per_pair_layout']=sorted(set(len(set(answer(b,p,L) for b in range(24))) for p in PAIRS for L in range(24)))

    agree=0;tot=0
    for p in PAIRS:
        for b in range(24):
            for b2 in range(24):
                if b2==b: continue
                tot+=1; agree+= (PERMS[b2][p[0]]==PERMS[b][p[0]] and PERMS[b2][p[1]]==PERMS[b][p[1]])
    out['permuted_labels_agree_on_both_words_exact']=agree/tot
    return out,best
def sampled(rng,name,bset,pairs,N,best,ptype):
    res=dict(n=N,viol=0,viol_kinds={},tries_mean=0.0)
    ors=ord_=orp=0;rnd=0;cp=0;chp=0;so_s=0;so_d=0;do_s=0;do_d=0;npr=0;nprs=0;nprd=0
    pc=0;pagree=0;ppair=0
    for _ in range(N):
        b=int(bset[rng.randint(len(bset))]); qp=pairs[rng.randint(len(pairs))]
        ep=gen_episode(rng,b,qp); res['tries_mean']+=ep['tries']/N
        v=validate(ep,ptype)
        for k in v: res['viol_kinds'][k]=res['viol_kinds'].get(k,0)+1
        res['viol']+= 1 if v else 0
        gs,gd=ep['qans']; QL=PERMS[ep['ql']]; X,Y=qp
        col={}
        for p,l,a in ep['sup']:
            L=PERMS[l]; col[p[0]]=L[a[0]]; col[p[1]]=L[a[1]]
        os_=cell_of(QL,col[X]) if X in col else -1; od_=cell_of(QL,col[Y]) if Y in col else -1
        ors+=(os_==gs); ord_+=(od_==gd); orp+=(os_==gs and od_==gd)

        rest=[c for c in range(4) if c!=gs]; so_d+= (rest[rng.randint(3)]==gd)
        rest=[c for c in range(4) if c!=gd]; do_s+= (rest[rng.randint(3)]==gs)

        r1=rng.randint(4); r2=[c for c in range(4) if c!=r1][rng.randint(3)]; chp+=((r1,r2)==(gs,gd))

        cs=[a[0] for p,l,a in ep['sup'] if p[0]==X]; cd=[a[1] for p,l,a in ep['sup'] if p[1]==Y]
        ps_=cs[0] if cs else rng.randint(4); pd_=cd[0] if cd else rng.randint(4)
        cp+=((ps_,pd_)==(gs,gd))

        pr=best[(qp,ep['ql'])]; npr+=(pr==(gs,gd)); nprs+=(pr[0]==gs); nprd+=(pr[1]==gd)

        b2=int(rng.choice([x for x in range(24) if x!=b]))
        sup2=[(p,l,answer(b2,p,l)) for p,l,a in ep['sup']]
        col2={}
        for p,l,a in sup2:
            L=PERMS[l]; col2[p[0]]=L[a[0]]; col2[p[1]]=L[a[1]]
        p2=(cell_of(QL,col2[X]),cell_of(QL,col2[Y])); ppair+=(p2==(gs,gd))
    res.update(oracle_src=ors/N,oracle_dst=ord_/N,oracle_pair=orp/N,
      srconly_pair=(ors/N)*(so_d/N) if False else None)
    res['srconly_oracle_pair']=so_d/N; res['dstonly_oracle_pair']=do_s/N
    res['random_distinct_cells_pair']=chp/N; res['copy_cells_same_role_pair']=cp/N
    res['noctx_trainprior_pair']=npr/N; res['noctx_trainprior_src']=nprs/N; res['noctx_trainprior_dst']=nprd/N
    res['permuted_labels_oracle_pair_matches_gold']=ppair/N
    del res['srconly_pair']
    return res
def main():
    out={'rss_start_mib':rss_hwm()}
    t=time.perf_counter(); gcw,gcn=gradcheck(); out['gradcheck']=dict(worst_rel_err=gcw,n_entries=gcn,sec=time.perf_counter()-t)
    if gcw>1e-5: out['STOP']='gradcheck failed'; return out
    guard()
    t=time.perf_counter(); en,best=enum_noctx(); out['enumeration']=en; out['enum_sec']=time.perf_counter()-t
    out['space']=dict(bindings=24,layouts=24,pairs=12,seen_pairs=len(SEEN),reserved_pairs=len(RES),train_bindings=len(TRB),new_bindings=len(NEWB),
        triples_binding_pair_layout=24*12*24,support_sets_enumerated=False,support_size=NS)
    rng=np.random.RandomState(1234); N=1000; sp={}
    t=time.perf_counter()
    sp['TRAIN_seen_trainb']=sampled(rng,'T',TRB,SEEN,N,best,'seen'); guard()
    sp['E1_seen_newb']=sampled(rng,'E1',NEWB,SEEN,N,best,'seen'); guard()
    sp['E2_res_trainb']=sampled(rng,'E2',TRB,RES,N,best,'res'); guard()
    sp['E3_res_newb']=sampled(rng,'E3',NEWB,RES,N,best,'res'); guard()
    out['sampled']=sp; out['sampled_sec']=time.perf_counter()-t

    d,H=16,64; P=init(np.random.RandomState(7),d,H,np.float32); rb=np.random.RandomState(99)
    tg=[];tf=[];cpu0=time.process_time(); lossesum=0.0; gnorm=0.0; steps=0; tl=time.perf_counter()
    while steps<300 and time.perf_counter()-tl<25:
        a=time.perf_counter(); eps,sup,q,tgt=batch(rb,32,TRB,SEEN)
        for e in eps:
            if validate(e,'seen'): raise RuntimeError('validation violation in cost loop')
        b=time.perf_counter(); loss,c=fwd(P,sup,q,tgt); g=bwd(P,sup,q,tgt,c); e_=time.perf_counter()
        tg.append(b-a); tf.append(e_-b); steps+=1
        if not np.isfinite(loss): raise RuntimeError('nonfinite loss')
    out['cost']=dict(arch='word-token concat-MLP item encoder H=64 d=16 + dot-attn over NS=4 support + 2 heads; params=%d'%sum(v.size for v in P.values()),
        batch=32,steps=steps,gen_validate_sec_per_batch_median=float(np.median(tg)),fwd_bwd_sec_per_batch_median=float(np.median(tf)),
        fwd_bwd_sec_p90=float(np.percentile(tf,90)),cpu_sec_total=time.process_time()-cpu0,updates_applied=0,loss_last=float(loss))
    out['rss_peak_mib']=rss_hwm(); out['wall_sec']=time.perf_counter()-T0
    return out
if __name__=='__main__':
    r=main(); json.dump(r,open(sys.argv[1],'w'),indent=1); print(json.dumps(r)[:200])
