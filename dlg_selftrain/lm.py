# Small char-level GPT for the dialogue self-training experiment.
# Architecture adapted from es4_char_lm (tied embeddings, pre-LN blocks); smaller.
import math, os, json, time, hashlib
import numpy as np, torch, torch.nn as nn, torch.nn.functional as F
torch.set_num_threads(1)

D,L,H,CTX = 128,4,4,352
BATCH = 16

class Blk(nn.Module):
    def __init__(s):
        super().__init__(); s.l1=nn.LayerNorm(D); s.qkv=nn.Linear(D,3*D); s.o=nn.Linear(D,D)
        s.l2=nn.LayerNorm(D); s.f1=nn.Linear(D,4*D); s.f2=nn.Linear(4*D,D)
    def forward(s,x):
        b,t,_=x.shape
        q,k,v=s.qkv(s.l1(x)).view(b,t,3,H,D//H).permute(2,0,3,1,4)
        a=F.scaled_dot_product_attention(q,k,v,is_causal=True).transpose(1,2).reshape(b,t,D)
        x=x+s.o(a)
        return x+s.f2(F.gelu(s.f1(s.l2(x))))

class LM(nn.Module):
    def __init__(s,V):
        super().__init__(); s.V=V
        s.te=nn.Embedding(V,D); s.pe=nn.Embedding(CTX,D); s.bl=nn.ModuleList([Blk() for _ in range(L)]); s.ln=nn.LayerNorm(D)
    def forward(s,x):
        h=s.te(x)+s.pe(torch.arange(x.shape[1]))
        for b in s.bl: h=b(h)
        return s.ln(h)@s.te.weight.T

def n_params(m): return sum(p.numel() for p in m.parameters())

class Vocab:
    def __init__(s, texts):
        chars = sorted(set("".join(texts)) | set("\n"))
        s.chars = chars; s.cmap = {c:i for i,c in enumerate(chars)}; s.size = len(chars)
    def enc(s, t): return np.array([s.cmap[c] for c in t], dtype=np.int64)
    def dec(s, ids): return "".join(s.chars[i] for i in ids)

def make_stream(vocab, texts, rng):
    """Concatenate dialogues (shuffled) with newline separators into one id array."""
    ts = list(texts); rng.shuffle(ts)
    return vocab.enc("\n".join(ts) + "\n")

def sample_windows(rng, arr, k):
    W = np.zeros((k, CTX+1), dtype=np.int64)
    for j in range(k):
        st = int(rng.randint(0, len(arr)-CTX-1))
        W[j] = arr[st:st+CTX+1]
    return W

def train_steps(model, opt, arr, rng, n, lr=1e-3, warm=50, log_every=100, tag=""):
    model.train(); out=[]
    for s in range(n):
        for g in opt.param_groups: g['lr'] = lr*min(1.0,(s+1)/warm)
        w = torch.from_numpy(sample_windows(rng, arr, BATCH))
        lg = model(w[:,:-1]); lo = F.cross_entropy(lg.reshape(-1,model.V), w[:,1:].reshape(-1))
        if not torch.isfinite(lo): raise RuntimeError("nonfinite loss")
        opt.zero_grad(set_to_none=True); lo.backward()
        gn = float(torch.nn.utils.clip_grad_norm_(model.parameters(),1.0))
        opt.step(); out.append(float(lo.detach()))
        if (s+1)%log_every==0: print("  %s step %d/%d loss %.3f" % (tag, s+1, n, float(np.mean(out[-50:]))), flush=True)
    return out

@torch.no_grad()
def answer(model, vocab, prefix, max_new=14, temp=0.0, rng=None):
    """Generate until '.' ; returns decoded answer string (without the dot)."""
    model.eval()
    ids = list(vocab.enc(prefix))[-(CTX-max_new-1):]
    x = torch.tensor([ids], dtype=torch.long)
    gen = []
    for _ in range(max_new):
        lg = model(x)[0,-1]
        if temp and temp > 0:
            probs = F.softmax(lg/temp, dim=-1).numpy()
            probs = probs/probs.sum()
            t = int(rng.choice(len(probs), p=probs)) if rng is not None else int(np.random.choice(len(probs), p=probs))
        else:
            t = int(lg.argmax())
        c = vocab.chars[t]
        if c == ".": break
        if c == "\n": break
        gen.append(c)
        x = torch.cat([x, torch.tensor([[t]])], dim=1)[:, -(CTX-1):]
    model.train()
    return "".join(gen).strip()

@torch.no_grad()
def eval_acc(model, vocab, prompts):
    """prompts: list of (prefix, expected, meta). Greedy decoding, exact match."""
    n = len(prompts); ok = 0; ok_p = np = 0; ok_e = ne = 0; fails = []
    for prefix, expected, meta in prompts:
        got = answer(model, vocab, prefix, temp=0.0)
        good = (got == expected)
        ok += good
        if meta.get("pron"): np += 1; ok_p += good
        else: ne += 1; ok_e += good
        if not good and len(fails) < 8: fails.append({"prefix_tail": prefix[-60:], "expected": expected, "got": got})
    return {"n": n, "acc": ok/max(1,n),
            "pron_n": np, "pron_acc": (ok_p/np if np else None),
            "expl_n": ne, "expl_acc": (ok_e/ne if ne else None), "fail_examples": fails}

def rss_mb():
    for l in open('/proc/self/status'):
        if l.startswith('VmRSS'): return int(l.split()[1])/1024
    return -1

def train_steps_mix(model, opt, arrays_probs, rng, n, lr=1e-3, warm=50, log_every=100, tag=""):
    """arrays_probs: list of (array, prob). Each window drawn from one array by prob."""
    model.train(); out=[]
    arrs = [a for a,_ in arrays_probs]; ps = np.array([p for _,p in arrays_probs]); ps = ps/ps.sum()
    for s in range(n):
        for g in opt.param_groups: g['lr'] = lr*min(1.0,(s+1)/warm)
        ai = int(rng.choice(len(arrs), p=ps))
        w = torch.from_numpy(sample_windows(rng, arrs[ai], BATCH))
        lg = model(w[:,:-1]); lo = F.cross_entropy(lg.reshape(-1,model.V), w[:,1:].reshape(-1))
        if not torch.isfinite(lo): raise RuntimeError("nonfinite loss")
        opt.zero_grad(set_to_none=True); lo.backward()
        float(torch.nn.utils.clip_grad_norm_(model.parameters(),1.0))
        opt.step(); out.append(float(lo.detach()))
        if (s+1)%log_every==0: print("  %s step %d/%d loss %.3f" % (tag, s+1, n, float(np.mean(out[-50:]))), flush=True)
    return out

# ---- atomic periodic checkpointing (weights+Adam+numpy RNG+step+config) ----
def save_ckpt(base, model, opt, rng, step, extra):
    import os
    st = dict(model=model.state_dict(), opt=opt.state_dict(), step=step, extra=extra)
    torch.save(st, base + '.ckpt.tmp'); os.replace(base + '.ckpt.tmp', base + '.ckpt')
    name, keys, pos, has_g, g = rng.get_state()
    np.savez(base + '.rng.tmp.npz', name=np.array(name), keys=keys, pos=np.array(pos),
             has_g=np.array(has_g), g=np.array(g))
    os.replace(base + '.rng.tmp.npz', base + '.rng.npz')

def load_ckpt(base, model, opt):
    st = torch.load(base + '.ckpt', weights_only=True)
    model.load_state_dict(st['model']); opt.load_state_dict(st['opt'])
    z = np.load(base + '.rng.npz', allow_pickle=False)
    rng = np.random.RandomState()
    rng.set_state((str(z['name']), z['keys'], int(z['pos']), int(z['has_g']), float(z['g'])))
    return rng, int(st['step']), st['extra']

def train_steps_ckpt(model, opt, arr, rng, n, ckpt_base=None, ckpt_every=100, s0=0, lr=1e-3, warm=50, log_every=100, tag=""):
    """Absolute-step lr (s0 offset) so resume is exact; periodic atomic checkpoint.
    arr: one id array, or [(array, prob), ...] to mix sources per window."""
    model.train(); out=[]
    if isinstance(arr, list):
        arrs = [a for a,_ in arr]; ps = np.array([p for _,p in arr]); ps = ps/ps.sum()
    else:
        arrs = [arr]; ps = np.array([1.0])
    for s in range(s0, s0+n):
        for g in opt.param_groups: g['lr'] = lr*min(1.0,(s+1)/warm)
        ai = int(rng.choice(len(arrs), p=ps)) if len(arrs) > 1 else 0
        w = torch.from_numpy(sample_windows(rng, arrs[ai], BATCH))
        lg = model(w[:,:-1]); lo = F.cross_entropy(lg.reshape(-1,model.V), w[:,1:].reshape(-1))
        if not torch.isfinite(lo): raise RuntimeError("nonfinite loss")
        opt.zero_grad(set_to_none=True); lo.backward()
        float(torch.nn.utils.clip_grad_norm_(model.parameters(),1.0))
        opt.step(); out.append(float(lo.detach()))
        if (s+1)%log_every==0: print("  %s step %d/%d loss %.3f" % (tag, s+1, s0+n, float(np.mean(out[-50:]))), flush=True)
        if ckpt_base and (s+1)%ckpt_every==0:
            save_ckpt(ckpt_base, model, opt, rng, s+1, dict(tag=tag))
    return out
