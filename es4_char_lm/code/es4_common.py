# ES4 common (NO torch import: S0 must not pay ~200 MiB of torch RSS)
import sys,time,json,os,hashlib,numpy as np
T0=time.perf_counter(); MARKS=[]
R=os.environ['ES4_ROOT']; DATA=R+'data/'; ES3=None; E4=None; RUNROOT=None
SEED,CTX,B=14014,256,2
def mk(tag,hwm_kill=640.0,wall=90.0):
    h=None
    for l in open('/proc/self/status'):
        if l.startswith('VmHWM'): h=int(l.split()[1])/1024
    t=time.perf_counter()-T0; MARKS.append((tag,round(t,2),round(h,1)))
    print('MARK %s t=%.2f hwm=%.1f'%(tag,t,h),flush=True)
    if h>hwm_kill: raise RuntimeError('HWM %.1f > %.0f at %s'%(h,hwm_kill,tag))
    if t>wall: raise RuntimeError('wall guard %.0fs at %s'%(wall,tag))
    return h
def sha_file(p):
    h=hashlib.sha256()
    with open(p,'rb') as f:
        for c in iter(lambda:f.read(1<<20),b''): h.update(c)
    return h.hexdigest()
def sha_str(s): return hashlib.sha256(s.encode()).hexdigest()
def code_shas():
    d=os.path.dirname(os.path.abspath(__file__)); return {f:sha_file(os.path.join(d,f)) for f in sorted(os.listdir(d)) if f.endswith('.py')}
def load_vocab():
    v=json.load(open(DATA+'vocab.json')); V=v['size']; assert V==len(v['chars'])+1 and v['unk_id']==0
    return V,{c:i+1 for i,c in enumerate(v['chars'])}
def check_data_hashes(names=('train.jsonl','val.jsonl','vocab.json')):
    REP=json.load(open(DATA+'prep2_report.json'))['sha256']; out={}
    for n in names: out[n]=sha_file(DATA+n); assert out[n]==REP[n],('data hash mismatch',n)
    return out
def load_split(split,cmap):
    assert split in ('train','val')      # test is never opened
    return [np.array([cmap.get(c,0) for c in ' '.join(json.loads(ln)['paras'])],dtype=np.uint8) for ln in open(DATA+'%s.jsonl'%split)]
class Sampler:                            # identical to ES3
    def __init__(s,docs):
        s.d=docs; s.n=np.array([len(x)-CTX for x in docs],dtype=np.int64); assert (s.n>0).all(); s.cum=np.cumsum(s.n); s.last=[]
    def get(s,rng,k):
        W=np.zeros((k,CTX+1),dtype=np.int64); s.last=[]
        for j in range(k):
            r=int(rng.randint(s.cum[-1])); i=int(np.searchsorted(s.cum,r,side='right')); st=r-int(s.cum[i]-s.n[i]); W[j]=s.d[i][st:st+CTX+1]; s.last.append(i)
        return W
def eval_windows(STR,SVA):               # the SAME fixed 64+64 windows as ES3
    evtr=STR.get(np.random.RandomState(SEED+1),64); evva=SVA.get(np.random.RandomState(SEED+2),64); return evtr,evva,np.array(SVA.last)
def clus(x,docs):
    ids=np.unique(docs); dm=np.array([x[docs==i].mean() for i in ids])
    return dict(window_mean=float(x.mean()),doc_mean=float(dm.mean()),doc_se=float(dm.std(ddof=1)/np.sqrt(len(dm))),n_docs=int(len(ids)),n_windows=int(len(x)),docs_pos=int((dm>0).sum()),docs_neg=int((dm<0).sum()))
