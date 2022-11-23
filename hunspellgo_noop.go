//go:build !cgo

package hunspellgo

type Hunhandle struct {
}

func Hunspell(affpath string, dpath string) *Hunhandle {
	return &Hunhandle{}
}

// Add adds a word to the dictionary.
func (handle *Hunhandle) Add(word string) error {
	return nil
}

// AddDict adds a custom dictionary.
func (handle *Hunhandle) AddDict(path string) error {
	return nil
}

func (handle *Hunhandle) Suggest(word string) []string {
	return nil
}

func (handle *Hunhandle) Stem(word string) []string {
	return nil
}

func (handle *Hunhandle) Spell(word string) bool {
	return true
}

func (handle *Hunhandle) Encoding() string {
	return ""
}
