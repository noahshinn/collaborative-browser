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
	"webbot/trajectory"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
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
	BrowserOptionHeadful                           BrowserOption = "headful"
	BrowserOptionAttemptToDisableAutomationMessage BrowserOption = "attempt-to-disable-automation-message"
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
		return b.Navigate(action.URL)
	default:
		return fmt.Errorf("unsupported browser action type: %s", action.Type)
	}
}

func (b *Browser) run(actions ...chromedp.Action) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return chromedp.Run(b.ctx, actions...)
}

func (b *Browser) Click(id virtualid.VirtualID) error {
	if !b.vIDGenerator.IsValidVirtualID(id) {
		return fmt.Errorf("invalid virtual id: %s", id)
	}
	return b.run(enableLifeCycleEvents(), clickAndWaitFor(virtualid.VirtualIDElementQuery(id), "networkIdle"))
}

func (b *Browser) SendKeys(id virtualid.VirtualID, keys string) error {
	if !b.vIDGenerator.IsValidVirtualID(id) {
		return fmt.Errorf("invalid virtual id: %s", id)
	} else if keys == "" {
		return errors.New("keys cannot be empty")
	}
	return b.run(chromedp.SendKeys(virtualid.VirtualIDElementQuery(id), keys))
}

func (b *Browser) Navigate(u string) error {
	if valid, err := IsValidURL(u); !valid {
		return err
	}
	return b.run(enableLifeCycleEvents(), navigateAndWaitFor(u, "networkIdle"))
}

func enableLifeCycleEvents() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		err := page.Enable().Do(ctx)
		if err != nil {
			return err
		}
		err = page.SetLifecycleEventsEnabled(true).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	}
}

func navigateAndWaitFor(url string, eventName string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		_, _, _, err := page.Navigate(url).Do(ctx)
		if err != nil {
			return err
		}

		return waitFor(ctx, eventName)
	}
}

func clickAndWaitFor(id string, eventName string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		err := chromedp.Click(id).Do(ctx)
		if err != nil {
			return err
		}

		return waitFor(ctx, eventName)
	}
}

// Implementation from https://github.com/chromedp/chromedp/issues/431#issuecomment-592950397
// waitFor blocks until eventName is received.
// Examples of events you can wait for:
//
//	init, DOMContentLoaded, firstPaint,
//	firstContentfulPaint, firstImagePaint,
//	firstMeaningfulPaintCandidate,
//	load, networkAlmostIdle, firstMeaningfulPaint, networkIdle
//
// This is not super reliable, I've already found incidental cases where
// networkIdle was sent before load. It's probably smart to see how
// puppeteer implements this exactly.
func waitFor(ctx context.Context, eventName string) error {
	ch := make(chan struct{})
	cctx, cancel := context.WithCancel(ctx)
	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventLifecycleEvent:
			if e.Name == eventName {
				cancel()
				close(ch)
			}
		}
	})
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}

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
		case BrowserOptionHeadful:
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
