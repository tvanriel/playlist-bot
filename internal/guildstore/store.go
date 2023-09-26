package guildstore

import (
	"errors"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NewGuildStoreParams struct {
	fx.In

	Log *zap.Logger
	Db  *gorm.DB
}

type GuildRepository struct {
	Log *zap.Logger
	Db  *gorm.DB
}

type GuildModel struct {
	gorm.Model

	Snowflake    string
	Name         string
	Prefix       string
	Voicechannel string
	Icon         string

	CurrentlyPlayingID *uint
}

func NewGuildRepository(p NewGuildStoreParams) *GuildRepository {
	return &GuildRepository{
		Db:  p.Db,
		Log: p.Log.Named("guildstore"),
	}
}

func MigrateGuildRepository(r *GuildRepository) error {
	return r.Db.AutoMigrate(
		&GuildModel{},
	)
}

const DEFAULT_PREFIX = "%"

func (r *GuildRepository) LoadGuild(guildId string, name string, icon string) error {
	var model GuildModel
	tx := r.Db.First(&model, "snowflake = ?", guildId)

	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return tryQuery(r.Db.Create(&GuildModel{
			Snowflake: guildId,
			Prefix:    DEFAULT_PREFIX,
			Name:      name,
			Icon:      icon,
		}))
	}
	model.Name = name
	return tryQuery(r.Db.Save(&model))

}
func (r *GuildRepository) GetVoiceChannels() (map[string]string, error) {

	m := make(map[string]string)

	var models []GuildModel
	tx := r.Db.Find(&models, "voicechannel != ''")
	if tx.Error != nil {
		return m, tx.Error
	}

	for i := range models {
		m[models[i].Snowflake] = models[i].Voicechannel
	}

	return m, nil

}

func (r *GuildRepository) UpdatePrefix(guildId, prefix string) error {
	return tryQuery(
		r.Db.Model(&GuildModel{}).
			Where("snowflake", guildId).
			Update("prefix", prefix),
	)
}

func (r *GuildRepository) JoinVoiceChannel(guildId, channelId string) error {
	return tryQuery(
		r.Db.Model(&GuildModel{}).
			Where("snowflake = ?", guildId).
			UpdateColumn("voicechannel", channelId),
	)

}

func (r *GuildRepository) GetVoiceChannel(guildId string) (string, error) {
	var channelId string

	tx := r.Db.Model(&GuildModel{}).Where("snowflake = ?", guildId).
		Pluck("voicechannel", &channelId)

	return channelId, tx.Error
}

// Fetch the map of guildId => guildname
func (r *GuildRepository) GetGuilds() ([]GuildModel, error) {

	var models []GuildModel
	tx := r.Db.Find(&models, "voicechannel != ''")
	if tx.Error != nil {
		return models, tx.Error
	}

	return models, tx.Error
}

func (r *GuildRepository) GetPrefix(snowflake string) (string, error) {

	model := new(GuildModel)
	err := tryQuery(
		r.Db.Where("snowflake = ?", snowflake).Find(model),
	)
	if err != nil {
		return "", err
	}
	return model.Prefix, nil
}

func (r *GuildRepository) SetPlaying(snowflake string, playlistId uint) error {
	return tryQuery(r.Db.Model(&GuildModel{}).Where("snowflake = ?", snowflake).UpdateColumn("currently_playing_id", playlistId))
}

func (r *GuildRepository) GetPlaying(snowflake string) (uint, error) {
	var currentlyPlaying uint
	err := tryQuery(r.Db.Model(&GuildModel{}).Where("snowflake = ?", snowflake).Pluck("currently_playing_id", &currentlyPlaying))
	return currentlyPlaying, err

}
