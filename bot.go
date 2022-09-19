package playlistdiscordbot

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/andersfylling/disgord"
	"github.com/andersfylling/disgord/std"
)

type Bot struct {
	Authentication string
	Guilds         *GuildRepository
	session        disgord.Session
	Logger         *log.Logger
	ready          bool
	voicestates    *voiceStateTracker
}

func (b *Bot) Connect() {
	b.session = disgord.New(disgord.Config{
		BotToken: b.Authentication,
		Intents:  disgord.AllIntents(),
	})

	b.session.Gateway().BotReady(func() {
		b.Logger.Println(b.session.BotAuthorizeURL(disgord.PermissionAll, []string{}))
		log.Println(b.session.BotAuthorizeURL(disgord.PermissionVoiceSpeak|disgord.PermissionVoiceConnect, []string{}))
		user, err := b.session.CurrentUser().Get()
		if err != nil {
			b.Logger.Fatal("Unable to fetch own user?")
			return
		}
		b.Logger.Printf("Logged in as %s#%s", user.Username, user.Discriminator)
		b.ready = true

		guilds, err := b.session.CurrentUser().GetGuilds(&disgord.GetCurrentUserGuilds{
			Limit: 50,
		})
		if err != nil {
			b.Logger.Fatalf("Cannot fetch guilds: %v\n", err)
		}

		for _, s := range guilds {
			b.Guilds.CreateOrUpdateGuild(s.ID, s)
			guild, _ := b.Guilds.GetGuildBySnowflake(s.ID)

			go func(s disgord.Snowflake) {

				for {
					// Refresh
					guild, _ := b.Guilds.GetGuildBySnowflake(s)
					if guild.CurrentPlaylist == "" {
						time.Sleep(10 * time.Second)
						continue
					}
					if guild.VoiceChannel == "" || guild.Connection == nil {
						time.Sleep(5 * time.Second)
						continue
					}
					playlist := guild.PlaylistRepository.GetPlaylistByName(guild.CurrentPlaylist)

					if len(playlist.EntryNames()) == 0 {
						time.Sleep(5 * time.Second)
						continue
					}

					filename := playlist.Current(int(guild.CurrentSeed), guild.CurrentIndex).Filename()

					f, err := os.Open(filename)
					if err != nil {
						time.Sleep(3 * time.Second)
						continue
					}

					b.session.SendMsg(disgord.ParseSnowflakeString(guild.PlayingChannel), "Currentply playing: "+filepath.Base(filename))

					guild.Connection.StartSpeaking()
					guild.Connection.SendDCA(f)

					guild.CurrentIndex++
					b.Guilds.UpdateGuild(guild)

					guild.Connection.StopSpeaking()
				}
			}(s.ID)
			if guild.VoiceChannel == "" {
				continue
			}
			err = nil
			guild.Connection, err = b.session.Guild(s.ID).VoiceChannel(disgord.ParseSnowflakeString(guild.VoiceChannel)).Connect(false, false)
			guild.connectedTo = guild.VoiceChannel

			if err != nil {
				b.Logger.Printf("Cannot connect to voicechannel %s for guild %s: %v\n", guild.VoiceChannel, s.String(), err)
			}
			os.Mkdir(guild.RootDirectory(), 0755)

		}
	})

	b.session.Gateway().GuildUpdate(func(s disgord.Session, h *disgord.GuildUpdate) {
		b.Guilds.CreateOrUpdateGuild(h.Guild.ID, h.Guild)
		guild, _ := b.Guilds.GetGuildBySnowflake(h.Guild.ID)
		os.Mkdir(guild.RootDirectory(), 0755)
	})

	b.voicestates = NewVoiceStateTracker()
	b.voicestates.Register(b.session.Gateway().WithContext(context.TODO()))

	b.Logger.Println("Connecting...")
	err := b.session.Gateway().StayConnectedUntilInterrupted()
	if err != nil {
		b.Logger.Fatalln(err)
	}
}

func (b *Bot) Disconnect() {
	b.session.Gateway().Disconnect()
}

func (b *Bot) ListenForMessages() {
	for b.session == nil {
		time.Sleep(1 * time.Second)
	}
	for prefix, executor := range commands {
		filter, err := std.NewMsgFilter(context.Background(), b.session)
		if err != nil {
			b.Logger.Fatalln(err)
		}
		filter.SetPrefix(string(prefix))
		b.session.Gateway().WithMiddleware(
			filter.HasPrefix,
			filter.NotByBot,
			filter.NotByWebhook,
			filter.StripPrefix,
		).MessageCreate(executor(b))
	}

}
