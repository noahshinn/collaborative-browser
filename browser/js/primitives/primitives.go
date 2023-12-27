package primitives

import (
	"fmt"

	"github.com/chromedp/chromedp"
)

func ClickByQuerySelector(query string) chromedp.Action {
	js := fmt.Sprintf(`function clickByQuerySelector(query) {
	const element = document.querySelector(query);
	if (element) {
		element.click();
	} else {
		throw new Error("element not found");
	}
}
clickByQuerySelector('%s');`, query)
	return chromedp.Evaluate(js, nil)
}

func SendTextByQuerySelector(query string, text string) chromedp.Action {
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
	return chromedp.Evaluate(js, nil)
}

func WaitForPageLoad(ms int) chromedp.Action {
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
waitForPageLoad(%d);`, ms)
	return chromedp.Evaluate(js, nil)
}

// func LocalStorageRead(key string, valuePtr *string) chromedp.Action {
// 	js := fmt.Sprintf(`function localStorageRead(key) {
// 	return localStorage.getItem(key);
// }
// localStorageRead('%s');`, key)
// 	return chromedp.Evaluate(js, valuePtr)
// }

func LocalStorageWrite(key string, valueJsonString string) chromedp.Action {
	js := fmt.Sprintf(`async function localStorageWrite(key, value) {
	const response = await fetch("http://localhost:2334/api-sls", {
		method: "POST",
		body: JSON.stringify({
			key: key,
			value: value
		}),
	});
	if (!response.ok) {
		throw new Error("HTTP error " + response.status);
	}
}
localStorageWrite('%s', '%s')
	.then(() => console.log("completed"));`, key, valueJsonString)
	return chromedp.Evaluate(js, nil)
}

func AddClickEventHandler() chromedp.Action {
	return AddEventHandler("click", `async (event) => {
	const port = globalThis.collaborativeBrowserStore?.port;
	if (port === undefined) {
		throw new Error("port is undefined; must initialize it first");
	}
	const host = "localhost";
	const endpoint = "/api-sls/traj";
	const id = event.target.getAttribute("data-vid");
	if (id === null) {
		throw new Error("data-vid attribute not found");
	}
	const response = await fetch("http://" + host + ":" + port + endpoint, {
		method: "POST",
		body: JSON.stringify({
			action: "click",
			args: {
				id,
			}
		}),
	});
	
	if (!response.ok) {
		throw new Error("HTTP error " + response.status);
	}
}`)
}

func AddEventHandler(eventType string, jsCallback string) chromedp.Action {
	js := fmt.Sprintf(`function addEventHandler(eventType, jsCallback) {
	document.addEventListener(eventType, jsCallback);
}
addEventHandler('%s', %s);`, eventType, jsCallback)
	return chromedp.Evaluate(js, nil)
}

func InitializeGlobalStore() chromedp.Action {
	js := `function initializeGlobalStore() {
	globalThis.collaborativeBrowserStore = {};
}
initializeGlobalStore();`
	return chromedp.Evaluate(js, nil)
}

func ReadGlobalVar(key string) chromedp.Action {
	js := fmt.Sprintf(`function readGlobalVar(key) {
	const collaborativeBrowserStore = globalThis.collaborativeBrowserStore;
	if (collaborativeBrowserStore === undefined) {
		throw new Error("collaborativeBrowserStore is undefined; must initialize it first");
	}
	if (collaborativeBrowserStore[key] === undefined) {
		throw new Error("key not found");
	}
	return collaborativeBrowserStore[key];
}
readGlobalVar('%s');`, key)
	return chromedp.Evaluate(js, nil)
}

func WriteGlobalVar(key string, value string) chromedp.Action {
	js := fmt.Sprintf(`function addGlobalVar(key, value) {
	if (globalThis.collaborativeBrowserStore === undefined) {
		throw new Error("collaborativeBrowserStore is undefined; must initialize it first");
	}
	globalThis.collaborativeBrowserStore[key] = value;
}
addGlobalVar('%s', '%s');`, key, value)
	return chromedp.Evaluate(js, nil)
}
