package trajectory

import "fmt"

type Message struct {
	Author MessageAuthor
	Text   string
}

type MessageAuthor string

const (
	MessageAuthorUser  MessageAuthor = "user"
	MessageAuthorAgent MessageAuthor = "agent"
)

func NewUserMessage(text string) TrajectoryItem {
	return &Message{
		Author: MessageAuthorUser,
		Text:   text,
	}
}

func NewAgentMessage(text string) TrajectoryItem {
	return &Message{
		Author: MessageAuthorAgent,
		Text:   text,
	}
}

func (m *Message) GetText() string {
	return fmt.Sprintf("%s: %s", m.Author, m.Text)
}

const DefaultAgentMessageAbbreviationLength = 100

func (m *Message) GetAbbreviatedText() string {
	text := m.GetText()
	if m.Author == MessageAuthorAgent && len(m.Text) > DefaultAgentMessageAbbreviationLength {
		return fmt.Sprintf("%s...", text[:DefaultAgentMessageAbbreviationLength])
	}
	return text
}

func (m *Message) ShouldHandoff() bool {
	return true
}
