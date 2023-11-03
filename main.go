package main

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

func main() {
	ops := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))
	parentCtx, _ := chromedp.NewExecAllocator(context.Background(), ops...)
	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://scholar.google.com/"),
	)
	if err != nil {
		log.Fatal("Error while performing the automation logic:", err)
	}
	err = chromedp.Run(ctx,
		chromedp.SendKeys(`#gs_hdr_tsi`, "transformers is all you need", chromedp.ByID, chromedp.NodeVisible),
	)
	if err != nil {
		log.Fatal("Error while typing the query:", err)
	}
}
