// Parser del formato prefijo compartido con juez_vectores.txt y el port
// Rust: c5, va, neg(e), add(e,e), mul(e,e).
package main

import "strconv"

type pparser struct {
	s string
	i int
}

func (p *pparser) peek() byte { return p.s[p.i] }

func (p *pparser) expr() *Expr {
	switch p.peek() {
	case 'c':
		p.i++
		start := p.i
		if p.peek() == '-' {
			p.i++
		}
		for p.i < len(p.s) && p.s[p.i] >= '0' && p.s[p.i] <= '9' {
			p.i++
		}
		n, _ := strconv.Atoi(p.s[start:p.i])
		return C(n)
	case 'v':
		p.i++
		start := p.i
		for p.i < len(p.s) && (p.s[p.i] >= 'a' && p.s[p.i] <= 'z' || p.s[p.i] >= '0' && p.s[p.i] <= '9') {
			p.i++
		}
		return V(p.s[start:p.i])
	}
	start := p.i
	for p.i < len(p.s) && p.s[p.i] != '(' {
		p.i++
	}
	op := p.s[start:p.i]
	p.i++ // '('
	a := p.expr()
	if op == "neg" {
		p.i++ // ')'
		return ENeg(a)
	}
	p.i++ // ','
	b := p.expr()
	p.i++ // ')'
	if op == "add" {
		return EAdd(a, b)
	}
	return EMul(a, b)
}

func parsePrefix(s string) *Expr {
	p := &pparser{s: s}
	return p.expr()
}
