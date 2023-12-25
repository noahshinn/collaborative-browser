package browser

import (
	"collaborativebrowser/browser/js/primitives"
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
	"strconv"
	"strings"
	"sync"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Browser struct {
	mu                *sync.Mutex
	ctx               context.Context
	cancel            context.CancelFunc
	options           *Options
	vIDGenerator      virtualid.VirtualIDGenerator
	translators       map[language.Language]translators.Translator
	display           *BrowserDisplay
	isRunningHeadless bool
	localServerPort   int
}

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

const pageLoadWaitTimeoutMs = 10000

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

func (b *Browser) AcceptAction(action *trajectory.TrajectoryItem) (string, error) {
	var err error
	var response string
	switch action.Type {
	case trajectory.TrajectoryItemClick:
		if err = b.Click(action.ID); err != nil {
			return "", fmt.Errorf("error clicking: %w", err)
		}
		response = fmt.Sprintf("clicked %s", action.ID)
	case trajectory.TrajectoryItemSendKeys:
		if err = b.SendKeys(action.ID, action.Text); err != nil {
			return "", fmt.Errorf("error sending keys: %w", err)
		}
		keysDisplay := action.Text
		if len(keysDisplay) > 10 {
			keysDisplay = keysDisplay[:10] + "..."
		}
		response = fmt.Sprintf("sent keys \"%s\" to %s", keysDisplay, action.ID)
	case trajectory.TrajectoryItemNavigate:
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

func (b *Browser) Click(id string) error {
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
		primitives.WaitForPageLoad(pageLoadWaitTimeoutMs)
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

func (b *Browser) SendKeys(id string, keys string) error {
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
		primitives.WaitForPageLoad(pageLoadWaitTimeoutMs)
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
		primitives.WaitForPageLoad(pageLoadWaitTimeoutMs)
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
	newOptions := b.options
	newOptions.RunHeadful = true
	newBrowserCtx, newBrowserCancelFunc := newBrowser(ctx, newOptions)
	b.cancel()
	b.ctx = newBrowserCtx
	b.cancel = newBrowserCancelFunc
	if err := b.Navigate(b.display.Location); err != nil {
		return fmt.Errorf("error navigating to current location %s: %w", b.display.Location, err)
	}
	b.isRunningHeadless = false
	return nil
}

func (b *Browser) DoesVirtualIDExist(virtualID string) (bool, error) {
	var exists bool
	if err := b.run(chromedp.Evaluate(fmt.Sprintf("document.querySelector('[data-vid=\"%s\"]') !== null", virtualID), &exists)); err != nil {
		return false, err
	} else {
		return exists, nil
	}
}

func (b *Browser) ClickByVirtualID(virtualID string) error {
	return b.run(primitives.ClickByQuerySelector(fmt.Sprintf("[data-vid=\"%s\"]", virtualID)))
}

func (b *Browser) SendTextByVirtualID(virtualID string, text string) error {
	return b.run(primitives.SendTextByQuerySelector(fmt.Sprintf("[data-vid=\"%s\"]", virtualID), text))
}

func (b *Browser) CheckElementTypeForQuerySelector(query string) (ElementType, error) {
	js := fmt.Sprintf(`function checkElementTypeForQuerySelector(query) {
	const element = document.querySelector(query);
	if (element) {
		if (element.tagName === 'INPUT') {
			return 'input';
		} else if (element.tagName === 'TEXTAREA') {
			return 'textarea';
		} else if (element.tagName === 'BUTTON') {
			return 'button';
		} else if (element.tagName === 'A') {
			return 'a';
		} else {
			return 'other';
		}
	} else {
		throw new Error("element not found");
	}
}
checkElementTypeForQuerySelector('%s');`, query)
	var elementType string
	if err := b.run(chromedp.Evaluate(js, &elementType)); err != nil {
		return "", fmt.Errorf("error checking element type for query selector: %w", err)
	} else {
		return ElementType(elementType), nil
	}
}

func (b *Browser) CheckElementTypeForVirtualID(virtualID string) (ElementType, error) {
	return b.CheckElementTypeForQuerySelector(fmt.Sprintf("[data-vid=\"%s\"]", virtualID))
}

func (b *Browser) GetAllVisibleVirtualIDs() ([]string, error) {
	js := `function getAllVisibleVirtualIDs() {
	const elements = document.querySelectorAll('[data-vid]');
	const dataVids = Array.from(elements).map(element => element.getAttribute('data-vid'));
	return dataVids;
}
getAllVisibleVirtualIDs();`
	var virtualIDs []string
	if err := b.run(chromedp.Evaluate(js, &virtualIDs)); err != nil {
		return nil, fmt.Errorf("error getting all visible virtual IDs: %w", err)
	} else if virtualIDs == nil {
		return nil, fmt.Errorf("error getting all visible virtual IDs: virtual IDs is nil")
	} else {
		return virtualIDs, nil
	}
}

func (b *Browser) DoesSupportAriaLabels() (bool, error) {
	js := `function doesSupportAriaLabels() {
	const ariaLabelElems = document.querySelectorAll('[aria-label]');
	return ariaLabelElems.length > 0;
}
doesSupportAriaLabels();
`
	var supportsAriaLabels bool
	if err := b.run(chromedp.Evaluate(js, &supportsAriaLabels)); err != nil {
		return false, fmt.Errorf("error checking if browser supports aria labels: %w", err)
	} else {
		return supportsAriaLabels, nil
	}
}

func (b *Browser) isPageLoaded() (bool, error) {
	js := `function isPageLoaded() {
	return document.readyState !== 'loading';
}
isPageLoaded();`
	var isLoaded bool
	if err := b.run(chromedp.Evaluate(js, &isLoaded)); err != nil {
		return false, fmt.Errorf("error checking if page is loaded: %w", err)
	} else {
		return isLoaded, nil
	}
}

func (b *Browser) RunHeadless(ctx context.Context) error {
	if b.isRunningHeadless {
		log.Println("requested to run the browser in headless mode but this browser is already running in headless mode")
		return nil
	}
	log.Println("running the browser in headless mode; warning: you will lose all non-location state from the current browser")
	newOptions := b.options
	newOptions.RunHeadful = false
	newBrowserCtx, newBrowserCancelFunc := newBrowser(ctx, newOptions)
	b.cancel()
	b.ctx = newBrowserCtx
	b.cancel = newBrowserCancelFunc
	if err := b.Navigate(b.display.Location); err != nil {
		return fmt.Errorf("error navigating to current location %s: %w", b.display.Location, err)
	}
	b.isRunningHeadless = true
	return nil
}

func buildChromedpOptions(options *Options) []func(*chromedp.ExecAllocator) {
	ops := chromedp.DefaultExecAllocatorOptions[:]
	if options != nil {
		if options.RunHeadful {
			ops = append(ops, chromedp.Flag("headless", false))
		}
		if options.AttemptToDisableAutomationMessage {
			ops = append(ops, chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"))
			ops = append(ops, chromedp.Flag("enable-automation", false))
		}
	}
	return ops
}

func newBrowser(ctx context.Context, options *Options) (browserCtx context.Context, cancelFunc context.CancelFunc) {
	ops := buildChromedpOptions(options)
	parentCtx, _ := chromedp.NewExecAllocator(ctx, ops...)
	browserCtx, cancel := chromedp.NewContext(parentCtx)
	return browserCtx, cancel
}

type Options struct {
	LocalStorageServerPort            int
	RunHeadful                        bool
	AttemptToDisableAutomationMessage bool
}

// TODO: initialization when toggling headful/headless mode

func (b *Browser) Initialize() error {
	if err := b.run(primitives.InitializeGlobalStore()); err != nil {
		return fmt.Errorf("error initializing global store: %w", err)
	} else if err := b.run(primitives.WriteGlobalVar("localStorageServerPort", strconv.Itoa(b.localServerPort))); err != nil {
		return fmt.Errorf("error writing local storage server port: %w", err)
	}
	return nil
}

func NewBrowser(ctx context.Context, options *Options) *Browser {
	vIDGenerator := virtualid.NewIncrIntVirtualIDGenerator()
	htmlToMDTranslator := html2md.NewHTML2MDTranslator(nil)
	translatorMap := map[language.Language]translators.Translator{
		language.LanguageMD: htmlToMDTranslator,
	}
	isRunningHeadless := true
	localServerPort := 2334
	if options != nil {
		if options.LocalStorageServerPort > 0 {
			localServerPort = options.LocalStorageServerPort
		}
		if options.RunHeadful {
			isRunningHeadless = false
		}
	}
	browserCtx, cancel := newBrowser(ctx, options)
	b := Browser{
		mu:                &sync.Mutex{},
		ctx:               browserCtx,
		cancel:            cancel,
		options:           options,
		vIDGenerator:      vIDGenerator,
		translators:       translatorMap,
		display:           &BrowserDisplay{},
		isRunningHeadless: isRunningHeadless,
		localServerPort:   localServerPort,
	}
	if err := b.Initialize(); err != nil {
		log.Println("error initializing browser:", err)
		panic(err)
	}
	return &b
}
