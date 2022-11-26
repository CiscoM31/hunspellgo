package hunspellgo

import "testing"

func TestExtendedSpell(t *testing.T) {
	h := Hunspell("testdics/de_DE.aff", "testdics/de_DE.dic")
	encodingtests := []struct {
		in       string
		expected bool
	}{
		{"Maßstäbe", true},
		{"Maßstab", true},
		{"Maßstäb", false},
		{"das", true},
		{"für", true},
	}
	for _, st := range encodingtests {
		if st.expected != h.Spell(st.in) {
			t.Errorf("Unexpected Spell result expected %v for %v", st.expected, st.in)
		}
	}
}

func TestExtendedSuggest(t *testing.T) {
	h := Hunspell("testdics/de_DE.aff", "testdics/de_DE.dic")
	encodingtests := []string{
		"Maßstäbe",
		"Maßstab",
		"Maßstäb",
		"das",
		"für",
	}
	for _, st := range encodingtests {
		suggestions := h.Suggest(st)
		if len(suggestions) == 0 {
			t.Errorf("Expected suggestions for %s", st)
		}
	}
}

func TestEncoding(t *testing.T) {
	h := Hunspell("testdics/de_DE.aff", "testdics/de_DE.dic")
	enc := h.Encoding()
	if enc != "UTF-8" {
		t.Errorf("Expected UTF-8 got %s", enc)
	}
}

func TestAddWord(t *testing.T) {
	h := Hunspell("testdics/en_US.aff", "testdics/en_US.dic")
	h.Add("customWord")
	encodingtests := []struct {
		in       string
		expected bool
	}{
		{"hello", true},
		{"customWord", true},
		{"customWordNotInDictionary", false},
	}
	for _, st := range encodingtests {
		if st.expected != h.Spell(st.in) {
			t.Errorf("Unexpected Spell result expected %v for %v", st.expected, st.in)
		}
	}
}

func TestCustomDict(t *testing.T) {
	h := Hunspell("testdics/en_US.aff", "testdics/en_US.dic")
	/*
		err := h.AddDict("testdics/foo.dic")
		if err == nil {
			t.Errorf("expected to get failure while loading custom dictionary")
		}
	*/
	err := h.AddDict("testdics/custom.dic")
	if err != nil {
		t.Errorf("failed to load custom dictionary")
	}
	encodingtests := []struct {
		in       string
		expected bool
	}{
		{"hello", true},
		{"CustomWord", false}, // TODO: should be true.
	}
	for _, st := range encodingtests {
		if st.expected != h.Spell(st.in) {
			t.Errorf("Unexpected Spell result expected %v for %v", st.expected, st.in)
		}
	}
}
