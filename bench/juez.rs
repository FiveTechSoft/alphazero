// juez.rs - Port a Rust del juez exacto de MathRing (forma canonica de
// polinomio, coeficientes enteros, mismas 14 reglas que Go/Python).
// Verifica 0 discrepancias contra juez_vectores.txt (generado por Go)
// y mide velocidad: juicios canon/s y nodos BFS/s.
//
//   rustc -O juez.rs -o juez_rust
//   ./juez_rust [ruta a juez_vectores.txt]
use std::collections::{HashMap, HashSet, VecDeque};
use std::fs;
use std::time::Instant;

#[derive(Clone, Debug)]
enum Expr {
    Const(i64),
    Var(String),
    Neg(Box<Expr>),
    Add(Box<Expr>, Box<Expr>),
    Mul(Box<Expr>, Box<Expr>),
}
use Expr::*;

fn bx(e: Expr) -> Box<Expr> { Box::new(e) }
fn is0(x: &Expr) -> bool { matches!(x, Const(0)) }
fn is1(x: &Expr) -> bool { matches!(x, Const(1)) }

// ---- parser del formato prefijo: c5, va, neg(e), add(e,e), mul(e,e) ----
struct P<'a> { b: &'a [u8], i: usize }
impl<'a> P<'a> {
    fn int(&mut self) -> i64 {
        let mut n: i64 = 0; let mut sgn = 1;
        if self.b[self.i] == b'-' { sgn = -1; self.i += 1; }
        while self.i < self.b.len() && self.b[self.i].is_ascii_digit() {
            n = n * 10 + (self.b[self.i] - b'0') as i64; self.i += 1;
        }
        sgn * n
    }
    fn ident(&mut self) -> String {
        let st = self.i;
        while self.i < self.b.len() && (self.b[self.i].is_ascii_alphanumeric()) { self.i += 1; }
        String::from_utf8_lossy(&self.b[st..self.i]).into_owned()
    }
    fn expect(&mut self, ch: u8) { assert_eq!(self.b[self.i], ch, "parse error en {}", self.i); self.i += 1; }
    fn expr(&mut self) -> Box<Expr> {
        match self.b[self.i] {
            b'c' => { self.i += 1; bx(Const(self.int())) }
            b'v' => { self.i += 1; bx(Var(self.ident())) }
            _ => {
                let name = self.ident();
                self.expect(b'(');
                let a = self.expr();
                if name == "neg" { self.expect(b')'); bx(Neg(a)) } else {
                    self.expect(b',');
                    let b2 = self.expr();
                    self.expect(b')');
                    if name == "add" { bx(Add(a, b2)) } else { bx(Mul(a, b2)) }
                }
            }
        }
    }
}
fn parse(s: &str) -> Box<Expr> { let mut p = P { b: s.as_bytes(), i: 0 }; p.expr() }

// ---- juez exacto: forma canonica monomio -> coef ----
fn merge_mon(a: &str, b: &str) -> String {
    if a.is_empty() { return b.to_string(); }
    if b.is_empty() { return a.to_string(); }
    let mut vs: Vec<&str> = a.split('*').chain(b.split('*')).collect();
    vs.sort();
    vs.join("*")
}

fn canon(e: &Expr) -> HashMap<String, i64> {
    match e {
        Const(n) => { let mut m = HashMap::new(); if *n != 0 { m.insert(String::new(), *n); } m }
        Var(s) => { let mut m = HashMap::new(); m.insert(s.clone(), 1); m }
        Neg(a) => { let mut r = canon(a); for v in r.values_mut() { *v = -*v; } r }
        Add(a, b2) => {
            let mut r = canon(a);
            for (m2, c2) in canon(b2) { *r.entry(m2).or_insert(0) += c2; }
            r.retain(|_, v| *v != 0);
            r
        }
        Mul(a, b2) => {
            let ca = canon(a); let cb = canon(b2);
            let mut r: HashMap<String, i64> = HashMap::new();
            for (m1, c1) in &ca { for (m2, c2) in &cb {
                *r.entry(merge_mon(m1, m2)).or_insert(0) += c1 * c2;
            }}
            r.retain(|_, v| *v != 0);
            r
        }
    }
}
fn canon_eq(a: &HashMap<String, i64>, b: &HashMap<String, i64>) -> bool { a == b }

// ---- clave estructural normalizada ----
fn key(e: &Expr) -> String {
    match e {
        Const(n) => format!("c{}", n),
        Var(s) => format!("v{}", s),
        Neg(a) => format!("neg({})", key(a)),
        Add(..) | Mul(..) => {
            let is_add = matches!(e, Add(..));
            let mut parts: Vec<String> = Vec::new();
            fn rec(x: &Expr, is_add: bool, parts: &mut Vec<String>) {
                match x {
                    Add(a, b2) if is_add => { rec(a, is_add, parts); rec(b2, is_add, parts); }
                    Mul(a, b2) if !is_add => { rec(a, is_add, parts); rec(b2, is_add, parts); }
                    _ => parts.push(key(x)),
                }
            }
            rec(e, is_add, &mut parts);
            parts.sort();
            format!("{}({})", if is_add { "add" } else { "mul" }, parts.join(","))
        }
    }
}

// ---- posiciones y sustitucion ----
fn sub_positions(e: &Expr) -> Vec<Vec<usize>> {
    let mut out: Vec<Vec<usize>> = vec![vec![]];
    fn rec(x: &Expr, p: &[usize], out: &mut Vec<Vec<usize>>) {
        match x {
            Add(a, b2) | Mul(a, b2) => {
                let mut p0 = p.to_vec(); p0.push(0);
                let mut p1 = p.to_vec(); p1.push(1);
                out.push(p0.clone()); out.push(p1.clone());
                rec(a, &p0, out); rec(b2, &p1, out);
            }
            Neg(a) => {
                let mut p0 = p.to_vec(); p0.push(0);
                out.push(p0.clone());
                rec(a, &p0, out);
            }
            _ => {}
        }
    }
    rec(e, &[], &mut out);
    out
}

fn get<'a>(mut e: &'a Expr, p: &[usize]) -> &'a Expr {
    for &i in p {
        e = match e {
            Add(a, b2) | Mul(a, b2) => if i == 0 { a } else { b2 },
            Neg(a) => a,
            _ => e,
        };
    }
    e
}

fn replace(e: &Expr, p: &[usize], nw: Box<Expr>) -> Box<Expr> {
    if p.is_empty() { return nw; }
    match e {
        Neg(a) => bx(Neg(replace(a, &p[1..], nw))),
        Add(a, b2) => if p[0] == 0 { bx(Add(replace(a, &p[1..], nw), b2.clone())) } else { bx(Add(a.clone(), replace(b2, &p[1..], nw))) },
        Mul(a, b2) => if p[0] == 0 { bx(Mul(replace(a, &p[1..], nw), b2.clone())) } else { bx(Mul(a.clone(), replace(b2, &p[1..], nw))) },
        _ => bx(e.clone()),
    }
}

// ---- reglas: mismas 14 que Go ----
type RF = fn(&Expr) -> Option<Box<Expr>>;
struct Rule { name: &'static str, f: Option<RF>, r: Option<RF> }

fn f_add0(x: &Expr) -> Option<Box<Expr>> {
    if let Add(a, b2) = x { if is0(b2) { return Some(a.clone()); } if is0(a) { return Some(b2.clone()); } }
    None
}
fn r_add0(x: &Expr) -> Option<Box<Expr>> { Some(bx(Add(bx(x.clone()), bx(Const(0))))) }
fn f_mul1(x: &Expr) -> Option<Box<Expr>> {
    if let Mul(a, b2) = x { if is1(b2) { return Some(a.clone()); } if is1(a) { return Some(b2.clone()); } }
    None
}
fn r_mul1(x: &Expr) -> Option<Box<Expr>> { Some(bx(Mul(bx(x.clone()), bx(Const(1))))) }
fn f_mul0(x: &Expr) -> Option<Box<Expr>> {
    if let Mul(a, b2) = x { if is0(a) || is0(b2) { return Some(bx(Const(0))); } }
    None
}
fn f_comm_add(x: &Expr) -> Option<Box<Expr>> {
    if let Add(a, b2) = x { return Some(bx(Add(b2.clone(), a.clone()))); } None
}
fn f_comm_mul(x: &Expr) -> Option<Box<Expr>> {
    if let Mul(a, b2) = x { return Some(bx(Mul(b2.clone(), a.clone()))); } None
}
fn f_assoc_add(x: &Expr) -> Option<Box<Expr>> {
    if let Add(a, b2) = x { if let Add(x_, y_) = a.as_ref() {
        return Some(bx(Add(x_.clone(), bx(Add(y_.clone(), b2.clone()))))); } }
    None
}
fn r_assoc_add(x: &Expr) -> Option<Box<Expr>> {
    if let Add(a, b2) = x { if let Add(x_, y_) = b2.as_ref() {
        return Some(bx(Add(bx(Add(a.clone(), x_.clone())), y_.clone()))); } }
    None
}
fn f_assoc_mul(x: &Expr) -> Option<Box<Expr>> {
    if let Mul(a, b2) = x { if let Mul(x_, y_) = a.as_ref() {
        return Some(bx(Mul(x_.clone(), bx(Mul(y_.clone(), b2.clone()))))); } }
    None
}
fn r_assoc_mul(x: &Expr) -> Option<Box<Expr>> {
    if let Mul(a, b2) = x { if let Mul(x_, y_) = b2.as_ref() {
        return Some(bx(Mul(bx(Mul(a.clone(), x_.clone())), y_.clone()))); } }
    None
}
fn f_dist(x: &Expr) -> Option<Box<Expr>> {
    if let Mul(a, b2) = x { if let Add(x_, y_) = b2.as_ref() {
        return Some(bx(Add(bx(Mul(a.clone(), x_.clone())), bx(Mul(a.clone(), y_.clone()))))); } }
    None
}
fn r_dist(x: &Expr) -> Option<Box<Expr>> {
    if let Add(a, b2) = x { if let (Mul(aa, ab), Mul(ba, bb)) = (a.as_ref(), b2.as_ref()) {
        if key(aa) == key(ba) { return Some(bx(Mul(aa.clone(), bx(Add(ab.clone(), bb.clone()))))); } } }
    None
}
fn f_negneg(x: &Expr) -> Option<Box<Expr>> {
    if let Neg(a) = x { if let Neg(aa) = a.as_ref() { return Some(aa.clone()); } } None
}
fn r_negneg(x: &Expr) -> Option<Box<Expr>> { Some(bx(Neg(bx(Neg(bx(x.clone())))))) }
fn f_neg_add(x: &Expr) -> Option<Box<Expr>> {
    if let Neg(a) = x { if let Add(x_, y_) = a.as_ref() {
        return Some(bx(Add(bx(Neg(x_.clone())), bx(Neg(y_.clone()))))); } } None
}
fn r_neg_add(x: &Expr) -> Option<Box<Expr>> {
    if let Add(a, b2) = x { if let (Neg(aa), Neg(ba)) = (a.as_ref(), b2.as_ref()) {
        return Some(bx(Neg(bx(Add(aa.clone(), ba.clone()))))); } } None
}
fn f_sub0(x: &Expr) -> Option<Box<Expr>> {
    if let Add(a, b2) = x {
        if let Neg(ba) = b2.as_ref() { if key(a) == key(ba) { return Some(bx(Const(0))); } }
        if let Neg(aa) = a.as_ref() { if key(b2) == key(aa) { return Some(bx(Const(0))); } }
    }
    None
}
fn f_neg_mul(x: &Expr) -> Option<Box<Expr>> {
    if let Neg(a) = x { if let Mul(x_, y_) = a.as_ref() {
        return Some(bx(Mul(bx(Neg(x_.clone())), y_.clone()))); } } None
}
fn r_neg_mul(x: &Expr) -> Option<Box<Expr>> {
    if let Mul(a, b2) = x { if let Neg(aa) = a.as_ref() {
        return Some(bx(Neg(bx(Mul(aa.clone(), b2.clone()))))); } } None
}
fn f_fold_add(x: &Expr) -> Option<Box<Expr>> {
    if let Add(a, b2) = x { if let (Const(m), Const(n)) = (a.as_ref(), b2.as_ref()) {
        return Some(bx(Const(m + n))); } } None
}
fn f_fold_mul(x: &Expr) -> Option<Box<Expr>> {
    if let Mul(a, b2) = x { if let (Const(m), Const(n)) = (a.as_ref(), b2.as_ref()) {
        return Some(bx(Const(m * n))); } } None
}

fn rules() -> [Rule; 14] {
    [
        Rule { name: "add0",   f: Some(f_add0),     r: Some(r_add0) },
        Rule { name: "mul1",   f: Some(f_mul1),     r: Some(r_mul1) },
        Rule { name: "mul0",   f: Some(f_mul0),     r: None },
        Rule { name: "comm+",  f: Some(f_comm_add), r: None },
        Rule { name: "comm*",  f: Some(f_comm_mul), r: None },
        Rule { name: "assoc+", f: Some(f_assoc_add),r: Some(r_assoc_add) },
        Rule { name: "assoc*", f: Some(f_assoc_mul),r: Some(r_assoc_mul) },
        Rule { name: "dist",   f: Some(f_dist),     r: Some(r_dist) },
        Rule { name: "negneg", f: Some(f_negneg),   r: Some(r_negneg) },
        Rule { name: "neg+",   f: Some(f_neg_add),  r: Some(r_neg_add) },
        Rule { name: "+-0",    f: Some(f_sub0),     r: None },
        Rule { name: "neg*",   f: Some(f_neg_mul),  r: Some(r_neg_mul) },
        Rule { name: "fold+",  f: Some(f_fold_add), r: None },
        Rule { name: "fold*",  f: Some(f_fold_mul), r: None },
    ]
}

// ---- acciones y BFS (misma semantica y orden que Go) ----
struct MAction { rule: usize, pos: Vec<usize> }

fn math_actions(e: &Expr, rs: &[Rule; 14]) -> Vec<MAction> {
    let mut out = Vec::new();
    for p in sub_positions(e) {
        let sub = get(e, &p);
        let ksub = key(sub);
        for (i, rl) in rs.iter().enumerate() {
            for fnc in [rl.f, rl.r] {
                let Some(fnc) = fnc else { continue };
                if let Some(nw) = fnc(sub) {
                    if key(&nw) != ksub { out.push(MAction { rule: i, pos: p.clone() }); }
                }
            }
        }
    }
    out
}

fn apply_action(e: &Expr, a: &MAction, rs: &[Rule; 14]) -> Box<Expr> {
    let sub = get(e, &a.pos);
    let ksub = key(sub);
    let rl = &rs[a.rule];
    for fnc in [rl.f, rl.r] {
        let Some(fnc) = fnc else { continue };
        if let Some(nw) = fnc(sub) {
            if key(&nw) != ksub { return replace(e, &a.pos, nw); }
        }
    }
    bx(e.clone())
}

fn bfs_solve(cur: &Expr, target: &Expr, cap: usize, rs: &[Rule; 14]) -> (bool, usize) {
    let tkey = key(target);
    let k0 = key(cur);
    let mut seen: HashSet<String> = HashSet::new();
    seen.insert(k0.clone());
    let mut queue: VecDeque<(Box<Expr>, String)> = VecDeque::new();
    queue.push_back((bx(cur.clone()), k0));
    let mut nodes = 0usize;
    while !queue.is_empty() && nodes < cap {
        let (cur, k) = queue.pop_front().unwrap();
        nodes += 1;
        if k == tkey { return (true, nodes); }
        for a in math_actions(&cur, rs) {
            let nxt = apply_action(&cur, &a, rs);
            let k2 = key(&nxt);
            if seen.insert(k2.clone()) { queue.push_back((nxt, k2)); }
        }
    }
    (false, nodes)
}

fn main() {
    let path = std::env::args().nth(1).unwrap_or_else(|| "juez_vectores.txt".to_string());
    let data = fs::read_to_string(&path).expect("no se puede leer juez_vectores.txt");
    struct Vec5 { a: Box<Expr>, b: Box<Expr>, eq: bool, solved: bool, nodes: usize }
    let mut vecs: Vec<Vec5> = Vec::new();
    for line in data.lines() {
        if line.trim().is_empty() { continue; }
        let parts: Vec<&str> = line.split(" ; ").collect();
        assert!(parts.len() == 5, "linea mal formada: {}", line);
        vecs.push(Vec5 {
            a: parse(parts[0].trim()), b: parse(parts[1].trim()),
            eq: parts[2].trim() == "1", solved: parts[3].trim() == "1",
            nodes: parts[4].trim().parse().unwrap(),
        });
    }
    let rs = rules();
    // 1) correccion: 0 discrepancias exigidas
    let mut bad = 0;
    for (i, v) in vecs.iter().enumerate() {
        let eq2 = canon_eq(&canon(&v.a), &canon(&v.b));
        let (s2, n2) = bfs_solve(&v.a, &v.b, 2000, &rs);
        if eq2 != v.eq || s2 != v.solved || n2 != v.nodes {
            bad += 1;
            eprintln!("DISCREPANCIA #{}: esperado eq={} solved={} nodos={}, obtenido eq={} solved={} nodos={}",
                i, v.eq, v.solved, v.nodes, eq2, s2, n2);
        }
    }
    println!("[juez Rust] {} vectores verificados, discrepancias = {}", vecs.len(), bad);
    // 2) velocidad canon
    let reps = 200;
    let t = Instant::now();
    let mut n: u64 = 0;
    for _ in 0..reps {
        for v in &vecs { let _ = canon_eq(&canon(&v.a), &canon(&v.b)); n += 1; }
    }
    let el = t.elapsed().as_secs_f64();
    println!("[juez Rust] {} juicios canon en {:.3}s -> {:.0} juicios/s", n, el, n as f64 / el);
    // 3) velocidad BFS
    let t = Instant::now();
    let mut nodes: u64 = 0;
    for v in &vecs { let (_, nd) = bfs_solve(&v.a, &v.b, 2000, &rs); nodes += nd as u64; }
    let el = t.elapsed().as_secs_f64();
    println!("[juez Rust] BFS {} nodos en {:.3}s -> {:.0} nodos/s", nodes, el, nodes as f64 / el);
      }
