package store

import "testing"

func Test_cleanBrackets(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{"Title (feat. B)", "Title"},
		{" Title (feat. B)", "Title"},
		{"Title (feat. B) ", "Title"},

		{" Title (feat. B) ", "Title"},
		{"Title ) (", "Title ) ("},
		{"Title ()     )(", "Title )("},
	}
	for _, tt := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			if got := cleanBrackets(tt.arg); got != tt.want {
				t.Errorf("cleanBrackets() = %v, want %v", got, tt.want)
			}
		})
	}
}
