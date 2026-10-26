# Training-time helpers (save_generation/load_generation) also use pickle for the training state; the recovery and generation scripts do not call them.
import pickle,math,torch,torch.nn as nn,torch.nn.functional as F
from es4_common import *
torch.set_num_threads(1)
V,CMAP=load_vocab()
D,L,H=256,6,4
LR_MAX,LR_MIN,WARM,TOTAL,SEG_STEPS,EVAL_EVERY=1e-3,1e-4,100,22500,300,1500
def lr_at(s):                             # declared schedule: linear warmup 100 steps, cosine to LR_MIN at TOTAL (differs from ES3 constant lr)
    if s<WARM: return LR_MAX*(s+1)/WARM
    return LR_MIN+0.5*(LR_MAX-LR_MIN)*(1+math.cos(math.pi*(s-WARM)/(TOTAL-WARM)))
class Blk(nn.Module):
    def __init__(s):
        super().__init__(); s.l1=nn.LayerNorm(D); s.qkv=nn.Linear(D,3*D); s.o=nn.Linear(D,D); s.l2=nn.LayerNorm(D); s.f1=nn.Linear(D,4*D); s.f2=nn.Linear(4*D,D)
    def forward(s,x):
        b,t,_=x.shape; q,k,v=s.qkv(s.l1(x)).view(b,t,3,H,D//H).permute(2,0,3,1,4)
        a=F.scaled_dot_product_attention(q,k,v,is_causal=True).transpose(1,2).reshape(b,t,D); x=x+s.o(a)
        return x+s.f2(F.gelu(s.f1(s.l2(x))))
class LM(nn.Module):
    def __init__(s):
        super().__init__(); s.te=nn.Embedding(V,D); s.pe=nn.Embedding(CTX,D); s.bl=nn.ModuleList([Blk() for _ in range(L)]); s.ln=nn.LayerNorm(D)
    def forward(s,x):
        h=s.te(x)+s.pe(torch.arange(x.shape[1]))
        for b in s.bl: h=b(h)
        return s.ln(h)@s.te.weight.T
def new_opt(model): return torch.optim.AdamW(model.parameters(),lr=LR_MAX,betas=(0.9,0.999),weight_decay=0.0)   # same as ES3
def train_steps(model,opt,rng,STR,a,b,tick=None):
    model.train(); losses=[]; gns=[]
    for s in range(a,b):
        lr=lr_at(s)
        for g in opt.param_groups: g['lr']=lr
        w=torch.from_numpy(STR.get(rng,B)); lg=model(w[:,:-1]); lo=F.cross_entropy(lg.reshape(-1,V),w[:,1:].reshape(-1))
        if not torch.isfinite(lo): raise RuntimeError('nonfinite loss at step %d'%s)
        opt.zero_grad(set_to_none=True); lo.backward(); g=float(torch.nn.utils.clip_grad_norm_(model.parameters(),1.0))
        if not np.isfinite(g): raise RuntimeError('nonfinite grad at step %d'%s)
        opt.step(); losses.append(float(lo.detach())); gns.append(g)
        if tick and (s+1-a)%50==0: tick('step%d loss50=%.3f'%(s+1,float(np.mean(losses[-50:]))))
    return losses,gns
def evaluate(model,W):                    # same as ES3 (batch 2, full logits)
    model.eval(); tl=[];ul=[]
    with torch.no_grad():
        for i in range(0,len(W),2):
            w=torch.from_numpy(W[i:i+2]); lg=model(w[:,:-1]); tl.append(F.cross_entropy(lg.reshape(-1,V),w[:,1:].reshape(-1),reduction='none')); ul.append(w[:,1:].reshape(-1))
    nl=torch.cat(tl); y=torch.cat(ul); mm=y!=0; model.train()
    return dict(loss=float(nl.mean()),loss_noUNK=float(nl[mm].mean()),unk_rate=float((~mm).float().mean()),perwin=nl.reshape(len(W),CTX).mean(1).tolist())
def tree_equal(a,b):
    if isinstance(a,torch.Tensor): return isinstance(b,torch.Tensor) and a.shape==b.shape and a.dtype==b.dtype and bool(torch.equal(a,b))
    if isinstance(a,dict): return isinstance(b,dict) and a.keys()==b.keys() and all(tree_equal(a[k],b[k]) for k in a)
    if isinstance(a,(list,tuple)): return type(a)==type(b) and len(a)==len(b) and all(tree_equal(x,y) for x,y in zip(a,b))
    return a==b
def tree_sha(o,h=None):                   # content hash of nested tensors/scalars (for exact-continuity test)
    top=h is None; h=h or hashlib.sha256()
    if isinstance(o,torch.Tensor): h.update(str(o.dtype).encode()); h.update(o.detach().contiguous().numpy().tobytes())
    elif isinstance(o,dict):
        for k in sorted(o,key=str): h.update(str(k).encode()); tree_sha(o[k],h)
    elif isinstance(o,(list,tuple)):
        for x in o: tree_sha(x,h)
    else: h.update(repr(o).encode())
    return h.hexdigest() if top else None
# ---- immutable generations + manifest (commit point = last manifest line, appended LAST) ----
def fsync_path(p):
    fd=os.open(p,os.O_RDONLY)
    try: os.fsync(fd)
    finally: os.close(fd)
def read_manifest(root):
    p=root+'MANIFEST.jsonl'; rows=[]; torn=False; prev=''; last_line=''
    if not os.path.exists(p): return rows,torn,last_line
    txt=open(p).read(); parts=txt.split('\n'); tail=parts[-1]; torn=(tail!='')
    for ln in parts[:-1]:
        row=json.loads(ln); assert row['prev']==prev,('manifest chain broken',row['n']); rows.append(row); prev=sha_str(ln); last_line=ln
    return rows,torn,last_line
def save_generation(root,n,step,model,opt,rng,extra):
    gd=root+'gens/'; os.makedirs(gd,exist_ok=True); g=gd+'g%03d/'%n; assert not os.path.exists(g),('generation exists',g); os.makedirs(g)
    st=dict(step=step,np_rng=rng.get_state(),torch_rng=torch.get_rng_state(),extra=extra)
    msd=model.state_dict(); osd=opt.state_dict()
    for name,saver in (('model.pt',lambda p:torch.save(msd,p)),('adam.pt',lambda p:torch.save(osd,p)),('state.pkl',lambda p:pickle.dump(st,open(p,'wb')))):
        tmp=g+name+'.tmp'; saver(tmp); fsync_path(tmp); os.replace(tmp,g+name)
    row=dict(n=n,step=step,gen='g%03d'%n,sha_model=sha_file(g+'model.pt'),sha_adam=sha_file(g+'adam.pt'),sha_state=sha_file(g+'state.pkl'),code=code_shas(),time=time.strftime('%Y-%m-%dT%H:%M:%S'))
    assert tree_equal(torch.load(g+'model.pt',weights_only=True),msd) and tree_equal(torch.load(g+'adam.pt',weights_only=True),osd),'readback mismatch'
    for f in ('model.pt','adam.pt','state.pkl'): os.chmod(g+f,0o444)
    fsync_path(g); fsync_path(gd); return row
def publish(root,row):                    # commit: append one manifest line LAST
    rows,torn,last=read_manifest(root); assert not torn,'manifest torn'; assert len(rows)==row['n']-1
    row=dict(row,prev=sha_str(last) if last else ''); line=json.dumps(row,sort_keys=True); assert '\n' not in line
    with open(root+'MANIFEST.jsonl','a') as f: f.write(line+'\n'); f.flush(); os.fsync(f.fileno())
    fsync_path(root); return row
def verify_gen(root,row):
    g=root+'gens/'+row['gen']+'/'
    for f,k in (('model.pt','sha_model'),('adam.pt','sha_adam'),('state.pkl','sha_state')): assert sha_file(g+f)==row[k],('sha mismatch',row['gen'],f)
def load_generation(root,row,model,opt):
    verify_gen(root,row); g=root+'gens/'+row['gen']+'/'
    model.load_state_dict(torch.load(g+'model.pt',weights_only=True)); opt.load_state_dict(torch.load(g+'adam.pt',weights_only=True)); st=pickle.load(open(g+'state.pkl','rb'))
    assert st['step']==row['step']
    for p in model.parameters(): assert int(opt.state[p]['step'])==row['step'],'adam step mismatch'
    rng=np.random.RandomState(); rng.set_state(st['np_rng']); torch.set_rng_state(st['torch_rng']); return rng,st
