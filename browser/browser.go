package browser

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"webbot/actor"
	"webbot/browser/language"
	"webbot/browser/virtualid"
	"webbot/compilers/html2md"

	"github.com/chromedp/cdproto/dom"
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
	return chromedp.Run(b.ctx, actions...)
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

func (b *Browser) Render(lang language.Language) (string, error) {
	node, err := dom.GetDocument().Do(b.ctx)
	if err != nil {
		return "", fmt.Errorf("error getting document: %w", err)
	}
	html, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(b.ctx)
	if err != nil {
		return "", fmt.Errorf("error getting outer html: %w", err)
	}
	switch lang {
	case language.LanguageHTML:
		return html, nil
	case language.LanguageMD:
		return b.htmlToMDTranslater.Translate(html, b.vIDGenerator)
	default:
		return "", fmt.Errorf("unsupported language: %s", lang)
	}
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
