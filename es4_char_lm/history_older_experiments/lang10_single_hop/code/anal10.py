import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')

import numpy as np,json,sys
sys.path.insert(0,(EXP_ROOT+'/lang8')); import lang7 as L
PM=np.array(L.PERMS); TRB=L.TRB; NEW=L.NEWB; C0=4
ES=np.load((EXP_ROOT+'/lang10/eval10.npz')); out={}
trpairs={(w,PM[b][w]) for b in TRB for w in range(4)}
out['heldout_pairs_unseen_in_train']=sum((w,PM[b][w]) not in trpairs for b in NEW for w in range(4))
def dist(b1,b2): return int((PM[b1]!=PM[b2]).sum())
out['heldout_min_hamming_to_train']={int(b):min(dist(b,t) for t in TRB) for b in NEW}
out['n_train_bindings_at_min_dist']={int(b):sum(dist(b,t)==min(dist(b,u) for u in TRB) for t in TRB) for b in NEW}
for s in (21,22):
    P=np.load((EXP_ROOT+'/lang10/runs/c10_ctx_s%d/pred_ctx_s%d.npz')%(s,s)); R={}
    for k in ('TR','E_b'):
        x=ES[k+'_x']; g=ES[k+'_gold']; b=ES[k+'_grp']; p=P[k+'_ctx'][:,:2]-C0; sw=x[:,1:9:2]; sc=x[:,2:9:2]-C0; qs=x[:,20:22]
        ar=np.arange(len(x)); ok=(p==g); R[k+'_first_ok']=float(ok[:,0].mean()); R[k+'_second_ok']=float(ok[:,1].mean()); R[k+'_both']=float(ok.all(1).mean())

        R[k+'_per_binding_exact']={int(u):round(float(ok.all(1)[b==u].mean()),2) for u in np.unique(b)}

        nb_ans=np.zeros_like(g); nd=np.zeros(len(x),dtype=int); ntie=np.zeros(len(x),dtype=int)
        for i in range(len(x)):
            truth=PM[b[i]]; ds=[dist(b[i],t) for t in TRB]; m=min(ds); c=[t for t,d in zip(TRB,ds) if d==m]; nd[i]=m; ntie[i]=len(c)

            ans=[(PM[t][qs[i,0]],PM[t][qs[i,1]]) for t in c]; nb_ans[i]=max(set(ans),key=ans.count)
        diff=(nb_ans!=g).any(1)
        R[k+'_dist_to_nearest_train_hist']={int(d):int((nd==d).sum()) for d in np.unique(nd)}
        R[k+'_episodes_where_nearest_binding_answer_differs_from_gold']=int(diff.sum())
        if diff.sum():
            R[k+'_on_those:_exact']=float(ok.all(1)[diff].mean()); R[k+'_on_those:_pred_equals_nearest_binding_answer']=float((p==nb_ans).all(1)[diff].mean())
            R[k+'_on_those:_pred_other_wrong']=float((~ok.all(1)&~(p==nb_ans).all(1))[diff].mean())

        for j in (0,1):
            wq=qs[:,j]; supc=sc[ar,(sw==wq[:,None]).argmax(1)]; assert (supc==g[:,j]).all()
            wrong=p[:,j]!=g[:,j]; R[k+'_tok%d_wrong'%j]=int(wrong.sum())
            if wrong.sum():

                nbc=nb_ans[:,j]; R[k+'_tok%d_wrong_equals_nearest_binding_colour'%j]=float((p[wrong,j]==nbc[wrong]).mean())
                other=np.array([p[i,j] in [sc[i,m] for m in range(4) if sw[i,m]!=wq[i]] for i in np.where(wrong)[0]]); R[k+'_tok%d_wrong_is_colour_of_another_support_word'%j]=float(other.mean())


        agree=np.array([PM[t][w]==c for i in range(len(x)) for (w,c,t) in [(qs[i,0],g[i,0],None)]]) if False else None

        for j in (0,1):
            same=nb_ans[:,j]==g[:,j]; tokok=ok[:,j]
            R[k+'_tok%d_acc_when_pair_agrees_with_nearest_train_binding'%j]=float(tokok[same].mean()) if same.any() else None
            R[k+'_tok%d_acc_when_pair_differs_from_nearest_train_binding'%j]=float(tokok[~same].mean()) if (~same).any() else None
            R[k+'_tok%d_n_agree_n_differ'%j]=[int(same.sum()),int((~same).sum())]

    for k in ('TR','E_b'):
        p0=P[k+'_ctx'][:,:2]-C0; p2=P[k+'_cf'][:,:2]-C0; g0=ES[k+'_gold']; g2=ES[k+'_goldcf']-C0
        g2=ES[k+'_goldcf']
        R['CF_'+k]=dict(orig_ok=float((p0==g0).all(1).mean()),cf_ok=float((p2==g2).all(1).mean()),both_ok=float(((p0==g0).all(1)&(p2==g2).all(1)).mean()),
            cf_ok_given_orig_ok=float((p2==g2).all(1)[(p0==g0).all(1)].mean()),cf_ok_given_orig_wrong=float((p2==g2).all(1)[~(p0==g0).all(1)].mean()) if (~(p0==g0).all(1)).any() else None,
            cf_pred_equals_orig_pred=float((p2==p0).all(1).mean()),first_tok_cf_ok=float((p2[:,0]==g2[:,0]).mean()),second_tok_cf_ok=float((p2[:,1]==g2[:,1]).mean()))
    out[s]=R
json.dump(out,open((EXP_ROOT+'/lang10/runs_anal_c10.json'),'w'),indent=1); print(json.dumps(out,indent=0))
