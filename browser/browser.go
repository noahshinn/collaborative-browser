package browser

import (
	"collaborativebrowser/browser/language"
	"collaborativebrowser/browser/virtualid"
	"collaborativebrowser/trajectory"
	"collaborativebrowser/translators"
	"collaborativebrowser/translators/html2md"
	"collaborativebrowser/utils/slicesx"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Browser struct {
	mu                *sync.Mutex
	ctx               context.Context
	cancel            context.CancelFunc
	options           []BrowserOption
	vIDGenerator      virtualid.VirtualIDGenerator
	translators       map[language.Language]translators.Translator
	display           *BrowserDisplay
	isRunningHeadless bool
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

type ElementType string

const (
	ElementTypeButton   ElementType = "button"
	ElementTypeInput    ElementType = "input"
	ElementTypeLink     ElementType = "a"
	ElementTypeTextArea ElementType = "textarea"
	ElementTypeOther    ElementType = "other"
)

func (b *Browser) updateDisplay() error {
	if location, err := b.getLocation(); err != nil {
		return fmt.Errorf("error getting location: %w", err)
	} else if html, err := b.getHTML(); err != nil {
		return fmt.Errorf("error getting html for location %s: %w", location, err)
	} else if md, err := b.translators[language.LanguageMD].Translate(html); err != nil {
		return fmt.Errorf("error translating html to %s for location %s: %w", language.LanguageMD, location, err)
	} else {
		b.display.HTML = html
		b.display.Location = location
		b.display.MD = md
		return nil
	}
}

func (b *Browser) AcceptAction(action *trajectory.BrowserAction) (string, error) {
	var err error
	var response string
	switch action.Type {
	case trajectory.BrowserActionTypeClick:
		if err = b.Click(action.ID); err != nil {
			return "", fmt.Errorf("error clicking: %w", err)
		}
		response = fmt.Sprintf("clicked %s", action.ID)
	case trajectory.BrowserActionTypeSendKeys:
		if err = b.SendKeys(action.ID, action.Text); err != nil {
			return "", fmt.Errorf("error sending keys: %w", err)
		}
		keysDisplay := action.Text
		if len(keysDisplay) > 10 {
			keysDisplay = keysDisplay[:10] + "..."
		}
		response = fmt.Sprintf("sent keys \"%s\" to %s", keysDisplay, action.ID)
	case trajectory.BrowserActionTypeNavigate:
		if err = b.Navigate(action.URL); err != nil {
			return "", fmt.Errorf("error navigating: %w", err)
		}
		response = fmt.Sprintf("navigated to %s", action.URL)
	default:
		return "", fmt.Errorf("unsupported browser action type: %s", action.Type)
	}
	if err := b.updateDisplay(); err != nil {
		return "", fmt.Errorf("error updating display: %w", err)
	}
	return response, nil
}

func (b *Browser) run(actions ...chromedp.Action) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return chromedp.Run(b.ctx, actions...)
}

func (b *Browser) Click(id virtualid.VirtualID) error {
	previousLocation := b.display.Location
	if !b.vIDGenerator.IsValidVirtualID(id) {
		return fmt.Errorf("invalid virtual id: %s", id)
	} else if exists, err := b.DoesVirtualIDExist(string(id)); err != nil {
		return fmt.Errorf("error checking if virtual id exists: %w", err)
	} else if !exists {
		return fmt.Errorf("virtual id does not exist: %s", id)
	} else if elementType, err := b.CheckElementTypeForVirtualID(string(id)); err != nil {
		return fmt.Errorf("error checking element type for virtual id: %w", err)
	} else if elementType != ElementTypeButton && elementType != ElementTypeLink {
		return fmt.Errorf("cannot click element type %s", elementType)
	} else if err := b.ClickByVirtualID(string(id)); err != nil {
		return fmt.Errorf("error clicking by virtual id: %w", err)
	}
	if loaded, err := b.isPageLoaded(); err != nil {
		log.Println("error checking if page is loaded:", err)
	} else if !loaded {
		b.waitForPageLoad()
	}
	if err := b.updateDisplay(); err != nil {
		return fmt.Errorf("error updating display: %w", err)
	}
	if previousLocation != b.display.Location {
		if supportsAriaLabels, err := b.DoesSupportAriaLabels(); err != nil {
			log.Println("error checking if browser supports aria labels:", err)
		} else if !supportsAriaLabels {
			log.Println("warning: this page does not support aria labels")
		}
	}
	return nil
}

func (b *Browser) SendKeys(id virtualid.VirtualID, keys string) error {
	if !b.vIDGenerator.IsValidVirtualID(id) {
		return fmt.Errorf("invalid virtual id: %s", id)
	} else if keys == "" {
		return errors.New("keys cannot be empty")
	} else if exists, err := b.DoesVirtualIDExist(string(id)); err != nil {
		return fmt.Errorf("error checking if virtual id exists: %w", err)
	} else if !exists {
		return fmt.Errorf("virtual id does not exist: %s", id)
	} else if elementType, err := b.CheckElementTypeForVirtualID(string(id)); err != nil {
		return fmt.Errorf("error checking element type for virtual id: %w", err)
	} else if elementType != ElementTypeInput && elementType != ElementTypeTextArea {
		return fmt.Errorf("cannot send keys to element type %s", elementType)
	} else if err := b.SendTextByVirtualID(string(id), keys); err != nil {
		return fmt.Errorf("error sending text by virtual id: %w", err)
	}
	if loaded, err := b.isPageLoaded(); err != nil {
		log.Println("error checking if page is loaded:", err)
	} else if !loaded {
		b.waitForPageLoad()
	}
	return b.updateDisplay()
}

func (b *Browser) Navigate(URL string) error {
	u, err := GetCanonicalURL(URL)
	if err != nil {
		return fmt.Errorf("error ensuring scheme: %w", err)
	}
	valid, err := IsValidURL(u)
	if !valid {
		return fmt.Errorf("invalid url %s: %w", u, err)
	}
	err = b.run(chromedp.Navigate(u))
	if err != nil {
		return fmt.Errorf("error navigating to %s: %w", u, err)
	}
	if loaded, err := b.isPageLoaded(); err != nil {
		log.Println("error checking if page is loaded:", err)
	} else if !loaded {
		b.waitForPageLoad()
	}
	if supportsAriaLabels, err := b.DoesSupportAriaLabels(); err != nil {
		log.Println("error checking if browser supports aria labels:", err)
	} else if !supportsAriaLabels {
		log.Println("warning: this page does not support aria labels")
	}
	return b.updateDisplay()
}

func (b *Browser) addVirtualIDs() error {
	existingVirtualIDs, err := b.GetAllVisibleVirtualIDs()
	if err != nil {
		log.Println("error getting existing virtual ids:", err)
		return nil
	}
	// TODO: invoke custom vID generator
	js := fmt.Sprintf(`function addDataVidAttribute(excludeIDs) {
	const reservedIDs = {};
	excludeIDs.forEach(id => reservedIDs[id] = true);
	const elements = document.querySelectorAll('button, input, a, textarea');
	let counter = 0;
	elements.forEach(element => {
		if (element.offsetParent !== null && !element.hasAttribute('data-vid')) {
			while (reservedIDs["vid-" + counter.toString()]) {
				counter++;
			}
			element.setAttribute('data-vid', "vid-" + counter.toString());
			counter++;
		}
	});
}
addDataVidAttribute(%s);`, "["+strings.Join(slicesx.Map(existingVirtualIDs, func(virtualID string, _ int) string {
		return fmt.Sprintf(`"%s"`, virtualID)
	}), ", ")+"]")
	return b.run(chromedp.Evaluate(js, nil))
}

func (b *Browser) Render(lang language.Language) (content string, err error) {
	if location, err := b.getLocation(); err != nil {
		return "", fmt.Errorf("error getting location: %w", err)
	} else if translator, ok := b.translators[lang]; !ok {
		return "", fmt.Errorf("unsupported language: %s", lang)
	} else if err := b.addVirtualIDs(); err != nil {
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

func (b *Browser) Cancel() {
	b.cancel()
}

func (b *Browser) IsRunningHeadless() bool {
	return b.isRunningHeadless
}

func (b *Browser) RunHeadful(ctx context.Context) error {
	if !b.isRunningHeadless {
		log.Println("requested to run the browser in headful mode but this browser is already running in headful mode")
		return nil
	}
	log.Println("running the browser in headful mode; warning: you will lose all non-location state from the current browser")
	newOps := append(b.options, BrowserOptionHeadful)
	newBrowserCtx, newBrowserCancelFunc := newBrowser(ctx, newOps...)
	b.cancel()
	b.ctx = newBrowserCtx
	b.cancel = newBrowserCancelFunc
	if err := b.Navigate(b.display.Location); err != nil {
		return fmt.Errorf("error navigating to current location %s: %w", b.display.Location, err)
	}
	b.isRunningHeadless = false
	return nil
}

func (b *Browser) RunHeadless(ctx context.Context) error {
	if b.isRunningHeadless {
		log.Println("requested to run the browser in headless mode but this browser is already running in headless mode")
		return nil
	}
	log.Println("running the browser in headless mode; warning: you will lose all non-location state from the current browser")
	newOps := slicesx.Filter(b.options, func(option BrowserOption, _ int) bool {
		return option != BrowserOptionHeadful
	})
	newBrowserCtx, newBrowserCancelFunc := newBrowser(ctx, newOps...)
	b.cancel()
	b.ctx = newBrowserCtx
	b.cancel = newBrowserCancelFunc
	if err := b.Navigate(b.display.Location); err != nil {
		return fmt.Errorf("error navigating to current location %s: %w", b.display.Location, err)
	}
	b.isRunningHeadless = true
	return nil
}

func buildOptions(options ...BrowserOption) []func(*chromedp.ExecAllocator) {
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
	return ops
}

func newBrowser(ctx context.Context, options ...BrowserOption) (browserCtx context.Context, cancelFunc context.CancelFunc) {
	ops := buildOptions(options...)
	parentCtx, _ := chromedp.NewExecAllocator(ctx, ops...)
	browserCtx, cancel := chromedp.NewContext(parentCtx)
	return browserCtx, cancel
}

func NewBrowser(ctx context.Context, options ...BrowserOption) *Browser {
	isRunningHeadless := !slicesx.Contains(options, BrowserOptionHeadful)
	vIDGenerator := virtualid.NewIncrIntVirtualIDGenerator()
	htmlToMDTranslator := html2md.NewHTML2MDTranslator(nil)
	translatorMap := map[language.Language]translators.Translator{
		language.LanguageMD: htmlToMDTranslator,
	}
	browserCtx, cancel := newBrowser(ctx, options...)
	return &Browser{
		mu:                &sync.Mutex{},
		ctx:               browserCtx,
		cancel:            cancel,
		options:           options,
		vIDGenerator:      vIDGenerator,
		translators:       translatorMap,
		display:           &BrowserDisplay{},
		isRunningHeadless: isRunningHeadless,
	}
}
