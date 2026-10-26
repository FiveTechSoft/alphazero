# Generation probe for the frozen ES4 character LM. See PREREG_GEN.txt. No training, no test data.
import os,sys,json
G=os.environ['GEN_ROOT']; os.environ['ES4_ROOT']=os.environ['ES4_ROOT']
from es4_common import *
mk('start')
from es4_torch import *
mk('torch')
MODEL=os.environ['MODEL_PT']; WANT='1dd8d510dea6399504b955ba53341467c151dfc770a119d1a2d74347e0f2744d'
sha0=sha_file(MODEL); assert sha0==WANT,sha0
dh=check_data_hashes(('vocab.json',))
chars=json.load(open(DATA+'vocab.json'))['chars']; inv={i+1:c for i,c in enumerate(chars)}
m=LM(); m.load_state_dict(torch.load(MODEL)); m.eval(); [p.requires_grad_(False) for p in m.parameters()]; mk('model')
unk_log=[]
def enc(s):
    ids=[]
    for c in s:
        i=CMAP.get(c,0)
        if i==0: unk_log.append(c)
        ids.append(i)
    return ids
def dec(ids): return ''.join(inv.get(i,'<UNK>') for i in ids)
def gen(prefix,mode,seed=None,n=120):
    ids=enc(prefix); rng=np.random.RandomState(seed) if mode=='sample' else None; out=[]
    with torch.no_grad():
        for _ in range(n):
            x=torch.tensor([ids[-CTX:]],dtype=torch.long); lg=m(x)[0,-1]
            if mode=='greedy': nxt=int(torch.argmax(lg))
            else:
                p=torch.softmax(lg.double(),-1).numpy(); p=p/p.sum(); nxt=int(rng.choice(len(p),p=p))
            ids.append(nxt); out.append(nxt)
    return dec(out)
P=["Hola, buenos días. ","Hola, ¿cómo estás? ","¿Cuál es la capital de España? ","Pregunta: ¿Qué es un perro? Respuesta: ","Pregunta: ¿Cuánto es dos más dos? Respuesta: "]
rows=[]; g4=None
for k,p in enumerate(P,1):
    g=gen(p,'greedy'); s=gen(p,'sample',26102600+k)
    if k==4: g4=g
    rows.append(dict(id='P%d'%k,prefix=p,greedy=g,sample=s,sample_seed=26102600+k,n_new=120,temperature=1.0)); mk('P%d'%k)
p6=P[3]+g4+" Pregunta: ¿Y un gato? Respuesta: "
rows.append(dict(id='P6',prefix=p6,greedy=gen(p6,'greedy'),sample=gen(p6,'sample',26102606),sample_seed=26102606,n_new=120,temperature=1.0)); mk('P6')
sha1=sha_file(MODEL); assert sha1==sha0
res=dict(model_sha_before=sha0,model_sha_after=sha1,vocab_sha=dh,unk_chars_in_prompts=sorted(set(unk_log)),ctx=CTX,rows=rows,rss_mib=mk('end'),wall=time.perf_counter()-T0)
json.dump(res,open(G+'/out/gen_results.json','w'),ensure_ascii=False,indent=1)
with open(G+'/out/gen_results.txt','w') as f:
    for r in rows:
        f.write('=== %s  prefijo: %r\n[greedy]\n%s\n[sample T=1.0 seed %d]\n%s\n\n'%(r['id'],r['prefix'],r['greedy'],r['sample_seed'],r['sample']))
print('GEN_OK',flush=True)
