package models

type LicenseModules struct {
	LicenseModules []LicenseModule `json:"modules"`
	Links          []Link          `json:"links"`
}

type Link struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type LicenseModule struct {
	Name  string `json:"name"`
	Links []Link `json:"links"`
}

// Properties : this is a list of Property
type Properties struct {
	Properties []Property `json:"properties"`
}

// Property : this is a Property of a license
type Property struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Usages : list of License Usage objects
type Usages struct {
	Usages []Usage
}

// Usage represents a single license usage for a period of 5 minutes
type Usage struct {
	ID               string `json:"id"`
	AccountID        int    `json:"accountId"`
	AgentType        string `json:"agentType"`
	CreatedOn        int64  `json:"createdOn"`
	CreatedOnISODate string `json:"createdOnIsoDate"`
	UnitsUsed        int    `json:"unitsUsed"`
	UnitsProvisioned int    `json:"unitsProvisioned"`
	UnitsAllowed     int    `json:"unitsAllowed"`
}
