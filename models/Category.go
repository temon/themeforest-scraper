package models

type Category struct {
	Url      string     `json:"url"`
	Name     string     `json:"name"`
	Children []Category `json:"children"`
}
