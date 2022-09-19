package playlistdiscordbot

import (
	"errors"
	"os"
	"strings"

	"github.com/andersfylling/disgord"
	"gorm.io/gorm"
)

type Guild struct {
	gorm.Model
	Snowflake          string
	Name               string
	VoiceChannel       string
	Root               string                  `gorm:"-;all"`
	PlaylistRepository *PlaylistRepository     `gorm:"-;all"`
	Connection         disgord.VoiceConnection `gorm:"-;all"`
	connectedTo        string                  `gorm:"-;all"`

	PlayingChannel string

	CurrentSeed     int64
	CurrentIndex    int
	CurrentPlaylist string
}

func (g *Guild) RootDirectory() string {
	return strings.Join([]string{g.Root, g.Snowflake}, string(os.PathSeparator))
}

type GuildRepository struct {
	DB              *gorm.DB
	SoundRoot       string
	Guilds          map[string]*Guild
	CurrentPlaylist string
}

func (g *GuildRepository) GetGuildBySnowflake(snowflake disgord.Snowflake) (*Guild, error) {
	if found, ok := g.Guilds[snowflake.String()]; ok {
		return found, nil
	}
	guild := &Guild{Snowflake: snowflake.String()}
	g.DB.Where(guild, "snowflake").First(guild)
	guild.Root = g.SoundRoot

	guild.PlaylistRepository = &PlaylistRepository{
		Root: guild.RootDirectory(),
	}

	g.Guilds[snowflake.String()] = guild
	return guild, nil
}

func (g *GuildRepository) Exists(snowflake disgord.Snowflake) bool {
	var count int64
	g.DB.Model(&Guild{}).Where("Snowflake", snowflake.String()).Count(&count)
	return count > 0

}
func (g *GuildRepository) CreateOrUpdateGuild(snowflake disgord.Snowflake, discordGuild *disgord.Guild) {
	if g.Exists(snowflake) {
		guild, _ := g.GetGuildBySnowflake(snowflake)
		guild.Name = discordGuild.Name
		guild.Root = g.SoundRoot
		g.DB.Save(guild)
		g.Guilds[snowflake.String()] = guild

		os.Mkdir(guild.RootDirectory(), 0755)
		return
	}
	guild := Guild{Snowflake: snowflake.String(), Root: g.SoundRoot, Name: discordGuild.Name}

	g.Guilds[snowflake.String()] = &guild
	g.DB.Create(&guild)

	os.Mkdir(guild.RootDirectory(), 0755)
}

func (g *GuildRepository) UpdateGuild(guild *Guild) {

	old := g.Guilds[guild.Snowflake]
	old.Snowflake = guild.Snowflake
	old.Connection = guild.Connection
	old.Name = guild.Name
	old.Root = guild.Root
	old.VoiceChannel = guild.VoiceChannel
	old.connectedTo = guild.connectedTo
	old.CurrentPlaylist = guild.CurrentPlaylist
	old.CurrentIndex = guild.CurrentIndex
	old.CurrentSeed = guild.CurrentSeed
	g.DB.Save(guild)
}

func (g *Guild) AttemptVoiceConnect(s disgord.Session) error {
	if g.VoiceChannel == "" {
		return errors.New("voice channel is nil")
	}
	guildSnowflake := disgord.ParseSnowflakeString(g.Snowflake)
	voiceChannelSnowflake := disgord.ParseSnowflakeString(g.VoiceChannel)

	var err error
	if g.Connection == nil {
		g.Connection, err = s.Guild(guildSnowflake).
			VoiceChannel(voiceChannelSnowflake).
			Connect(false, false)
		if err != nil {
			g.connectedTo = g.VoiceChannel
		}

		return err
	}

	if g.connectedTo != g.VoiceChannel {

		err = g.Connection.Close()
		if err != nil {
			return err
		}
		g.connectedTo = ""
		g.Connection = nil

		g.Connection, err = s.Guild(guildSnowflake).
			VoiceChannel(voiceChannelSnowflake).
			Connect(false, false)
		if err != nil {
			g.connectedTo = g.VoiceChannel
		}
	}

	return err
}
