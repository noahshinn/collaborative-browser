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
	"time"

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
	} else {
		b.display.HTML = html
		b.display.Location = location
		if b.display.MD == "" {
			b.display.MD = "No MD display available yet."
		}
		return nil
	}
}

func (b *Browser) AcceptAction(action *trajectory.BrowserAction) (string, error) {
	var err error
	switch action.Type {
	case trajectory.BrowserActionTypeClick:
		if err = b.Click(action.ID); err != nil {
			return "", fmt.Errorf("error clicking: %w", err)
		} else if err = b.updateDisplay(); err != nil {
			return "", fmt.Errorf("error updating display: %w", err)
		}
		return fmt.Sprintf("clicked %s", action.ID), nil
	case trajectory.BrowserActionTypeSendKeys:
		if err = b.SendKeys(action.ID, action.Text); err != nil {
			return "", fmt.Errorf("error sending keys: %w", err)
		}
		keysDisplay := action.Text
		if len(keysDisplay) > 10 {
			keysDisplay = keysDisplay[:10] + "..."
		}
		return fmt.Sprintf("sent keys \"%s\" to %s", keysDisplay, action.ID), nil
	case trajectory.BrowserActionTypeNavigate:
		if err = b.Navigate(action.URL); err != nil {
			return "", fmt.Errorf("error navigating: %w", err)
		} else if err = b.updateDisplay(); err != nil {
			return "", fmt.Errorf("error updating display: %w", err)
		}
		return fmt.Sprintf("navigated to %s", action.URL), nil
	default:
		return "", fmt.Errorf("unsupported browser action type: %s", action.Type)
	}
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
	} else if err := b.updateDisplay(); err != nil {
		return fmt.Errorf("error updating display: %w", err)
	} else if previousLocation != b.display.Location {
		if supportsAriaLabels, err := b.DoesSupportAriaLabels(); err != nil {
			log.Println("error checking if browser supports aria labels:", err)
		} else if !supportsAriaLabels {
			log.Println("warning: browser does not support aria labels")
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
	} else {
		return b.SendTextByVirtualID(string(id), keys)
	}
}

func (b *Browser) Navigate(URL string) error {
	if u, err := GetCanonicalURL(URL); err != nil {
		return fmt.Errorf("error ensuring scheme: %w", err)
	} else if valid, err := IsValidURL(u); !valid {
		return fmt.Errorf("invalid url %s: %w", u, err)
	} else if err := b.run(chromedp.Navigate(u), chromedp.Sleep(1*time.Second)); err != nil {
		return fmt.Errorf("error navigating to %s: %w", u, err)
	} else if supportsAriaLabels, err := b.DoesSupportAriaLabels(); err != nil {
		log.Println("error checking if browser supports aria labels:", err)
	} else if !supportsAriaLabels {
		log.Println("warning: browser does not support aria labels")
	}
	return nil
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
	b.ctx = newBrowserCtx
	b.cancel = newBrowserCancelFunc
	if err := b.Navigate(b.display.Location); err != nil {
		return fmt.Errorf("error navigating to current location %s: %w", b.display.Location, err)
	}
	b.isRunningHeadless = false
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
