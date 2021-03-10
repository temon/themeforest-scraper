package models

type Design struct {
	Url               string `json:"url"`
	CatName           string `json:"catName"`
	PreviewUrl        string `json:"previewUrl"`
	Name              string `json:"name"`
	Image             string `json:"image"`
	Price             string `json:"price"`
	Sales             string `json:"sales"`
	Comments          string `json:"comments"`
	SellerName        string `json:"sellerName"`
	SellerUrl         string `json:"sellerUrl"`
	Created           string `json:"created"`
	LastUpdated       string `json:"lastUpdated"`
	Description       string `json:"description"`
	HighResolution    string `json:"highResolution"`
	CompatibleBrowser string `json:"compatibleBrowser"`
	CompatibleWith    string `json:"compatibleWith"`
	Included          string `json:"included"`
	Column            string `json:"column"`
	Documentation     string `json:"documentation"`
	Layout            string `json:"layout"`
	Tags              string `json:"tags"`
}
