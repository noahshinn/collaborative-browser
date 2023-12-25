package trajectory

type MessageAuthor string

const (
	MessageAuthorUser             MessageAuthor = "user"
	MessageAuthorAgent            MessageAuthor = "agent"
	MessageAuthorInternalFeedback MessageAuthor = "internal_feedback"
)

func NewMessage(author MessageAuthor, text string) *TrajectoryItem {
	return &TrajectoryItem{
		Type:          TrajectoryItemMessage,
		ShouldHandoff: true,
		ShouldRender:  true,
		Author:        author,
		Text:          text,
	}
}
