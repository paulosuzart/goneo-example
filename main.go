package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gocolly/colly"
	neo "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

func getConnection(pool neo.DriverPool) neo.Conn {
	conn, err := pool.OpenPool()
	if err != nil {
		log.Panic("Unable to open a connection to storage")
	}
	return conn
}

func connect(sourceURL string, targetURL string, depth int, pool neo.DriverPool) {
	conn := getConnection(pool)
	defer conn.Close()

	_, err := conn.ExecNeo(`
	MATCH (source:Page {url: {sourceUrl}})
	MERGE (target:Page {url: {targetUrl}})
	MERGE (source)-[r:LINK]->(target) return r`,
		map[string]interface{}{
			"sourceUrl": sourceURL,
			"targetUrl": targetURL,
			"depth":     depth,
		},
	)

	if err != nil {
		log.Panic("Failed to create link data")
	}
}

func merge(absoluteURL string, depth int, pool neo.DriverPool) {
	conn := getConnection(pool)
	defer conn.Close()

	_, err := conn.ExecNeo(`
	MERGE (s:Page {url: {url}})
	return s`,
		map[string]interface{}{
			"url":   absoluteURL,
			"depth": depth,
		})
	if err != nil {
		log.Panic("Failed to merge page")
	}
}

func main() {
	driver, err := neo.NewClosableDriverPool("bolt://localhost:7687", 20)

	if err != nil {
		log.Panic("Unable to establish connection to neo4j")
	}
	defer driver.Close()

	c := colly.NewCollector(
		colly.AllowedDomains("paulosuzart.github.io"),
		colly.MaxDepth(2),
		colly.Async(true),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		RandomDelay: 5 * time.Second,
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		targetLink := e.Request.AbsoluteURL(e.Attr("href"))
		connect(e.Request.URL.String(), targetLink, e.Request.Depth, driver)
		c.Visit(targetLink)
	})

	c.OnResponse(func(r *colly.Response) {
		merge(r.Request.URL.String(), r.Request.Depth, driver)
		fmt.Printf("Just got response for path %s\n", r.Request.URL.EscapedPath())
	})

	c.Visit("https://paulosuzart.github.io/")
	c.Wait()

}
