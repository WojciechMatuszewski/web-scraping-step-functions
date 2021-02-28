package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/gocolly/colly"
)

type Handler func(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error)

func newHandler() Handler {
	return func(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error) {
		inURL, found := payload["url"].(string)
		if !found {
			return nil, errors.New("url not found within payload")
		}

		u, err := url.Parse(inURL)
		if err != nil {
			return nil, err
		}

		c := colly.NewCollector(
			colly.MaxDepth(1),
		)
		// On every a element which has href attribute call callback
		var urls []string
		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			link := e.Attr("href")
			urls = append(urls, fmt.Sprintf("%v%v", u.Scheme+u.Host, link))

			e.Request.Visit(link)
		})

		err = c.Visit(inURL)
		if err != nil {
			return nil, err
		}

		payload["urls"] = urls
		return payload, nil
	}
}
