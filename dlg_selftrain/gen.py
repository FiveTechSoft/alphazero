# Dialogue generator + independent parser/checker for state-tracking dialogues.
# World: objects with gender x colors. Turns: statements, explicit/pronoun updates,
# explicit/pronoun questions. All text lowercase, no accents (char-level model).
import json, random, re

OBJS = [("caja","f"),("pelota","f"),("mesa","f"),("silla","f"),("taza","f"),("mochila","f"),
        ("libro","m"),("vaso","m"),("coche","m"),("globo","m"),("cuaderno","m"),("lapiz","m")]
COLS = ["rojo","azul","verde","amarillo","negro","blanco","gris","morado"]
INVARIANT = {"azul","verde","gris"}          # same form for both genders
HELD_OUT = {("pelota","verde"),("libro","blanco"),("taza","gris"),
            ("coche","amarillo"),("mesa","morado"),("globo","negro")}

def art(g): return "la" if g=="f" else "el"
def agree(col,g): return col if col in INVARIANT else (col[:-1]+"a" if g=="f" else col)
OG = dict(OBJS)

# ---------- generator (bookkeeping) ----------
def gen_dialogue(rng, allow_heldout, min_q=1, max_extra=None):
    """Returns (text, qa_list). qa_list: list of (prefix_text_ends_after_'? ', expected_answer)."""
    objs = rng.sample(OBJS, rng.choice([2,2,3]))
    state, last = {}, None
    pairs = [(o[0],c) for o in OBJS for c in COLS if allow_heldout or (o[0],c) not in HELD_OUT]
    def cols_for(name): return [c for (oo,c) in pairs if oo==name]
    turns = []
    for name,g in objs:
        col = rng.choice(cols_for(name))
        state[name] = col; last = name
        turns.append("%s %s es %s" % (art(g), name, agree(col,g)))
    n_extra = rng.randint(2, max_extra if max_extra else 6)
    qa = []
    nq = 0
    for t in range(n_extra):
        last_turn = (t == n_extra-1)
        want_q = last_turn and nq < min_q
        kind = "q" if want_q else rng.choice(["q","q","u","u","s"])
        if kind == "s":
            name,g = rng.choice(objs)
            if name != last or rng.random() < 0.5:
                col = rng.choice(cols_for(name))
                state[name] = col; last = name
                turns.append("%s %s es %s" % (art(g), name, agree(col,g)))
            continue
        if kind == "u":
            name,g = rng.choice(objs)
            col = rng.choice(cols_for(name))
            pron = (name == last) and rng.random() < 0.5
            state[name] = col
            if pron: turns.append("ahora es %s" % agree(col,g))
            else:
                if rng.random() < 0.5: turns.append("ahora %s %s es %s" % (art(g), name, agree(col,g)))
                else: turns.append("pinta %s %s de %s" % (art(g), name, col))
            last = name
            continue
        pron = rng.random() < 0.4
        pre = ". ".join(turns) + (". " if turns else "")
        if pron:
            g = OG[last]
            qa.append((pre + "de que color es ahora? ", agree(state[last], g)))
            turns.append("de que color es ahora? %s" % agree(state[last], g))
        else:
            name,g = rng.choice(objs)
            qa.append((pre + "de que color es %s %s? " % (art(g), name), agree(state[name], g)))
            turns.append("de que color es %s %s? %s" % (art(g), name, agree(state[name], g)))
            last = name
        nq += 1
    text = ". ".join(turns) + "."
    return text, qa

# ---------- independent parser/checker ----------
RE_SET  = re.compile(r"^(el|la) ([a-z]+) es ([a-z]+)$")
RE_USET = re.compile(r"^ahora (el|la) ([a-z]+) es ([a-z]+)$")
RE_PINT = re.compile(r"^pinta (el|la) ([a-z]+) de ([a-z]+)$")
RE_UPRO = re.compile(r"^ahora es ([a-z]+)$")
RE_Q    = re.compile(r"^de que color es (el|la) ([a-z]+)\?$")
RE_QPRO = re.compile(r"^de que color es ahora\?$")
def base_color(w):
    if w in INVARIANT: return w
    if w.endswith("a") and w[:-1] in COLS: return w[:-1]
    return w

def _parse_turns(parts):
    state, last = {}, None
    for s in parts:
        s = s.strip()
        if not s: continue
        m = RE_SET.match(s) or RE_USET.match(s)
        if m:
            state[m.group(2)] = base_color(m.group(3)); last = m.group(2); continue
        m = RE_PINT.match(s)
        if m:
            state[m.group(2)] = m.group(3); last = m.group(2); continue
        m = RE_UPRO.match(s)
        if m:
            state[last] = base_color(m.group(1)); continue
        if "? " in s:
            qpart = s.split("? ", 1)[0]
            m2 = RE_Q.match(qpart+"?")
            if m2: last = m2.group(2)
            continue
        raise ValueError("unparseable turn: %r" % s)
    return state, last

def check(prefix):
    """Parse a prefix ending in '? ' and return expected answer string. Raises on unparseable."""
    assert prefix.endswith("? "), prefix[-20:]
    body = prefix[:-1]          # drop trailing space, keep "?"
    parts = body.split(". ")
    q = parts[-1]
    state, last = _parse_turns(parts[:-1])
    m = RE_Q.match(q)
    if m: return agree(state[m.group(2)], OG[m.group(2)])
    m = RE_QPRO.match(q)
    if m: return agree(state[last], OG[last])
    raise ValueError("unparseable question: %r" % q)

def selftest(n=4000, seed=99):
    rng = random.Random(seed)
    bad = 0
    for i in range(n):
        text, qa = gen_dialogue(rng, allow_heldout=True)
        for prefix, expected in qa:
            try:
                got = check(prefix)
            except Exception as e:
                bad += 1
                if bad <= 3: print("PARSEFAIL", e, "|", prefix[-80:])
                continue
            if got != expected:
                bad += 1
                if bad <= 3: print("MISMATCH gen=%r parse=%r | %s" % (expected, got, prefix[-80:]))
    print("selftest n=%d bad=%d" % (n, bad))
    return bad == 0

if __name__ == "__main__":
    rng = random.Random(7)
    for _ in range(3):
        t, qa = gen_dialogue(rng, allow_heldout=True)
        print(t); print("  QA:", [(p[-45:], a) for p,a in qa]); print()
    assert selftest()

# ---------- eval prompt sets ----------
def iter_qa(rng, allow_heldout, max_extra=None):
    """Yields (prefix, expected, meta) one per question in each dialogue."""
    while True:
        text, qa = gen_dialogue(rng, allow_heldout, max_extra=max_extra)
        # recover asked object per prefix for meta
        for prefix, expected in qa:
            yield prefix, expected

def asked_pair(prefix, expected):
    """Returns ((obj, base_color), is_pronoun). Pronoun questions resolved via parser state."""
    q = prefix[:-1].split(". ")[-1]
    m = RE_Q.match(q)
    if m: return (m.group(2), base_color(expected)), False
    if RE_QPRO.match(q):
        parts = prefix[:-1].split(". ")
        state, last = _parse_turns(parts[:-1])
        return (last, base_color(expected)), True
    raise ValueError("bad question: %r" % q)

def gen_eval_set(seed, n, novel, max_extra=None, seen_only_dialogues=False, ban_prefixes=None):
    """n prompts, deduplicated within set and against ban_prefixes.
    novel=True: keep questions whose resolved asked (obj,color) pair is held out
    (pronoun questions included, referent resolved by the parser).
    novel=False: keep only asked pairs NOT held out; with seen_only_dialogues=True
    the whole dialogue contains no held-out pair (zero exposure)."""
    rng = random.Random(seed)
    out = []; used = set(ban_prefixes or [])
    while len(out) < n:
        text, qa = gen_dialogue(rng, allow_heldout=not seen_only_dialogues, max_extra=max_extra)
        for prefix, expected in qa:
            if len(prefix) > 330: continue
            pair, pron = asked_pair(prefix, expected)
            if novel and pair not in HELD_OUT: continue
            if not novel and pair in HELD_OUT: continue
            if prefix in used: continue
            used.add(prefix)
            out.append((prefix, expected, {"pron": pron, "pair": pair}))
            if len(out) == n: break
    return out
