package models

type LicenseModules struct {
	LicenseModule []LicenseModule `json:"modules"`
	Links         []Link          `json:"links"`
}

type Link struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type LicenseModule struct {
	Name  string `json:"name"`
	Links []Link `json:"links"`
}
