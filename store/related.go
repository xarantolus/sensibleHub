package store

import (
	"sort"
	"strings"
	"xarantolus/sensibleHub/store/music"
)

// This doesn't really work

// GetRelatedSongs returns some related songs for a song
func (m *Manager) GetRelatedSongs(e music.Entry) (out []music.Entry) {
	if e.MusicData.Artist == "" {
		return
	}

	m.SongsLock.RLock()
	defer m.SongsLock.RUnlock()

	var items = make(map[int]music.Entry)

	for _, s := range m.Songs {
		if !strings.EqualFold(s.MusicData.Artist, e.MusicData.Artist) {
			continue
		}
		if s.ID == e.ID {
			continue
		}

		sc := score(strings.Fields(s.MusicData.Title), e.MusicData.Title, 3)
		sc += score(strings.Fields(s.MusicData.Album), e.MusicData.Album, 1)

		items[sc] = s
	}

	var nums []int
	for k := range items {
		nums = append(nums, k)
	}

	sort.Ints(nums)

	if len(nums) == 0 {
		return
	}

	// Return one
	if len(nums) < 2 {
		return []music.Entry{items[nums[0]]}
	}

	// Or two items
	return []music.Entry{items[nums[0]], items[nums[1]]}
}
