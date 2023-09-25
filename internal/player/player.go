package player

import (
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mitaka8/playlist-bot/internal/guildstore"
	"github.com/mitaka8/playlist-bot/internal/musicstore"
	"github.com/mitaka8/playlist-bot/internal/playliststore"
	"github.com/mitaka8/playlist-bot/internal/progresstracker"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type NewPlayerParams struct {
	fx.In

	PlaylistStore   *playliststore.PlaylistStore
	MusicStore      *musicstore.MusicStore
	GuildRepository *guildstore.GuildRepository
	Log             *zap.Logger
	ProgressTracker *progresstracker.ProgressTracker
}

type Player struct {
	PlaylistStore   *playliststore.PlaylistStore
	GuildRepository *guildstore.GuildRepository
	MusicStore      *musicstore.MusicStore
	Log             *zap.Logger
	ProgressTracker *progresstracker.ProgressTracker
}

type Progress struct {
	Uuid    string
	Current int
	Max     int
}

func NewPlayer(p NewPlayerParams) *Player {
	return &Player{
		PlaylistStore:   p.PlaylistStore,
		MusicStore:      p.MusicStore,
		Log:             p.Log,
		GuildRepository: p.GuildRepository,
		ProgressTracker: p.ProgressTracker,
	}
}

func (p *Player) Connect(ses *discordgo.Session, guildId string) error {
	conn, err := p.attemptConnect(ses, guildId)
	if err != nil {
		return err
	}

	p.Log.Info("Connected to voicechannel", zap.String("guild", guildId))

	go func() {
		for {

			playlist, err := p.GuildRepository.GetPlaying(guildId)
			if err != nil {
				p.Log.Warn("err while getting currently playing for guild",
					zap.String("guildId", guildId),
					zap.Error(err),
				)
				time.Sleep(10 * time.Second)
				continue
			}
			track, err := p.PlaylistStore.RandomTrack(playlist)
			if err != nil {
				p.Log.Warn("Error while attempting to select track", zap.Error(err), zap.String("guildId", guildId), zap.Uint("playlist", playlist))
				time.Sleep(10 * time.Second)
				continue
			}
			t := &musicstore.Track{
				GuildID: guildId,
				Uuid:    track.Uuid,
			}
			reader, err := p.MusicStore.GetDCA(t)
			if err != nil {
				p.Log.Warn("err while fetching track",
					zap.String("guildId", guildId),
					zap.Error(err),
					zap.Uint("playlistId", playlist),
				)
				time.Sleep(10 * time.Second)
				continue
			}

			buf, err := loadSound(reader)
			if err != nil {
	                        p.Log.Warn("err from dca reader",
					zap.String("guildId", guildId),
					zap.Error(err),
					zap.Uint("playlistId", playlist),
					zap.Uint("trackId", track.ID),
				)
				time.Sleep(10 * time.Second)
				continue
			}

			err = conn.Speaking(true)
			if err != nil {
				continue
			}
			m := len(buf)
			for i := range buf {
				p.ProgressTracker.Report(&progresstracker.Progress{
					Current: i,
					Max:     m,
					Track:   track.Uuid,
				}, guildId)
				conn.OpusSend <- buf[i]
			}
			_ = conn.Speaking(false)

		}
	}()
	return nil
}

var ErrNoSuchConnection = errors.New("no connection to guild voice channel")

func (p *Player) attemptConnect(ses *discordgo.Session, guildId string) (*discordgo.VoiceConnection, error) {

	conn, ok := ses.VoiceConnections[guildId]
	if ok {
		return conn, nil
	}

	channelId, err := p.GuildRepository.GetVoiceChannel(guildId)
	if err != nil {
		return nil, err
	}

	if channelId == "" {
		return nil, ErrNoSuchConnection
	}
	return ses.ChannelVoiceJoin(guildId, channelId, false, true)

}
