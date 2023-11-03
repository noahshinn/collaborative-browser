package browser

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"webbot/browser/language"
	"webbot/browser/virtualid"
	"webbot/compilers/html2md"
	"webbot/runner/trajectory"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Browser struct {
	mu                 *sync.Mutex
	ctx                context.Context
	cancel             context.CancelFunc
	options            []BrowserOption
	currentDisplay     *BrowserDisplay
	vIDGenerator       virtualid.VirtualIDGenerator
	htmlToMDTranslater *html2md.HTML2MDTranslater
}

type BrowserOption string

const (
	BrowserOptionHeadless BrowserOption = "headless"
)

type BrowserDisplay struct {
	text string
}

func (bd *BrowserDisplay) Text() string {
	return bd.text
}

func (b *Browser) AcceptAction(action *trajectory.BrowserAction) error {
	switch action.Type {
	case trajectory.BrowserActionTypeClick:
		return b.Click(action.ID)
	case trajectory.BrowserActionTypeSendKeys:
		return b.SendKeys(action.ID, action.Text)
	case trajectory.BrowserActionTypeNavigate:
		return b.GoTo(action.URL)
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

func NewBrowser(ctx context.Context, options ...BrowserOption) *Browser {
	vIDGenerator := virtualid.NewIncrIntVirtualIDGenerator()
	ops := chromedp.DefaultExecAllocatorOptions[:]
	for _, option := range options {
		switch option {
		case BrowserOptionHeadless:
			ops = append(ops, chromedp.Flag("headless", true))
		default:
		}
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
