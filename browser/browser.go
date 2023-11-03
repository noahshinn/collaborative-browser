package browser

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"webbot/actor"
	"webbot/browser/virtualid"
	"webbot/compilers/html2md"

	"github.com/chromedp/chromedp"
)

type Browser struct {
	mu                 *sync.Mutex
	ctx                context.Context
	cancel             context.CancelFunc
	options            *BrowserOptions
	currentDisplay     *BrowserDisplay
	vIDGenerator       virtualid.VirtualIDGenerator
	htmlToMDTranslater *html2md.HTML2MDTranslater
}

type BrowserOptions struct {
	RunHeadless bool
}

type BrowserDisplay struct {
	text string
}

func (bd *BrowserDisplay) Text() string {
	return bd.text
}

func (b *Browser) AcceptAction(action *actor.BrowserAction) error {
	switch action.Type {
	case actor.BrowserActionTypeClick:
		return b.Click(action.ID)
	case actor.BrowserActionTypeSendKeys:
		return b.SendKeys(action.ID, action.Text)
	default:
		return fmt.Errorf("unsupported browser action type: %s", action.Type)
	}
}

func (b *Browser) run(actions ...chromedp.Action) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := chromedp.Run(b.ctx, actions...); err != nil {
		return fmt.Errorf("error running actions: %w", err)
	}
	return b.UpdateDisplay()
}

// TODO: implement a general wait method that is robust for almost all page loads
func (b *Browser) Click(id virtualid.VirtualID) error {
	if !b.vIDGenerator.IsValidVirtualID(id) {
		return fmt.Errorf("invalid virtual id: %s", id)
	}
	return b.run(chromedp.Click(virtualid.VirtualIDElementQuery(id)))
}

func (b *Browser) SendKeys(id virtualid.VirtualID, keys string) error {
	if !b.vIDGenerator.IsValidVirtualID(id) {
		return fmt.Errorf("invalid virtual id: %s", id)
	} else if keys == "" {
		return errors.New("keys cannot be empty")
	}
	return b.run(chromedp.SendKeys(virtualid.VirtualIDElementQuery(id), keys))
}

func (b *Browser) GoTo(u string) error {
	if valid, err := IsValidURL(u); !valid {
		return err
	}
	return b.run(chromedp.Navigate(u))
}

type Language string

const (
	LanguageHTML Language = "html"
	LanguageMD   Language = "md"
)

func (b *Browser) Render(language Language) (string, error) {
	switch language {
	case LanguageHTML:
		// TODO: implement
		return "", nil
	case LanguageMD:
		// TODO: implement
		return "", nil
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}
}

func (b *Browser) UpdateDisplay() error {
	// TODO: implement
	return nil
}

func NewBrowser(ctx context.Context, options *BrowserOptions) *Browser {
	vIDGenerator := virtualid.NewIncrIntVirtualIDGenerator()
	ops := chromedp.DefaultExecAllocatorOptions[:]
	if options.RunHeadless {
		ops = append(ops, chromedp.Flag("headless", true))
	}
	parentCtx, _ := chromedp.NewExecAllocator(context.Background(), ops...)
	browserCtx, cancel := chromedp.NewContext(parentCtx)
	return &Browser{
		mu:             &sync.Mutex{},
		ctx:            browserCtx,
		cancel:         cancel,
		options:        options,
		currentDisplay: nil,
		vIDGenerator:   vIDGenerator,
	}
}
