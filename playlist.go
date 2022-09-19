package playlistdiscordbot

import (
	"os"
	"path/filepath"
	"strings"

	"math/rand"
)

type PlaylistEntry struct {
	Name              string
	PlaylistDirectory string `gorm:"-;all"`
}

func (e *PlaylistEntry) GetName() string {
	return e.Name
}

func (e *PlaylistEntry) Filename() string {
	var sb strings.Builder
	sb.WriteString(e.GetName())
	sb.WriteString(".dca")
	return strings.Join([]string{e.PlaylistDirectory, sb.String()}, string(os.PathSeparator))
}

type Playlist struct {
	Name          string
	RootDirectory string
}

func (p *Playlist) EntryNames() []string {
	files, err := os.ReadDir(p.Directory())
	if err != nil {
		return []string{}
	}
	var names []string
	for _, file := range files {
		names = append(names, file.Name()[0:len(file.Name())-4])
	}
	return names
}

func (p *Playlist) Directory() string {
	return strings.Join([]string{p.RootDirectory, p.GetName()}, string(os.PathSeparator))
}

func (p *Playlist) Shuffle(seed int) []*PlaylistEntry {
	rand.Seed(int64(seed))
	entries := p.Entries()
	for i := range entries {
		j := rand.Intn(i + 1)
		entries[j], entries[i] = entries[i], entries[j]
	}
	return entries
}
func (p *Playlist) Entries() []*PlaylistEntry {
	var entries []*PlaylistEntry

	files, err := os.ReadDir(p.Directory())
	if err != nil {
		return entries
	}

	for _, file := range files {
		filename := file.Name()
		entries = append(entries, &PlaylistEntry{
			Name:              filename[0 : len(filename)-4],
			PlaylistDirectory: p.Directory(),
		})
	}
	return entries
}

func (p *Playlist) MakeEntry(name string) *PlaylistEntry {
	return &PlaylistEntry{
		Name:              name,
		PlaylistDirectory: p.Directory(),
	}
}

func (p *Playlist) GetName() string {
	return p.Name
}

func (p *Playlist) Current(seed, index int) *PlaylistEntry {
	return p.Shuffle(seed)[index%len(p.EntryNames())]
}

type PlaylistRepository struct {
	Playlists []*Playlist
	Root      string
}

func (p *PlaylistRepository) GetPlaylistByName(name string) *Playlist {
	playlist := &Playlist{RootDirectory: p.Root, Name: name}
	return playlist
}

func (p *PlaylistRepository) Exists(name string) bool {

	path := strings.Join([]string{p.Root, name}, string(os.PathSeparator))
	_, err := os.Open(path)
	return err == nil || os.IsExist(err)
}

func (p *PlaylistRepository) GetPlaylists() []string {
	entries, err := os.ReadDir(p.Root)
	if err != nil {
		return []string{}
	}
	var names []string
	for _, entry := range entries {
		names = append(names, filepath.Base(entry.Name()))
	}
	return names
}

func (p *PlaylistRepository) CreatePlaylist(name string) error {

	path := strings.Join([]string{p.Root, name}, string(os.PathSeparator))
	return os.Mkdir(path, 0755)
}
