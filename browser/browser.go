package browser

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"webbot/browser/language"
	"webbot/browser/virtualid"
	"webbot/trajectory"
	"webbot/translators"
	"webbot/translators/html2md"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Browser struct {
	mu           *sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	options      []BrowserOption
	vIDGenerator virtualid.VirtualIDGenerator
	translators  map[language.Language]translators.Translator
	display      *BrowserDisplay
}

type BrowserOption string

const (
	BrowserOptionHeadful                           BrowserOption = "headful"
	BrowserOptionAttemptToDisableAutomationMessage BrowserOption = "attempt-to-disable-automation-message"
)

type BrowserDisplay struct {
	HTML     string
	MD       string
	Location string
}

func (b *Browser) AcceptAction(action *trajectory.BrowserAction) error {
	var err error
	switch action.Type {
	case trajectory.BrowserActionTypeClick:
		err = b.Click(action.ID)
	case trajectory.BrowserActionTypeSendKeys:
		err = b.SendKeys(action.ID, action.Text)
	case trajectory.BrowserActionTypeNavigate:
		err = b.Navigate(action.URL)
	default:
		return fmt.Errorf("unsupported browser action type: %s", action.Type)
	}
	if err != nil {
		return fmt.Errorf("error accepting action: %w", err)
	} else if location, err := b.getLocation(); err != nil {
		return fmt.Errorf("error getting location: %w", err)
	} else if html, err := b.getHTML(); err != nil {
		return fmt.Errorf("error getting html for location %s: %w", location, err)
	} else {
		b.display.HTML = html
		b.display.Location = location
		return nil
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
	return b.run(chromedp.Click(virtualid.VirtualIDElementQuery(id)), chromedp.Sleep(1*time.Second))
}

func (b *Browser) SendKeys(id virtualid.VirtualID, keys string) error {
	if !b.vIDGenerator.IsValidVirtualID(id) {
		return fmt.Errorf("invalid virtual id: %s", id)
	} else if keys == "" {
		return errors.New("keys cannot be empty")
	}
	return b.run(chromedp.SendKeys(virtualid.VirtualIDElementQuery(id), keys))
}

func (b *Browser) Navigate(URL string) error {
	if u, err := GetCanonicalURL(URL); err != nil {
		return fmt.Errorf("error ensuring scheme: %w", err)
	} else if valid, err := IsValidURL(u); !valid {
		return fmt.Errorf("invalid url %s: %w", u, err)
	} else {
		return b.run(chromedp.Navigate(u), chromedp.Sleep(1*time.Second))
	}
}

func (b *Browser) AddVirtualIDs() string {
	// TODO: invoke custom vID generator
	const f = `function addDataVidAttribute() {
	const elements = document.querySelectorAll('button, input, a');
	let counter = 0;
	elements.forEach(element => {
		element.setAttribute('data-vid', "vid-" + counter.toString());
		counter++;
		console.log(counter);
	});
}
addDataVidAttribute();`
	return f
}

func (b *Browser) Render(lang language.Language) (content string, err error) {
	if location, err := b.getLocation(); err != nil {
		return "", fmt.Errorf("error getting location: %w", err)
	} else if translator, ok := b.translators[lang]; !ok {
		return "", fmt.Errorf("unsupported language: %s", lang)
	} else if err := b.run(chromedp.Evaluate(b.AddVirtualIDs(), nil)); err != nil {
		return "", err
	} else if html, err := b.getHTML(); err != nil {
		return "", fmt.Errorf("error getting html for location %s: %w", location, err)
	} else if translation, err := translator.Translate(html); err != nil {
		return "", fmt.Errorf("error translating html to %s for location %s: %w", lang, location, err)
	} else {
		b.display = &BrowserDisplay{
			HTML:     html,
			MD:       translation,
			Location: location,
		}
		return translation, nil
	}
}

func (b *Browser) getLocation() (string, error) {
	var url string
	if err := b.run(chromedp.Location(&url)); err != nil {
		return "", err
	} else {
		return url, nil
	}
}

func (b *Browser) getHTML() (string, error) {
	var html string
	if err := chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		node, err := dom.GetDocument().Do(ctx)
		if err != nil {
			return err
		}
		html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
		return err
	})); err != nil {
		return "", err
	} else {
		return html, nil
	}
}

func (b *Browser) GetDisplay() *BrowserDisplay {
	return b.display
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
	translatorMap := map[language.Language]translators.Translator{
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
		display:      &BrowserDisplay{},
	}
}
