package playliststore

import (
	"errors"
	"strconv"
	"strings"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type PlaylistModel struct {
	gorm.Model

	GuildID string
	Name    string
}

type TrackModel struct {
	gorm.Model

	PlaylistID uint
	Playlist   PlaylistModel

	Uuid string

        ArtistName string
        TrackName string
}

type PlaylistStore struct {
	mysql *gorm.DB
}

type NewMySQLPlaylistStoreParams struct {
	fx.In

	Mysql *gorm.DB
}

func NewMySQLPlaylistStore(p NewMySQLPlaylistStoreParams) *PlaylistStore {
	return &PlaylistStore{
		mysql: p.Mysql,
	}
}

func MigratePlaylistStore(p *PlaylistStore) error {
	return p.mysql.AutoMigrate(
		&PlaylistModel{},
		&TrackModel{},
	)
}

func (p *PlaylistStore) AddPlaylist(guildId string, name string) error {
	if name == "" {
		return errors.New("cannot create empty playlist!")
	}
	return tryQuery(
		p.mysql.Create(&PlaylistModel{
			GuildID: guildId,
			Name:    name,
		}),
	)
}
func (p *PlaylistStore) PlaylistExists(guildId string, name string) (bool, error) {
	playlist := &PlaylistModel{}

	tx := p.mysql.Where("guild_id = ? AND name = ?", guildId, name).Limit(1).Find(playlist)

	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, tx.Error
}

func (p *PlaylistStore) ListPlaylists(guildId string) ([]string, error) {

	playlists := make([]PlaylistModel, 0)

	tx := p.mysql.Where("guild_id = ?", guildId).Find(&playlists)
	if tx.Error != nil {
		return nil, tx.Error
	}

	var names []string

	for i := range playlists {
		names = append(names, playlists[i].Name)
	}
	return names, nil
}

func (p *PlaylistStore) FindByGuildAndName(guildId string, name string) (*PlaylistModel, error) {
	playlist := new(PlaylistModel)
	err := tryQuery(p.mysql.Where("guild_id = ?", guildId).Where("name = ?", name).Find(playlist))
	return playlist, err
}

func (p *PlaylistStore) Append(guildId, playlistName, id, trackName, artistName string) error {
	playlist, err := p.FindByGuildAndName(guildId, playlistName)
	if err != nil {
		return err
	}

	t := &TrackModel{
		PlaylistID: playlist.ID,
		Uuid:       id,
                TrackName:  trackName,
                ArtistName: artistName,
	}
	return tryQuery(p.mysql.Save(t))

}

func (p *PlaylistStore) TrackIds(playlistId uint) (trackIds []string, err error) {
	err = tryQuery(p.mysql.
		Model(&TrackModel{}).
		Order("RAND()").
		Where("playlist_id = ?", playlistId).
		Pluck("uuid", &trackIds),
	)
	return
}
func (p *PlaylistStore) RandomTrack(playlistId uint) (*TrackModel, error) {
	t := &TrackModel{}
	err := tryQuery(p.mysql.Order("RAND()").Where("playlist_id = ?", playlistId).Limit(1).Find(t))
	if err != nil {
		return t, err
	}
	if t.Uuid == "" {
		return t, errors.New("no random track available")
	}
	return t, err
}

func (p *PlaylistStore) FindByID(id uint) (*PlaylistModel, error) {

	playlist := new(PlaylistModel)
	err := tryQuery(p.mysql.Where("id = ?", id).Find(playlist))
	return playlist, err
}

func (p *PlaylistStore) FindTrack(id uint) (*TrackModel, error) {
        t := new(TrackModel)
        t.ID = id
        err := tryQuery(p.mysql.Find(t))
        return t, err
}

func (p *PlaylistStore) Search(guildId, term string) ([]TrackModel, error) {
        var playlists []uint
        err := tryQuery(p.mysql.Model(&PlaylistModel{}).Where("guild_id = ?", guildId).Pluck("id", &playlists))
        if err != nil {
                return nil, err 
        }
        var ts []TrackModel
        err = tryQuery(p.mysql.Model(&TrackModel{}).
                Where("playlist_id IN ?", playlists).
                Where("artist_name LIKE ?", like(term)).
                Or("track_name LIKE ?", like(term)).
                Find(&ts),
        )
        return ts, err
}

func like(term string) string {
        return strings.Join([]string{"%", term, "%"}, "")
}

func (t TrackModel) String() string {
        return strings.Join([]string{
                strconv.FormatUint(uint64(t.ID), 10),
                ") **",
                t.ArtistName,
                "** - **",
                t.TrackName,
                "**",
        }, "")
}
