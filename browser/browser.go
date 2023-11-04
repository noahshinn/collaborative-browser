package browser

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"webbot/browser/language"
	"webbot/browser/virtualid"
	"webbot/compilers"
	"webbot/compilers/html2md"
	"webbot/runner/trajectory"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Browser struct {
	mu           *sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	options      []BrowserOption
	vIDGenerator virtualid.VirtualIDGenerator
	translators  map[language.Language]compilers.Translator
}

type BrowserOption string

const (
	BrowserOptionNotHeadless                       BrowserOption = "not-headless"
	BrowserOptionAttemptToDisableAutomationMessage               = "attempt-to-disable-automation-message"
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

func (b *Browser) Render(lang language.Language) (location string, content string, err error) {
	var html string
	if translator, ok := b.translators[lang]; !ok {
		return "", "", fmt.Errorf("unsupported language: %s", lang)
	} else if err := chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		node, err := dom.GetDocument().Do(ctx)
		if err != nil {
			return fmt.Errorf("error getting document: %w", err)
		}
		html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
		if err != nil {
			return fmt.Errorf("error getting outer html: %w", err)
		}
		return nil
	})); err != nil {
		return "", "", err
	} else if translation, err := translator.Translate(html); err != nil {
		return "", "", fmt.Errorf("error translating html to %s: %w", lang, err)
	} else if location, err := b.GetLocation(); err != nil {
		return "", "", fmt.Errorf("error getting location: %w", err)
	} else {
		return location, translation, nil
	}
}

func (b *Browser) GetLocation() (string, error) {
	var url string
	if err := chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return chromedp.Location(&url).Do(ctx)
	})); err != nil {
		return "", err
	} else {
		return url, nil
	}
}

func NewBrowser(ctx context.Context, options ...BrowserOption) *Browser {
	vIDGenerator := virtualid.NewIncrIntVirtualIDGenerator()
	ops := chromedp.DefaultExecAllocatorOptions[:]
	for _, option := range options {
		switch option {
		case BrowserOptionNotHeadless:
			ops = append(ops, chromedp.Flag("headless", false))
		case BrowserOptionAttemptToDisableAutomationMessage:
			ops = append(ops, chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"))
			ops = append(ops, chromedp.Flag("enable-automation", false))
		default:
		}
	}
	htmlToMDTranslator := html2md.NewHTML2MDTranslator(nil)
	translatorMap := map[language.Language]compilers.Translator{
		language.LanguageMD: htmlToMDTranslator,
	}
	parentCtx, _ := chromedp.NewExecAllocator(context.Background(), ops...)
	browserCtx, cancel := chromedp.NewContext(parentCtx)
	return &Browser{
		mu:           &sync.Mutex{},
		ctx:          browserCtx,
		cancel:       cancel,
		options:      options,
		vIDGenerator: vIDGenerator,
		translators:  translatorMap,
	}
}
