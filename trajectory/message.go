package trajectory

import "fmt"

type Message struct {
	Render
	Handoff
	Author MessageAuthor
	Text   string

	ItemIsMessage
}

type MessageAuthor string

const (
	MessageAuthorUser             MessageAuthor = "user"
	MessageAuthorAgent            MessageAuthor = "agent"
	MessageAuthorInternalFeedback MessageAuthor = "internal_feedback"
)

func NewMessage(author MessageAuthor, text string) TrajectoryItem {
	return &Message{
		Author: author,
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
