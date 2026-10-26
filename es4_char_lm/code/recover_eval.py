# Reload the final model from this package and recompute the fixed validation evaluation (64 windows). No test data, no training.
# Usage: ES4_ROOT=<path of this package, ending with /> python recover_eval.py   (needs torch 2.9.1 CPU, numpy)
import os,sys,json
from es4_common import *
mk('start')
from es4_torch import *
mk('torch')
dh=check_data_hashes(('train.jsonl','val.jsonl','vocab.json'))
TR=load_split('train',CMAP); VA=load_split('val',CMAP); STR,SVA=Sampler(TR),Sampler(VA); evtr,evva,evdocs=eval_windows(STR,SVA); mk('data')
rows=json.load(open(R+'results/GENERATIONS_PUBLIC.json'))['generations']; last=rows[-1]; assert last['gen']=='g075' and last['step']==22500
msha=sha_file(R+'model/model_g075.pt'); assert msha==last['sha_model'],('model sha',msha)
m=LM(); m.load_state_dict(torch.load(R+'model/model_g075.pt',weights_only=True)); mk('model')
ev=evaluate(m,evva); mk('eval')
st=json.load(open(R+'model/train_state_g075.json')); assert st['step']==22500
ref=st['extra']['eval']['val']; fr=json.load(open(R+'results/final_res.json'))
d_loss=abs(ev['loss']-ref['loss']); d_pw=float(np.abs(np.array(ev['perwin'])-np.array(ref['perwin'])).max()); d_fr=abs(ev['loss']-fr['comparison']['val']['model_loss'])
assert d_loss<1e-5 and d_pw<1e-5 and d_fr<1e-5,(d_loss,d_pw,d_fr)
res=dict(recovered_val_loss=ev['loss'],ref_val_loss_train_state=ref['loss'],ref_val_loss_final_res=fr['comparison']['val']['model_loss'],abs_diff_vs_train_state=d_loss,max_abs_diff_perwin=d_pw,abs_diff_vs_final_res=d_fr,model_sha=msha,data_sha=dh,
  windows_val_docs=int(len(np.unique(evdocs))),marks=MARKS,rss_mib=mk('end'),wall=time.perf_counter()-T0,note='check of the reloaded model on the fixed validation windows only; no test, no training')
os.makedirs(R+'recover_out',exist_ok=True); json.dump(res,open(R+'recover_out/recover_res.json','w'),indent=1); print('RECOVER_OK val %.10f ref %.10f maxdiff_perwin %.2e'%(ev['loss'],ref['loss'],d_pw),flush=True)
