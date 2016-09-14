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
