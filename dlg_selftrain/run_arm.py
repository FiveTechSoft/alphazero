import os, sys, json, time, random
import numpy as np, torch
import gen, lm

BASE_STEPS  = int(os.environ.get('DLGX_BASE_STEPS', 1500))
ROUNDS      = int(os.environ.get('DLGX_ROUNDS', 3))
ROUND_STEPS = int(os.environ.get('DLGX_ROUND_STEPS', 500))
N_PROMPTS   = int(os.environ.get('DLGX_N_PROMPTS', 800))
K_SAMPLES   = int(os.environ.get('DLGX_K_SAMPLES', 4))
TEMP        = float(os.environ.get('DLGX_TEMP', 0.9))
D0_N        = int(os.environ.get('DLGX_D0_N', 6000))

def log(path, rec):
    rec = dict(rec); rec['time'] = time.strftime('%H:%M:%S')
    with open(path, 'a') as f: f.write(json.dumps(rec) + "\n")

def round_prompts(r, n):
    # v2: seen pairs only — held-out pairs never appear in ANY training path
    prng = random.Random(6000 + r); out = []
    while len(out) < n:
        text, qa = gen.gen_dialogue(prng, False)
        for prefix, expected in qa:
            if len(prefix) <= 330:
                out.append((prefix, expected))
                if len(out) == n: break
    return out

def corpus_stats(texts):
    ans = set()
    for t in texts:
        a = t.rsplit('? ', 1)[-1].rstrip('.')
        ans.add(a)
    return dict(kept=len(texts), corpus_chars=sum(len(t) for t in texts),
                unique_answers=len(ans), unique_texts=len(set(texts)))

def main():
    t_arm0 = time.time()
    arm, seed = sys.argv[1], int(sys.argv[2])
    outdir = 'out/s%d/%s' % (seed, arm); os.makedirs(outdir, exist_ok=True)
    mpath = 'out/s%d/metrics.jsonl' % seed; os.makedirs('out/s%d' % seed, exist_ok=True)
    torch.manual_seed(seed)
    drng = random.Random(5000)                      # SAME D0 for all seeds/arms
    d0_qa = [gen.gen_dialogue(drng, False) for _ in range(D0_N)]
    d0_texts = [t for t, _ in d0_qa]
    d0_ban = set(p for _, qa in d0_qa for p, _ in qa)   # exact D0 question prefixes banned from eval
    vocab = lm.Vocab(d0_texts)
    d0_arr = lm.make_stream(vocab, d0_texts, random.Random(5001))
    model = lm.LM(vocab.size)
    opt = torch.optim.AdamW(model.parameters(), lr=1e-3, betas=(0.9,0.999), weight_decay=0.0)
    dev_seen  = gen.gen_eval_set(91001, 300, novel=False, seen_only_dialogues=True, ban_prefixes=d0_ban)
    dev_novel = gen.gen_eval_set(91002, 300, novel=True, ban_prefixes=d0_ban)
    def dev_eval(tag):
        for name, ps in (('dev_seen', dev_seen), ('dev_novel', dev_novel)):
            r = lm.eval_acc(model, vocab, ps)
            log(mpath, dict(arm=arm, seed=seed, stage=tag, split=name,
                            acc=r['acc'], n=r['n'], pron_acc=r['pron_acc'], expl_acc=r['expl_acc'],
                            fails=r['fail_examples'][:3]))
            print('EVAL', arm, seed, tag, name, round(r['acc'],3), flush=True)
    if arm == 'base':
        ck = 'out/s%d/base_ckpt' % seed
        if os.path.exists(ck + '.ckpt'):
            trng, done, extra = lm.load_ckpt(ck, model, opt)
            print('base RESUME from step', done, flush=True)
            log(mpath, dict(arm='base', seed=seed, stage='base_resume', from_step=done))
        else:
            trng = np.random.RandomState(seed); done = 0
        if done < BASE_STEPS:
            t0=time.time()
            lm.train_steps_ckpt(model, opt, d0_arr, trng, BASE_STEPS-done, s0=done,
                                ckpt_base=ck, ckpt_every=250, tag='base')
            log(mpath, dict(arm='base', seed=seed, stage='base_train', steps=BASE_STEPS,
                            from_step=done, seconds=round(time.time()-t0,1)))
        torch.save(model.state_dict(), 'out/s%d/base_model.pt' % seed)
        dev_eval('base')
        return
    model.load_state_dict(torch.load('out/s%d/base_model.pt' % seed, weights_only=True))
    if arm == 'b0':
        dev_eval('b0_final')
    else:
        for r in range(ROUNDS):
            prompts = round_prompts(r, N_PROMPTS)
            # v2: labels come from the INDEPENDENT parser gen.check, not generator bookkeeping
            checked = []; mism = 0
            for prefix, exp_book in prompts:
                try: exp = gen.check(prefix)
                except ValueError: continue
                if exp != exp_book: mism += 1
                checked.append((prefix, exp))
            if mism: print('WARN bookkeeping-vs-parser mismatches:', mism, flush=True)
            log(mpath, dict(arm=arm, seed=seed, stage='round%d_check' % r,
                            prompts=len(prompts), parse_ok=len(checked), mismatches=mism))
            prompts = checked
            srng = np.random.RandomState(seed*1000 + r)
            texts = []
            stats = dict(arm=arm, seed=seed, stage='round%d_gen' % r, prompts=len(prompts))
            t0 = time.time()
            if arm == 'sup':
                texts = [p + a + '.' for p, a in prompts]
                stats.update(samples_drawn=0)
            elif arm == 'self':
                n_any = 0; drawn = 0
                for prefix, expected in prompts:
                    kept = []
                    for k in range(K_SAMPLES):
                        s = lm.answer(model, vocab, prefix, temp=TEMP, rng=srng); drawn += 1
                        if s == expected: kept.append(s)
                    if kept: n_any += 1
                    for s in kept[:2]: texts.append(prefix + s + '.')
                stats.update(samples_drawn=drawn, prompts_with_correct=n_any,
                             accept_frac=round(n_any/len(prompts), 4))
            elif arm == 'selfuf':
                n_first_ok = 0
                for prefix, expected in prompts:
                    s = lm.answer(model, vocab, prefix, temp=TEMP, rng=srng)
                    if s == expected: n_first_ok += 1
                    texts.append(prefix + s + '.')
                stats.update(samples_drawn=len(prompts), first_sample_acc=round(n_first_ok/len(prompts), 4))
            else: raise SystemExit('unknown arm ' + arm)
            stats.update(corpus_stats(texts)); stats['gen_seconds'] = round(time.time()-t0, 1)
            log(mpath, stats); print('GEN', stats, flush=True)
            done_marker = outdir + '/model_r%d.pt' % r
            if os.path.exists(done_marker):
                print('round', r, 'already trained, loading (resume)', flush=True)
                model.load_state_dict(torch.load(done_marker, weights_only=True))
                log(mpath, dict(arm=arm, seed=seed, stage='round%d_resume_skip' % r))
                continue
            rarr = lm.make_stream(vocab, texts, random.Random(8000+r)) if texts else None
            if rarr is None or len(rarr) < lm.CTX + 2:
                log(mpath, dict(arm=arm, seed=seed, stage='round%d_skip' % r,
                                reason='insufficient_verified_data', train_seconds=0.0))
                print('SKIP round', arm, r, 'insufficient verified data', flush=True)
                torch.save(model.state_dict(), done_marker)   # marks round done (no update)
                dev_eval('round%d' % r)
                continue
            ck = outdir + '/ckpt_r%d' % r
            if os.path.exists(ck + '.ckpt'):
                trng, done, extra = lm.load_ckpt(ck, model, opt)
                print('round', r, 'RESUME from step', done, flush=True)
                log(mpath, dict(arm=arm, seed=seed, stage='round%d_resume' % r, from_step=done))
            else:
                trng = np.random.RandomState(seed*100 + r); done = 0
            if done < ROUND_STEPS:
                t0 = time.time()
                lm.train_steps_ckpt(model, opt, [(rarr, 0.5), (d0_arr, 0.5)], trng, ROUND_STEPS-done,
                                    s0=done, ckpt_base=ck, ckpt_every=100, tag='%s-r%d' % (arm, r))
                log(mpath, dict(arm=arm, seed=seed, stage='round%d_train' % r,
                                steps=ROUND_STEPS, from_step=done, seconds=round(time.time()-t0, 1)))
            torch.save(model.state_dict(), done_marker)
            dev_eval('round%d' % r)
    # RESERVED TEST: once per arm, at the end only
    for name, ps in (('test_seen', gen.gen_eval_set(92001, 400, novel=False, seen_only_dialogues=True, ban_prefixes=d0_ban)),
                     ('test_novel', gen.gen_eval_set(92002, 400, novel=True, ban_prefixes=d0_ban)),
                     ('test_novel_long', gen.gen_eval_set(92003, 300, novel=True, max_extra=8, ban_prefixes=d0_ban))):
        r = lm.eval_acc(model, vocab, ps)
        log(mpath, dict(arm=arm, seed=seed, stage='TEST', split=name,
                        acc=r['acc'], n=r['n'], pron_acc=r['pron_acc'], expl_acc=r['expl_acc'],
                        fails=r['fail_examples'][:5]))
        print('TEST', arm, seed, name, round(r['acc'],3), flush=True)
    log(mpath, dict(arm=arm, seed=seed, stage='arm_done', seconds=round(time.time()-t_arm0, 1)))

if __name__ == '__main__':
    main()
