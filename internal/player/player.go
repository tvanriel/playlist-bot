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
		Log:             p.Log.Named("player"),
		GuildRepository: p.GuildRepository,
		ProgressTracker: p.ProgressTracker,
	}
}

func (p *Player) Connect(ses *discordgo.Session, guildId string) error {
	conn, err := p.attemptConnect(ses, guildId)
        log := p.Log.With(zap.String("guildId", guildId))

        log.Info("Connecting to voicechannel")
	if err != nil {
                log.Warn("Cannot connect to voicechannel",
                        zap.Error(err),
                )
		return err
	}

	log.Info("Connected to voicechannel", zap.String("guild", guildId))

	go func() {
		for {
                        log.Info("Playing a new track")
                        log.Info("Selecting playlist")
			playlist, err := p.GuildRepository.GetPlaying(guildId)
			if err != nil {
				log.Warn("err while getting currently playing for guild",
					zap.Error(err),
				)
				time.Sleep(10 * time.Second)
				continue
			}
                        log = log.With(zap.Uint("playlist", playlist))
                        log.Info("Selected playlist")

			
                        log.Info("Selecting track")
                        track, err := p.PlaylistStore.RandomTrack(playlist)
			if err != nil {
				p.Log.Warn("Error while attempting to select track", zap.Error(err))
				time.Sleep(10 * time.Second)
				continue
			}
			t := &musicstore.Track{
				GuildID: guildId,
				Uuid:    track.Uuid,
			}
                        log = log.With(zap.String("track", track.Uuid))
                        log.Info("Selected track")

                        log.Info("Fetching DCA file")
			reader, err := p.MusicStore.GetDCA(t)
			if err != nil {
				log.Warn("err while fetching DCA",
					zap.Error(err),
				)
				time.Sleep(10 * time.Second)
				continue
			}

                        log.Info("Decoding DCA to OGG frames")
                        buf, err := loadSound(log, reader)
			if err != nil {
				log.Warn("err from dca reader",
					zap.Error(err),
				)
				time.Sleep(10 * time.Second)
				continue
			}

			err = conn.Speaking(true)
			if err != nil {
				continue
			}
			m := len(buf)
                        log.Info("Playing track to Discord", zap.Int("frames", m))

			for i := range buf {
				p.ProgressTracker.Report(&progresstracker.Progress{
					Current: i,
					Max:     m,
					Track:   track.Uuid,
				}, guildId)
				conn.OpusSend <- buf[i]
			}

                        log.Info("Finished playing track.")
				p.ProgressTracker.Report(&progresstracker.Progress{
					Current: 1,
					Max:     1,
					Track:   track.Uuid,
				}, guildId)

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
