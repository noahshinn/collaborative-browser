package browser

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"webbot/compilers/html2md"

	"github.com/chromedp/chromedp"
)

type Browser struct {
	mu                 *sync.Mutex
	ctx                context.Context
	cancel             context.CancelFunc
	options            *BrowserOptions
	currentDisplay     *BrowserDisplay
	vIDGenerator       VirtualIDGenerator
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

func (b *Browser) Run(actions ...chromedp.Action) (*BrowserDisplay, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := chromedp.Run(b.ctx, actions...); err != nil {
		return nil, fmt.Errorf("error running actions: %w", err)
	} else if err := b.UpdateDisplay(); err != nil {
		return nil, fmt.Errorf("error updating display: %w", err)
	} else {
		return b.currentDisplay, nil
	}
}

// TODO: implement a general wait method that is robust for almost all page loads
func (b *Browser) Click(id VirtualID) (*BrowserDisplay, error) {
	if !b.vIDGenerator.IsValidVirtualID(id) {
		return nil, fmt.Errorf("invalid virtual id: %s", id)
	} else if _, err := b.Run(chromedp.Click(VirtualIDElementQuery(id))); err != nil {
		return nil, fmt.Errorf("error clicking: %w", err)
	} else {
		return b.currentDisplay, nil
	}
}

func (b *Browser) SendKeys(id VirtualID, keys string) (*BrowserDisplay, error) {
	if !b.vIDGenerator.IsValidVirtualID(id) {
		return nil, fmt.Errorf("invalid virtual id: %s", id)
	} else if keys == "" {
		return nil, errors.New("keys cannot be empty")
	} else if _, err := b.Run(chromedp.SendKeys(VirtualIDElementQuery(id), keys)); err != nil {
		return nil, fmt.Errorf("error sending keys: %w", err)
	} else {
		return b.currentDisplay, nil
	}
}

func (b *Browser) GoTo(u string) (*BrowserDisplay, error) {
	if valid, err := IsValidURL(u); !valid {
		return nil, err
	} else if _, err := b.Run(chromedp.Navigate(u)); err != nil {
		return nil, fmt.Errorf("error navigating to url: %w", err)
	} else {
		return b.currentDisplay, nil
	}
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
	vIDGenerator := NewIncrIntVirtualIDGenerator()
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
