package playlistdiscordbot

type Member struct {
	Username string
	GuildId  string
}

type MemberRepository interface {
	GetMemberBySnowflake(guildId, memberId string) (Member, error)
	GetMemberById(guildId, memberId string) (Member, error)
	StoreMember(guildId string, member Member) error
}
