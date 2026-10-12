// Port del bucle caliente: MCTS PUCT sobre tres en raya, sin red
// (prior uniforme), mismo algoritmo y parametros que engine.go y
// bench_ttt.c: 200 partidas x 100 sims, c_puct=1.5.
use std::time::Instant;

const CP: f64 = 1.5;
const WINS: [[usize; 3]; 8] = [[0,1,2],[3,4,5],[6,7,8],[0,3,6],[1,4,7],[2,5,8],[0,4,8],[2,4,6]];
const NIL: i32 = -1;

#[derive(Clone, Default)]
struct Node {
    acts: [i32; 9],
    nacts: usize,
    p: [f64; 9],
    w: [f64; 9],
    n: [i32; 9],
    children: [i32; 9],
    expanded: bool,
}

fn terminal(b: &[i8; 9]) -> i32 {
    for w in WINS.iter() {
        let s = b[w[0]] + b[w[1]] + b[w[2]];
        if s == 3 || s == -3 { return -1; }
    }
    if b.iter().any(|&v| v == 0) { 2 } else { 0 }
}

fn expand(arena: &mut Vec<Node>, ni: usize, b: &[i8; 9]) {
    let mut k = 0;
    for i in 0..9 {
        if b[i] == 0 {
            arena[ni].acts[k] = i as i32;
            arena[ni].p[k] = 0.0;
            arena[ni].w[k] = 0.0;
            arena[ni].n[k] = 0;
            arena[ni].children[k] = NIL;
            k += 1;
        }
    }
    arena[ni].nacts = k;
    for i in 0..k { arena[ni].p[i] = 1.0 / k as f64; }
    arena[ni].expanded = true;
}

fn mcts(arena: &mut Vec<Node>, b: &[i8; 9], player: i8, sims: i32) -> i32 {
    let root = arena.len();
    arena.push(Node::default());
    expand(arena, root, b);
    for _ in 0..sims {
        let mut cb = *b;
        let mut pl = player;
        let mut path: Vec<(usize, usize)> = Vec::with_capacity(16);
        let mut node = root;
        let mut t = terminal(&cb);
        let v;
        while t == 2 && arena[node].expanded {
            let total: i32 = arena[node].n[..arena[node].nacts].iter().sum();
            let mut best = 0usize;
            let mut bs = -1e18f64;
            for j in 0..arena[node].nacts {
                let q = if arena[node].n[j] > 0 { arena[node].w[j] / arena[node].n[j] as f64 } else { 0.0 };
                let u = CP * arena[node].p[j] * ((total + 1) as f64).sqrt() / (1 + arena[node].n[j]) as f64;
                if q + u > bs { bs = q + u; best = j; }
            }
            path.push((node, best));
            cb[arena[node].acts[best] as usize] = pl;
            pl = -pl;
            if arena[node].children[best] == NIL {
                arena[node].children[best] = arena.len() as i32;
                arena.push(Node::default());
            }
            node = arena[node].children[best] as usize;
            t = terminal(&cb);
        }
        if t == 2 { v = 0.0; expand(arena, node, &cb); } else { v = t as f64; }
        let mut vv = v;
        for &(nd, j) in path.iter().rev() {
            vv = -vv;
            arena[nd].n[j] += 1;
            arena[nd].w[j] += vv;
        }
    }
    let mut best = 0usize;
    let mut bn = -1;
    for j in 0..arena[root].nacts {
        if arena[root].n[j] > bn { bn = arena[root].n[j]; best = j; }
    }
    arena[root].acts[best]
}

fn main() {
    let games = 200;
    let sims = 100;
    let mut total: i64 = 0;
    let mut arena: Vec<Node> = Vec::with_capacity(400000);
    let t0 = Instant::now();
    for _ in 0..games {
        let mut b = [0i8; 9];
        let mut pl = 1i8;
        while terminal(&b) == 2 {
            let a = mcts(&mut arena, &b, pl, sims);
            total += sims as i64;
            b[a as usize] = pl;
            pl = -pl;
        }
    }
    let el = t0.elapsed().as_secs_f64();
    println!("[benchttt Rust] {} partidas x {} sims = {} sims en {:.3}s -> {:.0} sims/s ({:.1} partidas/s)",
             games, sims, total, el, total as f64 / el, games as f64 / el);
}
