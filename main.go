package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"io/ioutil"
	"log"
	"strings"
	"temmo/models"
)

const baseUrl = "https://themeforest.net"
const startUrl = "/category/wordpress?sort=date"

var (
	baseCollector = buildCollector()
)

func main() {
	scrapeCategory()
}

func buildCollector() *colly.Collector {
	allowedDomains := []string{"themeforest.net", "www.themeforest.net"}

	collector := colly.NewCollector(
		colly.AllowedDomains(allowedDomains...),
		colly.CacheDir("./themeforest_cache"),
		colly.Async(true),
	)

	extensions.RandomUserAgent(collector)
	extensions.Referer(collector)

	return collector
}

func scrapeCategory() {
	allCategories := make([]models.Category, 0)
	collector := baseCollector.Clone()
	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting Base: ", request.URL.String())
	})

	collector.OnHTML(`ul[data-test-selector="category-filter"] li`, func(element *colly.HTMLElement) {
		url := element.ChildAttr(`li a`, "href")
		if !strings.HasPrefix(url, "/search?sort") && !strings.HasPrefix(url, "/category/wordpress?sort") {
			name := element.ChildText(`li a`)
			_cat := models.Category{
				Url:      url,
				Name:     name,
				Children: make([]models.Category, 0),
			}
			scrapeWPThemes(fmt.Sprint(baseUrl, url))
			allCategories = append(allCategories, _cat)
		}
	})

	url := fmt.Sprint(baseUrl, startUrl)
	err := collector.Visit(url)
	if err != nil {
		log.Println("Got error when scrape visit:", url)
	}

	collector.Wait()
	writeToJson(allCategories, "categories.json")
}

func scrapeWPThemes(catUrl string) {
	allDesigns := make([]models.Design, 0)

	collector := baseCollector.Clone()

	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting Page: ", request.URL.String())
	})

	detailCollector := collector.Clone()

	// visit all page
	collector.OnHTML(`nav[role="navigation"] li a[href]`, func(element *colly.HTMLElement) {
		link := element.Attr("href")

		if !strings.HasPrefix(link, "/category/site-templates/corporate") {
			return
		}

		err := element.Request.Visit(link)
		if err != nil {
			log.Println("Got error when scrape visit:", link)
		}
	})

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
		log.Println("Extract detail page", element.Request.URL)

		previewUrl := element.ChildAttr(`a.btn-icon.live-preview`, "href")
		log.Println("live preview:", previewUrl)

		name := element.ChildText(`div.item-header h1.t-heading.is-hidden-phone`)
		log.Println("name:", name)

		image := element.ChildAttr(`div.item-preview a img`, "src")
		log.Println("image:", image)

		price := element.ChildText("div.item-header__price b.t-currency span.js-item-header__price")
		log.Println("price:", price)

		sales := element.ChildText(`div.sidebar-stats__item div.box > strong.sidebar-stats__number`)
		log.Println("sales:", sales)

		comments := element.ChildText(`div.sidebar-stats__item div.box a.t-link strong.sidebar-stats__number`)
		log.Println("comments:", comments)

		sellerName := element.ChildText(`div.media__body h2 a.t-link[rel="author"]`)
		log.Println("sellerName:", sellerName)

		sellerUrl := element.ChildAttr(`div.media__body h2 a.t-link[rel="author"]`, "href")
		log.Println("sellerUrl:", sellerUrl)

		description, err := element.DOM.Find(`div.user-html`).Html()
		if err != nil {
			log.Println("Cannot get description")
		}
		log.Println("description:", description)

		design := models.Design{
			Url:         url,
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
				log.Println("lastUpdated:", lastUpdated)
			case "Created":
				created := element.ChildText(`td:nth-child(2) span`)
				design.Created = created
				log.Println("created:", created)
			case "High Resolution":
				highResolution := element.ChildText(`td:nth-child(2) a`)
				design.HighResolution = highResolution
				log.Println("highResolution:", highResolution)
			case "Compatible Browsers":
				compatibleBrowser := element.ChildText(`td:nth-child(2)`)
				design.CompatibleBrowser = compatibleBrowser
				log.Println("compatibleBrowser:", compatibleBrowser)
			case "Compatible With":
				compatibleWith := element.ChildText(`td:nth-child(2)`)
				design.CompatibleWith = compatibleWith
				log.Println("compatibleWith:", compatibleWith)
			case "ThemeForest Files Included":
				included := element.ChildText(`td:nth-child(2)`)
				design.Included = included
				log.Println("included:", included)
			case "Columns":
				columns := element.ChildText(`td:nth-child(2)`)
				design.Column = columns
				log.Println("columns:", columns)
			case "Documentation":
				documentation := element.ChildText(`td:nth-child(2)`)
				design.Documentation = documentation
				log.Println("Documentation:", documentation)
			case "Layout":
				layout := element.ChildText(`td:nth-child(2)`)
				design.Layout = layout
				log.Println("layout:", layout)
			case "Tags":
				tags := element.ChildText(`td:nth-child(2)`)
				design.Tags = tags
				log.Println("tags:", tags)
			}
		})

		allDesigns = append(allDesigns, design)
	})

	detailCollector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting Detail Page: ", request.URL.String())
	})

	err := collector.Visit(catUrl)
	if err != nil {
		log.Println("Got error when scrape visit:", catUrl)
	}

	collector.Wait()
	detailCollector.Wait()

	// write to json
	writeToJson(allDesigns, "designs.json")
}

func writeToJson(data interface{}, name string) {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Println("Can not create a json")
		return
	}
	_ = ioutil.WriteFile(name, file, 0644)
}
