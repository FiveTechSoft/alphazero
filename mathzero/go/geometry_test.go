package main

import (
	"strconv"
	"testing"
)

// El puente de geometria traduce construcciones a polinomios del anillo de
// coeficientes enteros. Estos tests validan las FORMULAS con el JUEZ EXACTO
// (canonEq) y que el banco opt-in contiene identidades verdaderas. No afirman
// que la busqueda encuentre un camino de prueba (eso se mide aparte en
// GeomMain); el juez exacto es la fuente de verdad de la identidad.

func canonZero(e *Expr) bool { return len(Canon(e)) == 0 }

func TestGeoDistSqSymmetry(t *testing.T) {
	a := Point{X: V("a"), Y: V("b")}
	b := Point{X: V("c"), Y: V("d")}
	ab, ba := DistSq(a, b), DistSq(b, a)
	if !canonEq(Canon(ab), Canon(ba)) {
		t.Fatalf("DistSq no simetrica: %s vs %s", Render(ab), Render(ba))
	}
}

func TestGeoDotCommutative(t *testing.T) {
	x1, y1, x2, y2 := V("a"), V("b"), V("c"), V("d")
	if !canonEq(Canon(Dot(x1, y1, x2, y2)), Canon(Dot(x2, y2, x1, y1))) {
		t.Fatal("Dot no conmutativo")
	}
}

func TestGeoPerpendicularResidual(t *testing.T) {
	o := Point{X: C(0), Y: C(0)}
	a := Point{X: V("a"), Y: C(0)} // eje x
	b := Point{X: C(0), Y: V("b")} // eje y
	if r := IsPerpendicular(o, a, o, b); !canonZero(r) {
		t.Fatalf("segmentos perpendiculares deberian dar residual 0: %s", Render(r))
	}
	c := Point{X: V("a"), Y: V("b")}
	if r := IsPerpendicular(o, a, o, c); canonZero(r) {
		t.Fatal("segmentos no perpendiculares no deberian dar residual 0")
	}
}

func TestGeoCollinearResidual(t *testing.T) {
	o := Point{X: C(0), Y: C(0)}
	a := Point{X: V("a"), Y: C(0)}
	b := Point{X: V("b"), Y: C(0)}
	if r := IsCollinear(o, a, b); !canonZero(r) {
		t.Fatalf("puntos colineales deberian dar residual 0: %s", Render(r))
	}
	c := Point{X: V("a"), Y: V("b")}
	if r := IsCollinear(o, a, c); canonZero(r) {
		t.Fatal("puntos no colineales no deberian dar residual 0")
	}
}

func TestGeoMidpointScaling(t *testing.T) {
	// A=(0,0), B=(4,6): el punto medio es M=(2,3); TwiceMidpoint da A+B=(4,6)=2M.
	a := Point{X: C(0), Y: C(0)}
	b := Point{X: C(4), Y: C(6)}
	tw := TwiceMidpoint(a, b)
	if !canonEq(Canon(tw.X), Canon(C(4))) || !canonEq(Canon(tw.Y), Canon(C(6))) {
		t.Fatalf("TwiceMidpoint=%s,%s; se esperaba 4,6", Render(tw.X), Render(tw.Y))
	}
	// MidpointResidual con el punto medio REAL M=(2,3) debe anularse.
	m := Point{X: C(2), Y: C(3)}
	rx, ry := MidpointResidual(m, a, b)
	if !canonZero(rx) || !canonZero(ry) {
		t.Fatalf("MidpointResidual(M=(2,3)) no nulo: %s,%s", Render(rx), Render(ry))
	}
	// Documenta la trampa: pasar TwiceMidpoint (2M) NO da 0 (2*(2M)-A-B = A+B).
	badx, _ := MidpointResidual(tw, a, b)
	if canonZero(badx) {
		t.Fatal("MidpointResidual(TwiceMidpoint) no deberia anularse")
	}
}

// TestGeoFixturesAreIdentities: las 4 identidades del banco son verdaderas
// segun el juez exacto (forma canonica de polinomio), independientemente de
// si la busqueda alcanza su estructura objetivo.
func TestGeoFixturesAreIdentities(t *testing.T) {
	total := 0
	for lvl := 1; lvl <= 4; lvl++ {
		probs := LoadGeometryProblems(strconv.Itoa(lvl), 64)
		if len(probs) == 0 {
			t.Fatalf("nivel %d sin problemas", lvl)
		}
		for i, p := range probs {
			total++
			if !canonEq(Canon(p.Cur), Canon(p.Target)) {
				t.Errorf("nivel %d prob %d NO es identidad exacta", lvl, i)
			}
		}
	}
	if total != 4 {
		t.Fatalf("se esperaban 4 identidades, hay %d", total)
	}
}
