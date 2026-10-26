
LIMIT_MIB=640.0
def usage():
    rss=hwm=None
    for l in open('/proc/self/status'):
        if l.startswith('VmRSS'): rss=int(l.split()[1])/1024
        elif l.startswith('VmHWM'): hwm=int(l.split()[1])/1024
    return rss,hwm
def check(where=''):
    rss,hwm=usage()
    if hwm is None or hwm>LIMIT_MIB: raise RuntimeError('memguard: VmHWM %s MiB > %.0f at %s'%(hwm,LIMIT_MIB,where))
    return rss,hwm
