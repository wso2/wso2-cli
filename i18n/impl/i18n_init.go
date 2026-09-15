package impl

var T TranslateFunc

func init() {
	t, err := GetTranslationFunc()

	if err != nil {
		panic(err)
	}

	T = t
}
