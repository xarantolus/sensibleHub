package music

// Entry represents the entry for an audio file
type Entry struct {
	// ID is the randomly-generated, unique ID for this file
	ID string

	// SourceURL is either the source input url where the audio file was downloaded from or
	// the `webpage_url` field from the youtube-dl info file. Both point to the website where one might listen to the song
	SourceURL string

	// SyncSettings defines if this Entrys' MP3 file should be synced
	SyncSettings SyncSettings

	// MetaFile is the `.info.json` metadata file from youtube-dl
	MetaFile MetaFile

	// FileData describes data about the original file that was downloaded
	FileData

	// AudioSettings stores settings such as audio start and end time
	AudioSettings AudioSettings

	// MusicData describes music metadata that is typically embedded into music files
	MusicData MusicData

	// PicureData describes the picture file (cover) that should be embedded into the file
	PicureData PictureData
}

type SyncSettings struct {
	Should bool
}

type MetaFile struct {
	Filename string
}

type FileData struct {
	Filename string
	Size     int64
}

type AudioSettings struct {
	// Start is the start time of the song in the given audio. If not set, it is < 0
	Start int
	// Start is the end time of the song in the given audio. If not set, it is < 0
	End int
}

type MusicData struct {
	Title  string
	Artist string
	Album  string
	Year   int
}

type PictureData struct {
	MimeType string
	Filename string
}
