package playlistdiscordbot

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/andersfylling/disgord"
	"github.com/kkdai/youtube/v2"
)

type CommandExecutor func(*Bot) disgord.HandlerMessageCreate
type CommandName string

var commands = map[CommandName]CommandExecutor{
	"%list":    showSongsInPlaylistCommand,
	"%join":    joinCommand,
	"%save ":   saveCommand,
	"%shuffle": shuffleCommand,
	"%playing": playingCommand,
	"%add ":    createPlaylistCommand,
	"%play ":   playCommand,
	"%reset":   resetCommand,
}

func formatListPlaylistsMessage(names []string) string {
	return "`" + strings.Join(names, "`,\n`") + "`"
}
func showSongsInPlaylistCommand(b *Bot) disgord.HandlerMessageCreate {
	return func(s disgord.Session, h *disgord.MessageCreate) {
		name := getArgument(h.Message.Content, 0)

		if name == "" {
			var guild *Guild
			var err error
			if guild, err = b.Guilds.GetGuildBySnowflake(h.Message.GuildID); err != nil {
				h.Message.Reply(context.Background(), s, "Cannot find guild")
				return
			}
			playlists := guild.PlaylistRepository.GetPlaylists()
			if len(playlists) == 0 {
				h.Message.Reply(context.Background(), s, "You haven't created any playlists yet.  Use `%add <name>` to make a new playlist.")
				return
			}

			h.Message.Reply(context.Background(), s, formatListPlaylistsMessage(guild.PlaylistRepository.GetPlaylists()))

			return
		}

		guild, _ := b.Guilds.GetGuildBySnowflake(h.Message.GuildID)
		if !guild.PlaylistRepository.Exists(name) {
			h.Message.Reply(context.Background(), s, "Playlist does not exist.")
			return
		}
		playlist := guild.PlaylistRepository.GetPlaylistByName(name)
		songs := playlist.EntryNames()
		h.Message.Reply(context.Background(), s, formatListPlaylistsMessage(songs))
	}
}

func joinCommand(b *Bot) disgord.HandlerMessageCreate {
	return func(s disgord.Session, h *disgord.MessageCreate) {
		b.Logger.Println("test")
		for _, vs := range b.voicestates.States(h.Message.GuildID) {
			if vs.UserID != h.Message.Member.UserID {
				continue
			}

			guild, err := b.Guilds.GetGuildBySnowflake(h.Message.GuildID)
			if err != nil {
				h.Message.Reply(context.Background(), s, "Unknown guild")
				return
			}
			guild.VoiceChannel = vs.ChannelID.String()
			guild.PlayingChannel = h.Message.ChannelID.String()

			b.Guilds.UpdateGuild(guild)

			go func() {
				if err = guild.AttemptVoiceConnect(s); err != nil {
					h.Message.Reply(context.Background(), s, fmt.Sprintf("Cannot connect to voice channel: %v\n", err))
				}
			}()

			return
		}

		h.Message.Reply(context.Background(), s, "Join a voicechannel first.")
	}
}
func saveCommand(b *Bot) disgord.HandlerMessageCreate {
	return func(s disgord.Session, h *disgord.MessageCreate) {
		playlistName := getArgument(h.Message.Content, 0)
		link := getArgument(h.Message.Content, 1)

		if playlistName == "" {
			h.Message.Reply(context.Background(), s, "Enter a playlist name")
			return
		}

		if isYouTubeLink(link) {
			h.Message.Reply(context.Background(), s, "Downloading youtube video...")
			saveYoutubeLink(b, s, h, playlistName, link)
			return
		} else if isYouTubePlaylist(link) {
			h.Message.Reply(context.Background(), s, "Downloading all videos in this playlist.")
			downloadPlaylist(b, s, h, playlistName, link)
			return
		}
		h.Message.Reply(context.Background(), s, "Downloading attachment...")
		saveAttachmentCommand(b, s, h, playlistName, link)
	}
}

func playCommand(b *Bot) disgord.HandlerMessageCreate {
	return func(s disgord.Session, h *disgord.MessageCreate) {

		guild, _ := b.Guilds.GetGuildBySnowflake(h.Message.GuildID)
		rand.Seed(time.Now().UnixNano())

		if !guild.PlaylistRepository.Exists(getArgument(h.Message.Content, 0)) {
			h.Message.Reply(context.Background(), s, "Playlist does not exist")
			return
		}

		guild.CurrentPlaylist = getArgument(h.Message.Content, 0)
		rand.Seed(time.Now().UnixNano())
		guild.CurrentSeed = int64(rand.Int())
		guild.CurrentIndex = 0
		b.Guilds.UpdateGuild(guild)
		h.Message.Reply(context.Background(), s, "Now playing "+h.Message.Content)
	}
}
func shuffleCommand(b *Bot) disgord.HandlerMessageCreate {
	return func(s disgord.Session, h *disgord.MessageCreate) {

		guild, _ := b.Guilds.GetGuildBySnowflake(h.Message.GuildID)
		rand.Seed(time.Now().UnixNano())
		guild.CurrentSeed = int64(rand.Int())
		b.Guilds.UpdateGuild(guild)
	}
}
func playingCommand(b *Bot) disgord.HandlerMessageCreate {
	return func(s disgord.Session, h *disgord.MessageCreate) {

	}
}

func resetCommand(b *Bot) disgord.HandlerMessageCreate {
	return func(s disgord.Session, h *disgord.MessageCreate) {
		h.Message.Reply(context.Background(), s, "I'll be back https://i.ytimg.com/vi/_aA8Le6Q_L0/hqdefault.jpg")
		os.Exit(0)
	}
}
func createPlaylistCommand(b *Bot) disgord.HandlerMessageCreate {
	return func(s disgord.Session, h *disgord.MessageCreate) {
		guild, _ := b.Guilds.GetGuildBySnowflake(h.Message.GuildID)
		name := getArgument(h.Message.Content, 0)

		if guild.PlaylistRepository.Exists(name) {
			h.Message.Reply(context.Background(), s, "Playlist already exists")
			return
		}

		err := guild.PlaylistRepository.CreatePlaylist(name)
		if err != nil {
			h.Message.Reply(context.Background(), s, err)
			return
		}

		h.Message.Reply(context.Background(), s, "Playlist created.  Use `%save "+name+"` to save a song in the playlist.\n A second argument to this function will be interpreted as a YouTube Link.  Leaving out the second argument will download an attachment into the playlist.")
	}
}
func getArgument(command string, n int) string {
	split := strings.Split(command,
		" ")
	if len(split) < n {
		return ""
	}
	return split[n]
}

func isYouTubeLink(possiblyLink string) bool {
	r := regexp.MustCompile(`https://((www\.)?youtube\.com/watch\?v=|youtube\.be/)[a-zA-Z0-9\-_]+`)
	return r.MatchString(possiblyLink)
}

func isYouTubePlaylist(possiblyLink string) bool {
	r := regexp.MustCompile(`https://((www\.)?youtube\.com/playlist\?list=)[a-zA-Z0-9\-_]+`)
	return r.MatchString(possiblyLink)
}
func saveAttachmentCommand(b *Bot, s disgord.Session, h *disgord.MessageCreate, playlistName, soundName string) {
	attachments := h.Message.Attachments
	if len(attachments) == 0 {
		h.Message.Reply(context.Background(), s, "Add your sound as the first attachment please.")
		return
	}
	attachment := attachments[0]

	guild, err := b.Guilds.GetGuildBySnowflake(h.Message.GuildID)
	if err != nil {
		h.Message.Reply(context.Background(), s, "You're in an unknown guild")
		return
	}
	sound := guild.PlaylistRepository.GetPlaylistByName(playlistName).MakeEntry(soundName)

	go func() {
		DownloadAndConvert(attachment.URL, sound.Filename(), func(err error) {
			h.Message.Reply(context.Background(), s, err)
		}, func() {
			h.Message.Reply(context.Background(), s, "Added "+soundName+" to "+playlistName)
		})
	}()

}

func downloadPlaylist(b *Bot, s disgord.Session, h *disgord.MessageCreate, playlistName, link string) {
	client := youtube.Client{}
	guild, err := b.Guilds.GetGuildBySnowflake(h.Message.GuildID)
	if err != nil {
		h.Message.Reply(context.Background(), s, "Unknown playlist")
		return
	}
	playlist, err := client.GetPlaylist(link)
	if err != nil {
		h.Message.Reply(context.Background(), s, "Cannot fetch playlist details")
		return
	}
	go func() {

		for _, entry := range playlist.Videos {
			video, err := client.VideoFromPlaylistEntry(entry)
			if err != nil {
				h.Message.Reply(context.Background(), s, "Cannot fetch playlist details")
				continue
			}

			soundName := sanitizeYouTubeTitle(video)
			sound := guild.PlaylistRepository.GetPlaylistByName(playlistName).MakeEntry(soundName)

			go YoutubeDownloadAndConvert(video.ID, sound.Filename(), func(err error) {
				h.Message.Reply(context.Background(), s, err)
			}, func() {
				h.Message.Reply(context.Background(), s, "Added "+soundName+" to "+playlistName)
			})

		}
	}()

}

func saveYoutubeLink(b *Bot, s disgord.Session, h *disgord.MessageCreate, playlistName, link string) {
	guild, err := b.Guilds.GetGuildBySnowflake(h.Message.GuildID)
	if err != nil {
		h.Message.Reply(context.Background(), s, "You're in an unknown guild")
		return
	}

	go func() {
		info, err := getVideoInfoFromYoutube(link)
		if err != nil {

			h.Message.Reply(context.Background(), s, err)
			return
		}
		soundName := sanitizeYouTubeTitle(info)
		sound := guild.PlaylistRepository.GetPlaylistByName(playlistName).MakeEntry(soundName)

		go YoutubeDownloadAndConvert(link, sound.Filename(), func(err error) {
			h.Message.Reply(context.Background(), s, err)
		}, func() {
			h.Message.Reply(context.Background(), s, "Added "+soundName+" to "+playlistName)
		})
	}()

}

func sanitizeYouTubeTitle(info *youtube.Video) string {
	str := info.Author + " - " + info.Title
	removeChars := []string{" ", "/", "<", ">", "\"", "|", "\\", "*", "?", ":"}
	for _, char := range removeChars {
		str = strings.ReplaceAll(str, char, "_")
	}
	return str
}

func getVideoInfoFromYoutube(link string) (*youtube.Video, error) {
	client := youtube.Client{}
	return client.GetVideo(link)
}
