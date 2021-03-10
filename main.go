package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	uuid "github.com/nu7hatch/gouuid"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"sync"
	"temmo/models"
)

const baseUrl = "https://themeforest.net"
const startUrl = "/category/wordpress?sort=date"

var (
	baseCollector = buildCollector()
	baseFolder    = fmt.Sprintf("data-%s", getUUID())
	smx           sync.Mutex
)

func main() {
	startScrape()
}

func getUUID() string {
	u, _ := uuid.NewV4()
	return u.String()
}

func buildCollector() *colly.Collector {
	allowedDomains := []string{"themeforest.net", "www.themeforest.net"}

	collector := colly.NewCollector(
		colly.AllowedDomains(allowedDomains...),
		colly.CacheDir(fmt.Sprintf("./%s/themeforest_cache", baseFolder)),
	)

	extensions.RandomUserAgent(collector)
	extensions.Referer(collector)

	return collector
}

func startScrape() {
	allCategories := make([]models.Category, 0)
	collector := baseCollector.Clone()
	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting Base: ", request.URL.String())
	})

	collector.OnHTML(`div[data-test-selector="search-filters"] ul[data-test-selector="category-filter"] li`, func(element *colly.HTMLElement) {
		url := element.ChildAttr(`li a`, "href")
		if !strings.HasPrefix(url, "/search?sort") && !strings.HasPrefix(url, "/category/wordpress?sort") {
			name := element.ChildText(`li a`)
			_cat := models.Category{
				Url:      url,
				Name:     name,
				Children: make([]models.Category, 0),
			}

			scrapePerCategory(fmt.Sprint(baseUrl, url), name)

			smx.Lock()
			allCategories = append(allCategories, _cat)
			smx.Unlock()
		}
	})

	url := fmt.Sprint(baseUrl, startUrl)
	err := collector.Visit(url)
	if err != nil {
		log.Println("Got error when scrape visit:", url)
	}

	collector.OnScraped(func(response *colly.Response) {
		writeToJson(allCategories, "categories.json")
	})

	collector.Wait()
}

func scrapePerCategory(catUrl string, catName string) {

	collector := baseCollector.Clone()

	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting Page: ", request.URL.String())
	})

	// visit all page
	collector.OnHTML(`nav[role="navigation"] li:nth-last-child(2) a[href]`, func(element *colly.HTMLElement) {
		link := element.Attr("href")

		if !strings.HasPrefix(link, "/category/wordpress/") {
			return
		}

		pageMax, _ := strconv.Atoi(element.Text)

		for i := 1; i <= pageMax; i++ {
			pageUrl := fmt.Sprintf("%s&page=%s", strings.ReplaceAll(catUrl, "#content", ""), strconv.Itoa(i))

			scrapePerPage(pageUrl, catName, strconv.Itoa(i))
		}
	})

	err := collector.Visit(catUrl)
	if err != nil {
		log.Println("Got error when scrape visit:", catUrl)
	}

	collector.Wait()
}

func scrapePerPage(catPageUrl string, catName string, page string) {
	allDesigns := make([]models.Design, 0)

	collector := baseCollector.Clone()

	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting Page: ", request.URL.String())
	})

	detailCollector := baseCollector.Clone()

	// visit detail page
	collector.OnHTML("._2Pk9X", func(element *colly.HTMLElement) {
		url := element.Attr("href")
		err := detailCollector.Visit(url)
		if err != nil {
			log.Println("Got error when scrape visit:", url)
		}
	})

	// collect detail page information
	detailCollector.OnHTML(`div.page`, func(element *colly.HTMLElement) {

		url := element.Request.URL.String()

		previewUrl := element.ChildAttr(`a.btn-icon.live-preview`, "href")

		name := element.ChildText(`div.item-header h1.t-heading.is-hidden-phone`)

		image := element.ChildAttr(`div.item-preview a img`, "src")

		price := element.ChildText("div.item-header__price b.t-currency span.js-item-header__price")

		sales := element.ChildText(`div.sidebar-stats__item div.box > strong.sidebar-stats__number`)

		comments := element.ChildText(`div.sidebar-stats__item div.box a.t-link strong.sidebar-stats__number`)

		sellerName := element.ChildText(`div.media__body h2 a.t-link[rel="author"]`)

		sellerUrl := element.ChildAttr(`div.media__body h2 a.t-link[rel="author"]`, "href")

		description, err := element.DOM.Find(`div.user-html`).Html()
		if err != nil {
			log.Println("Cannot get description")
		}

		design := models.Design{
			Url:         url,
			CatName:     catName,
			PreviewUrl:  previewUrl,
			Name:        name,
			Image:       image,
			Price:       price,
			Sales:       sales,
			Comments:    comments,
			SellerName:  sellerName,
			SellerUrl:   sellerUrl,
			Description: description,
		}

		element.ForEach(`div.meta-attributes tr`, func(i int, element *colly.HTMLElement) {
			switch element.ChildText(`td:first-child`) {
			case "Last Update":
				lastUpdated := element.ChildAttr(`time.updated`, "datetime")
				design.LastUpdated = lastUpdated
			case "Created":
				created := element.ChildText(`td:nth-child(2) span`)
				design.Created = created
			case "High Resolution":
				highResolution := element.ChildText(`td:nth-child(2) a`)
				design.HighResolution = highResolution
			case "Compatible Browsers":
				compatibleBrowser := element.ChildText(`td:nth-child(2)`)
				design.CompatibleBrowser = compatibleBrowser
			case "Compatible With":
				compatibleWith := element.ChildText(`td:nth-child(2)`)
				design.CompatibleWith = compatibleWith
			case "ThemeForest Files Included":
				included := element.ChildText(`td:nth-child(2)`)
				design.Included = included
			case "Columns":
				columns := element.ChildText(`td:nth-child(2)`)
				design.Column = columns
			case "Documentation":
				documentation := element.ChildText(`td:nth-child(2)`)
				design.Documentation = documentation
			case "Layout":
				layout := element.ChildText(`td:nth-child(2)`)
				design.Layout = layout
			case "Tags":
				tags := element.ChildText(`td:nth-child(2)`)
				design.Tags = tags
			}
		})

		smx.Lock()
		allDesigns = append(allDesigns, design)
		smx.Unlock()
	})

	detailCollector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting Detail Page: ", request.URL.String())
	})

	err := collector.Visit(catPageUrl)
	if err != nil {
		log.Println("Got error when scrape visit:", catPageUrl)
	}

	collector.Wait()

	detailCollector.Wait()

	// write data per category to json
	uuidName := getUUID()
	fileName := fmt.Sprintf("design-%s-%s-%s.json", catName, page, uuidName)
	writeToJson(allDesigns, fileName)
}

func writeToJson(data interface{}, name string) {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Println("Can not create a json")
		return
	}
	fileName := fmt.Sprintf("./%s/%s", baseFolder, name)
	_ = ioutil.WriteFile(fileName, file, 0644)
}
