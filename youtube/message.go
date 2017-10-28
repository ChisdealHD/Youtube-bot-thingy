package youtube

import "time"

type messageResponse struct {
	Kind string
	Etag string
	NextPageToken string
	PollingIntervalMillis int
	PageInfo messagePageInfo
	Items []Message
}
type messagePageInfo struct {
	TotalResults int
	ResultsPerPage int
}

type Message struct {
	Kind string
	Etag string
	Id string
	Snippet messageSnippet
	AuthorDetails messageAuthorDetails
	yt *Client
}

type messageSnippet struct {
	Type string
	LiveChatId string
	AuthorChannelId string
	PublishedAt string
	HasDisplayContent bool
	DisplayMessage string
	TextMessageDetails messageDetails
}
type messageDetails struct {
	MessageText string
}


type messageAuthorDetails struct {
	ChannelId string
	ChannelUrl string
	DisplayName string
	ProfileImageUrl string
	IsVerified bool
	IsChatOwner bool
	IsChatSponsor bool
	IsChatModerator bool
}

func (msg Message) Reply(content string) (Message, error) {
	return msg.yt.SendMessage(msg.Snippet.LiveChatId, "@" + msg.AuthorDetails.DisplayName + " " + content)git
}
func (msg Message) Delete(timeout time.Duration) {
	go func() {
		time.Sleep(timeout)
		msg.yt.DeleteMessage(msg.Id)
	}()
}