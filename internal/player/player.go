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

type session struct {
	index      int
	tracks     []string
	playlistId uint
}

type Player struct {
	PlaylistStore   *playliststore.PlaylistStore
	GuildRepository *guildstore.GuildRepository
	MusicStore      *musicstore.MusicStore
	Log             *zap.Logger
	ProgressTracker *progresstracker.ProgressTracker

	playlist map[string]*session
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
                playlist: make(map[string]*session),
	}
}

func (p *Player) Connect(ses *discordgo.Session, guildId string) error {
	log := p.Log.With(zap.String("guildId", guildId))

	log.Info("Connecting to voicechannel")
	conn, err := p.attemptConnect(ses, guildId)
	if err != nil {
		log.Warn("Cannot connect to voicechannel",
			zap.Error(err),
		)
		return err
	}

	go func() {
		for {
			log = p.Log.With(zap.String("guild", guildId))
			track, err := p.Next(guildId)
			if err != nil {
				log.Info("Failed to retrieve next track for guild")
				time.Sleep(10 * time.Second)
                                continue
			}
                        if track == nil {
                                log.Info("Track is nil")
                                time.Sleep(10 *time.Second)
                                continue
                        }

			log = log.With(zap.String("track", track.Uuid))
			log.Info("Selected track")

			log.Info("Fetching DCA file")
			reader, err := p.MusicStore.GetDCA(track)
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
func (p *Player) Next(guildId string) (*musicstore.Track, error) {
	log := p.Log.With(zap.String("guildId", guildId))

	// Fetch what is currently playing on this guild.
	playlist, err := p.GuildRepository.GetPlaying(guildId)
	if err != nil {
		log.Warn("err while getting currently playing for guild",
			zap.Error(err),
		)
		return nil, err
	}
	log = log.With(zap.Uint("playlist", playlist)) 

        // get the session
	ses, ok := p.playlist[guildId]
	if !ok {
                // Oh no, there is no session, create a new one.
		err = p.NewSession(guildId, playlist)
		if err != nil {
			log.Warn("cannot create new play session for guild",
				zap.Error(err),
			)
			return nil, err
		}
	        ses = p.playlist[guildId]

	}

        // Session playlist is not equal to whatever we want to play, create a new session.
	if playlist != p.playlist[guildId].playlistId {
		log.Info("playlist changed, updating session playlist")
		err = p.NewSession(guildId, playlist)
		if err != nil {
			log.Warn("cannot create new play session for guild",
				zap.Error(err),
			)
			return nil, err
		}
	}


        // Is there anything in the tracks entry?
        if len(ses.tracks) == 0 {
                return nil, nil
        }

        // We reached the end of the session. 
	if ses.index >= len(ses.tracks) {
		ses.index = 0
	}
	track := ses.tracks[ses.index]
	ses.index++
	if err != nil {
		p.Log.Warn("Error while attempting to select track", zap.Error(err))
	}

	return &musicstore.Track{
		GuildID: guildId,
		Uuid:    track,
	}, nil

}
func (p *Player) NewSession(guildId string, playlist uint) error {
	tracks, err := p.PlaylistStore.TrackIds(playlist)
	if err != nil {
		p.Log.Warn("cannot get trackIds from playlist",
                        zap.String("guildId", guildId),
                        zap.Uint("playlist", playlist),
			zap.Error(err),
		)
		return err
	}
	p.playlist[guildId] = &session{
		index:      0,
		tracks:     tracks,
		playlistId: playlist,
	}

        p.Log.Info("new session", 
                        zap.String("guildId", guildId),
                        zap.Uint("playlist", playlist),
                        zap.Int("len(tracks)", len(tracks)),
                        zap.Strings("tracks", tracks),


        )
	return nil
}
