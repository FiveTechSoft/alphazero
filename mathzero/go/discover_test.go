package main

import "testing"

func TestIsSingleRule(t *testing.T) {
	a := V("a")
	b := EMul(a, C(1))
	if !isSingleRule(b, a) {
		t.Fatal("mul1 deberia ser una sola regla")
	}
	compuesta := EMul(EAdd(a, C(0)), C(1))
	if isSingleRule(compuesta, a) {
		t.Fatal("(a+0)*1 -> a no es una sola regla (son dos pasos)")
	}
	if isSingleRule(a, a) {
		t.Fatal("una identidad vacia no puede ser un descubrimiento")
	}
}

func TestLibraryContains(t *testing.T) {
	L := NewLibrary()
	a, b := EMul(V("x"), C(1)), V("x")
	if !L.Insert(a, b, 2) {
		t.Fatal("Insert rechazo un lema valido")
	}
	if !L.Contains(b, a) {
		t.Fatal("Contains no encontro el lema en sentido inverso")
	}
	if L.Contains(a, V("y")) {
		t.Fatal("Contains invento un lema inexistente")
	}
}
