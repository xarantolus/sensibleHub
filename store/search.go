package store

import (
	"sort"
	"strings"
	"xarantolus/sensiblehub/store/music"
)

// Search offers search functionality
func (m *Manager) Search(query string) (list []music.Entry) {
	qs := strings.Fields(strings.ToUpper(query))
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
		if item.MusicData.Album != item.MusicData.Title && item.MusicData.Album != item.MusicData.Artist {
			sc += score(qs, item.MusicData.Album, 3)
		}

		if sc != 0 {
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
	tf := strings.Fields(tu)
	if len(tf) == 0 {
		return
	}

	for _, q := range query {
		if strings.HasPrefix(tu, q) {
			out += 10
		}
		if contains(tf, q) {
			out += 3
			continue
		}

		if strings.Contains(tu, q) {
			out += 2
			continue
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
