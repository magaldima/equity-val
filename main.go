package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

var (
	ticker string
)

func init() {
	flag.StringVar(&ticker, "ticker", "AAPL", "A stock ticker")
}

func main() {
	flag.Parse()
	c := colly.NewCollector()

	summaryQueue := make(chan struct {
		string
		float64
	})

	c.OnXML("//*[@id=\"quote-summary\"]/div[2]/table/tbody/tr", func(xml *colly.XMLElement) {
		cells := xml.ChildTexts("td")
		// build a tuple
		if len(cells) == 2 {
			num, err := parseNumber(cells[1])
			if err == nil {
				summaryQueue <- struct {
					string
					float64
				}{cells[0], num}
			} else {
				log.Println(fmt.Errorf("failed to parse summary cells: %s, %s", cells, err))
			}
		}
	})

	go func() {
		for summaryEntries := range summaryQueue {
			log.Printf("Retrieved financial summary entry: %v", summaryEntries)
		}
	}()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s", strings.ToUpper(ticker))
	c.Visit(url)
}

func parseNumber(val string) (float64, error) {
	trimmed := strings.TrimSpace(val)
	last := trimmed[len(trimmed)-1]
	multiplier := func(b byte) float64 {
		switch b {
		case 'M':
			return 1000000
		case 'B':
			return 1000000000
		default:
			return 1
		}
	}(last)
	num, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, err
	}
	return num * multiplier, nil
}
