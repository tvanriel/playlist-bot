package playliststore

import (
	"errors"

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

func (p *PlaylistStore) Append(guildId, playlistName, id string) error {
	playlist, err := p.FindByGuildAndName(guildId, playlistName)
	if err != nil {
		return err
	}

	t := &TrackModel{
		PlaylistID: playlist.ID,
		Uuid:       id,
	}
	return tryQuery(p.mysql.Save(t))

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
