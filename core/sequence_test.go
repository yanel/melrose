package core

import (
	"fmt"
	"testing"
)

func ExampleSequenceParse() {
	m, _ := ParseSequence("C C4 4C4")
	fmt.Println(m)
	// Output:
	// C C C
}

func ExampleSequenceParseGroups() {
	m, _ := ParseSequence("C (E G)")
	m2, _ := ParseSequence("C ( A )")
	m3, _ := ParseSequence("2C# (8D_ 8E_ 2F#)")
	m4, _ := ParseSequence("(C E)(.D .F)(E G)")
	canto, _ := ParseSequence("B_ 8F 8D_5 8B_5 8F A_ 8E_ 8C5 8A_5 8E_")
	fmt.Println(m)
	fmt.Println(m2)
	fmt.Println(m3)
	fmt.Println(m4)
	fmt.Println(canto)
	// Output:
	// C (E G)
	// C A
	// ½C♯ (⅛D♭ ⅛E♭ ½F♯)
	// (C E) (.D .F) (E G)
	// B♭ ⅛F ⅛D♭5 ⅛B♭5 ⅛F A♭ ⅛E♭ ⅛C5 ⅛A♭5 ⅛E♭
}

func TestSequence_Storex(t *testing.T) {
	m, _ := ParseSequence("C (E G)")
	if got, want := m.Storex(), `sequence('C (E G)')`; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func TestSequence_Duration(t *testing.T) {
	m, _ := ParseSequence("C (E G)")
	if got, want := m.NoteLength(), 0.5; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
	m, _ = ParseSequence("e5 d#5 2.c#5")
	if got, want := m.NoteLength(), 1.25; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
}
