package store

import (
	"sort"
	"strings"
	"unicode"
	"xarantolus/sensibleHub/store/music"
)

// Search offers search functionality
func (m *Manager) Search(query string) (list []music.Entry) {
	qs := splitString(strings.ToUpper(strings.TrimSpace(query)))
	e := m.AllEntries()

	type result struct {
		score int
		item  music.Entry
	}

	var res []result

	for _, item := range e {
		var sc int
		sc += score(qs, item.MusicData.Title, 5)
		sc += score(qs, item.MusicData.Artist, 4)

		// Only use album if it's not the same as other fields
		if doublePrefix(item.MusicData.Album, item.MusicData.Title) ||
			doublePrefix(item.MusicData.Album, item.MusicData.Artist) {
			sc += score(qs, item.MusicData.Album, 1)
		} else {
			sc += score(qs, item.MusicData.Album, 3)
		}

		if sc > 0 {
			res = append(res, result{sc, item})
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].score > res[j].score
	})

	for _, r := range res {
		list = append(list, r.item)
	}

	return
}

func score(query []string, title string, multiplier int) (out int) {
	tu := strings.ToUpper(title)
	tf := splitString(tu)
	if len(tf) == 0 {
		return
	}

	for _, q := range query {
		if strings.HasPrefix(tu, q) {
			out += 15
		} else {
			out -= 2
		}
		if strings.Contains(tu, q) {
			out += 10
		} else {
			out--
		}
		if contains(tf, q) {
			out += 5
		} else {
			out--
		}
	}

	out *= multiplier
	return
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func doublePrefix(s1, s2 string) bool {
	if strings.HasPrefix(s1, s2) {
		return true
	}
	return strings.HasPrefix(s2, s1)
}

// splitString splits a string at its spaces
// Original version / source: https://play.golang.org/p/ztqfYiPSlv / https://groups.google.com/d/msg/golang-nuts/pNwqLyfl2co/APaZSSvQUAAJ
func splitString(s string) (words []string) {
	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return unicode.IsSpace(c)
		}
	}

	// Now we loop over single words and remove quotes
	for _, word := range strings.FieldsFunc(s, f) {
		words = append(words, strings.TrimFunc(word, func(c rune) bool {
			return unicode.IsSpace(c) || unicode.In(c, unicode.Quotation_Mark)
		}))
	}
	return
}
