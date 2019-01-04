package main

import (
	"fmt"
	"time"

	"github.com/gocolly/colly"
)

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("paulosuzart.github.io"),
		colly.MaxDepth(2),
		colly.Async(true),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 5,
		RandomDelay: 5 * time.Second,
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		targetLink := e.Request.AbsoluteURL(e.Attr("href"))
		c.Visit(targetLink)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("Just got response for path %s\n", r.Request.URL.EscapedPath())
	})

	c.Visit("https://paulosuzart.github.io/")
	c.Wait()

}
