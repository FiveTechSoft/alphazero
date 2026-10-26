import math, torch, torch.nn as nn
import gt as G
def make(seed,nl=2):

    torch.manual_seed(seed); m=G.GT(64,4,nl)
    for mod in m.modules():
        if isinstance(mod,nn.Embedding): nn.init.normal_(mod.weight,0,0.02)
        if isinstance(mod,nn.Linear): nn.init.normal_(mod.weight,0,0.02); nn.init.zeros_(mod.bias)
    for b in m.blocks:
        nn.init.normal_(b.o.weight,0,0.02/math.sqrt(2*nl)); nn.init.normal_(b.f2.weight,0,0.02/math.sqrt(2*nl))
    return m
def prep(x,arm):
    x=x.copy()
    if arm=='nocue': x[:,19:24]=__import__('numpy').where((x[:,19:24]==G.DE)|(x[:,19:24]==G.A),G.PAD,x[:,19:24])
    return x
