package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"io/ioutil"
	"log"
	"strings"
)

type Design struct {
	Url string `json:"url"`
	PreviewUrl string `json:"previewUrl"`
	Name string `json:"name"`
	Image string `json:"image"`
	Price string `json:"price"`
	Sales string `json:"sales"`
	Comments string `json:"comments"`
	SellerName string `json:"sellerName"`
	SellerUrl string `json:"sellerUrl"`
	Created string `json:"created"`
	LastUpdated string `json:"lastUpdated"`
	Description string `json:"description"`
	HighResolution string `json:"highResolution"`
	CompatibleBrowser string `json:"compatibleBrowser"`
	CompatibleWith string `json:"compatibleWith"`
	Included string `json:"included"`
	Column string `json:"column"`
	Documentation string `json:"documentation"`
	Layout string `json:"layout"`
	Tags string `json:"tags"`
}

func main() {
	startUrl := "https://themeforest.net/category/site-templates/corporate?sort=date"
	allowedDomains := []string {"themeforest.net", "www.themeforest.net"}

	allDesigns := make([]Design, 0)

	collector := colly.NewCollector(
		colly.AllowedDomains(allowedDomains...),
		colly.CacheDir("./themeforest_cache"),
	)

	detailCollector := collector.Clone()

	// visit all page
	collector.OnHTML(`nav[role="navigation"] li a[href]`, func(element *colly.HTMLElement) {
		link := element.Attr("href")

		if !strings.HasPrefix(link, "/category/site-templates/corporate"){
			return
		}

		element.Request.Visit(link)
	})

	// visit detail page
	collector.OnHTML("._2Pk9X", func(element *colly.HTMLElement) {
		url := element.Attr("href")

		detailCollector.Visit(url)
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

		design := Design{
			Url: url,
			PreviewUrl: previewUrl,
			Name: name,
			Image: image,
			Price: price,
			Sales: sales,
			Comments: comments,
			SellerName: sellerName,
			SellerUrl: sellerUrl,
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

	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting Page: ", request.URL.String())
	})

	detailCollector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting Detail Page: ", request.URL.String())
	})

	collector.Visit(startUrl)

	// write to json
	writeToJson(allDesigns)
}

func writeToJson(data []Design) {
	file, err := json.MarshalIndent(data, "", " ")
	if(err != nil) {
		log.Println("Can not create a json")
		return
	}
	_ = ioutil.WriteFile("designs.json", file, 0644)
}
