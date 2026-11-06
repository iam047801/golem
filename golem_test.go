package golem

import (
	"reflect"
	"strings"
	"testing"
)

// TestDict is a pretend LanguagePack for testing purposes
type TestDict struct {
}

// GetResource shows the expected dictionary format.
// 1 line per base form, with additional forms to the right. Tab separated
func (t *TestDict) GetResource() ([]byte, error) {
	return []byte(`word	wordy	wordis	wordlike
thing	thingy	thingies	thingimagics`), nil
}

func (t *TestDict) GetLocale() string {
	return "test"
}

func TestTestDict(t *testing.T) {
	dict := &TestDict{}
	lem, err := New(dict)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := dict.GetResource()
	for _, line := range strings.Split(string(b), "\n") {
		words := strings.Split(line, "\t")
		base := words[0]
		for _, word := range words {
			lemma := lem.Lemma(word)
			if lemma != base {
				t.Errorf("Expected %v, but got %v for %v", base, lemma, word)
			}
		}
	}
}

func BenchmarkLookup(b *testing.B) {
	l, err := New(&TestDict{})
	if err != nil {
		b.Error(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N/2; i++ {
		l.Lemma("Thingimagics")
	}
}

func BenchmarkLookupLower(b *testing.B) {
	l, err := New(&TestDict{})
	if err != nil {
		b.Error(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N/2; i++ {
		l.LemmaLower("thingimagics")
	}
}

type overlappingDict struct{}

func (overlappingDict) GetResource() ([]byte, error) {
	return []byte("crear\tcrea\tcreo\tcreas\n" +
		"creer\tcree\tcreo\tcrees\n" +
		"coser\tcoso\tcosa\tcosas\n" +
		"cosa\tcosas"), nil
}

func (overlappingDict) GetLocale() string {
	return "es-fixture"
}

func TestOverlappingParadigms(t *testing.T) {
	lem, err := New(overlappingDict{})
	if err != nil {
		t.Fatal(err)
	}

	if got := lem.Lemmas("creo"); !reflect.DeepEqual(got, []string{"crear", "creer"}) {
		t.Errorf("Lemmas(creo) = %v, want every paradigm it belongs to", got)
	}
	if got := lem.Lemmas("cosas"); !reflect.DeepEqual(got, []string{"cosa", "coser"}) {
		t.Errorf("Lemmas(cosas) = %v", got)
	}

	if !lem.IsLemma("Cosa") || !lem.IsLemmaLower("cosa") {
		t.Error("cosa heads its own paradigm")
	}
	if lem.IsLemma("creo") {
		t.Error("creo heads no paradigm")
	}
	if got := lem.Lemma("creo"); got != "crear" {
		t.Errorf("Lemma(creo) = %v, want the first paradigm's base", got)
	}

	first := lem.Lemmas("creo")
	first[0] = "mutated"
	if got := lem.Lemmas("creo"); got[0] != "crear" {
		t.Errorf("Lemmas leaked its internal slice: %v", got)
	}
	if got := lem.Lemma("cosas"); got != "coser" {
		t.Errorf("Lemma(cosas) = %v, sorting in Lemmas must not reorder the first-found base", got)
	}
}
