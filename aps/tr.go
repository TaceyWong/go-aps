// transelation of tips-info in go-aps package

package aps

import "strings"

type Word struct {
	Ori   string
	MapTo map[string]string
}

var words = []*Word{
	{Ori: "China", MapTo: map[string]string{"ZH": "中国"}},
}

var Words = make(map[string]*Word)

func init() {
	for _, w := range words {
		Words[w.Ori] = w
	}
}

func (w Word) String() string {
	if _, ok := w.MapTo[strings.ToUpper(LANG)]; ok {
		return w.MapTo[strings.ToUpper(LANG)]
	}
	return w.Ori
}

func Tr(text string) string {
	if _, ok := Words[text]; ok && strings.ToUpper(LANG) != "EN" {
		return Words[text].MapTo[strings.ToUpper(LANG)]
	} else {
		return text
	}
}
