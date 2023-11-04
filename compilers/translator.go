package compilers

type Translator interface {
	Translate(text string) (string, error)
}
