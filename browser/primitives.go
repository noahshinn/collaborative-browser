package browser

import (
	"fmt"
	"log"

	"github.com/chromedp/chromedp"
)

func (b *Browser) DoesVirtualIDExist(virtualID string) (bool, error) {
	var exists bool
	if err := b.run(chromedp.Evaluate(fmt.Sprintf("document.querySelector('[data-vid=\"%s\"]') !== null", virtualID), &exists)); err != nil {
		return false, err
	} else {
		return exists, nil
	}
}

func (b *Browser) ClickByQuerySelector(query string) error {
	js := fmt.Sprintf(`function clickByQuerySelector(query) {
	const element = document.querySelector(query);
	if (element) {
		element.click();
	} else {
		throw new Error("element not found");
	}
}
clickByQuerySelector('%s');`, query)
	return b.run(chromedp.Evaluate(js, nil))
}

func (b *Browser) ClickByVirtualID(virtualID string) error {
	return b.ClickByQuerySelector(fmt.Sprintf("[data-vid=\"%s\"]", virtualID))
}

func (b *Browser) SendTextByQuerySelector(query string, text string) error {
	js := fmt.Sprintf(`function sendTextToSelector(selector, text) {
	try {
		var element = document.querySelector(selector);
		if (!element) {
			throw new Error('Element not found');
		}
		if (element.tagName === 'INPUT' || element.tagName === 'TEXTAREA') {
			element.value = text;
		} else if (element.textContent !== undefined) {
			element.textContent = text;
		} else {
			throw new Error('Element cannot receive text');
		}
	} catch (error) {
		throw new Error('Error in sendTextToSelector: ' + error.message);
	}
}
sendTextToSelector('%s', '%s');`, query, text)
	return b.run(chromedp.Evaluate(js, nil))
}

func (b *Browser) SendTextByVirtualID(virtualID string, text string) error {
	return b.SendTextByQuerySelector(fmt.Sprintf("[data-vid=\"%s\"]", virtualID), text)
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

const pageLoadWaitTimeoutMs = 10000

func (b *Browser) waitForPageLoad() {
	js := fmt.Sprintf(`function waitForPageLoad(timeoutMs) {
	return new Promise((resolve, reject) => {
		const timeout = setTimeout(() => {
			reject(new Error('Timeout waiting for page load'));
		}, timeoutMs);
		if (document.readyState === 'loading') {
			document.addEventListener('DOMContentLoaded', () => {
				clearTimeout(timeout);
				resolve();
			})
		} else {
			clearTimeout(timeout);
			resolve();
		}
	});
	}
}
waitForPageLoad(%d);`, pageLoadWaitTimeoutMs)
	if err := b.run(chromedp.Evaluate(js, nil)); err != nil {
		log.Printf("Error waiting for page load: %v", err)
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
