package translators

type Translator interface {
	Translate(text string) (string, error)
}
