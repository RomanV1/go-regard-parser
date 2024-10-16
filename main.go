package main

import (
	"bufio"
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type SiteContent struct {
	ProductName string
	Price       int
}

func main() {
	urls := make(chan string)

	go func() {
		err := ParseURLsFromFile(urls, "urls.txt")
		if err != nil {
			panic(err)
		}
	}()

	goods, totalPrice := ParallelDownload(urls, 10)

	for url, content := range goods {
		fmt.Printf("URL: %s \nProduct name: %s \nPrice: %vp. \n\n", url, content.ProductName, content.Price)
	}

	fmt.Printf("Total price: %vр.", totalPrice)
}

func ParallelDownload(urls chan string, numWorkers int) (map[string]SiteContent, int) {
	result := make(map[string]SiteContent)
	summary := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range urls {
				content := FetchContent(url)

				mu.Lock()
				result[url] = content
				summary += content.Price
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return result, summary
}

func FetchContent(url string) SiteContent {
	var content SiteContent

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (compatible; GoColly/1.0)"),
	)

	c.OnHTML(".Product_title__42hYI", func(e *colly.HTMLElement) {
		content.ProductName = e.Text
	})

	c.OnHTML(".PriceBlock_price__j_PbO > span:nth-child(1)", func(e *colly.HTMLElement) {
		priceString := e.Text
		cleanedString := strings.ReplaceAll(priceString, " ", "")
		re := regexp.MustCompile("[^0-9]")
		cleanedString = re.ReplaceAllString(cleanedString, "")

		price, err := strconv.Atoi(cleanedString)
		if err != nil {
			fmt.Println("Error converting string to integer:", err)
			return
		}

		content.Price = price
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Ошибка:", err)
	})

	err := c.Visit(url)
	if err != nil {
		log.Fatal(err)
	}

	return content
}

func ParseURLsFromFile(ch chan string, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ch <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	close(ch)
	return nil
}
