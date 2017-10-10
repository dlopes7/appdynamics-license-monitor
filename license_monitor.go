package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	models "github.com/dlopes7/license_monitor/models"
)

var templatesMetrics = map[string]string{
	"units-used":                     "name=Custom Metrics|Licensing|%s|Agents|%s|Units Used,value=%d\n",
	"maximum-allowed-licenses":       "name=Custom Metrics|Licensing|%s|Agents|%s|Units Allowed,value=%s\n",
	"number-of-provisioned-licenses": "name=Custom Metrics|Licensing|%s|Agents|%s|Units Provisioned,value=%s\n",
	"hours-to-expire":                "name=Custom Metrics|Licensing|%s|Agents|%s|Hours to Expire,value=%d\n",
}

var myClient = &http.Client{Timeout: 10 * time.Second}
var wg sync.WaitGroup

func getAccountID(controller models.Controller) string {

	urlTemplate := "%s://%s:%d/controller/api/accounts/myaccount"
	url := fmt.Sprintf(urlTemplate, controller.Protocol, controller.Host, controller.Port)

	acc := new(models.Account)
	fromJSONtoModel(controller, url, acc)

	return acc.ID
}

func fromJSONtoModel(controller models.Controller, url string, target interface{}) error {
	username := fmt.Sprintf("%s@%s", controller.User, controller.Account)
	password := controller.Password

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err.Error())
	}

	req.SetBasicAuth(username, password)

	resp, err := myClient.Do(req)
	if err != nil {
		panic(err.Error())
	}

	defer resp.Body.Close()
	if resp.StatusCode > 400 {
		err := fmt.Errorf("Error accessing the API %d", resp.StatusCode)
		return err
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func differenceFromNow(timeToCompare int64) int64 {
	now := time.Now().Unix() * 1000
	differenceHours := (timeToCompare - now) / 1000 / 60 / 60

	if differenceHours < 0 {
		differenceHours = 0
	}
	return differenceHours
}

func processLink(controller models.Controller, agentType string, link models.Link) {

	defer wg.Done()

	if link.Name == "usages" {
		params := "?showfiveminutesresolution=true"
		usages := new(models.Usages)
		url := strings.Replace(link.Href, "http", controller.Protocol, 1) + params
		fromJSONtoModel(controller, url, usages)

		if len(usages.Usages) == 0 {
			fmt.Printf(templatesMetrics["units-used"], controller.Name, agentType, 0)
		} else {
			mostRecentUsage := usages.Usages[len(usages.Usages)-1]
			fmt.Printf(templatesMetrics["units-used"], controller.Name, agentType, mostRecentUsage.UnitsUsed)
		}

	} else if link.Name == "properties" {
		properties := new(models.Properties)
		url := strings.Replace(link.Href, "http", controller.Protocol, 1)
		err := fromJSONtoModel(controller, url, properties)
		if err == nil {
			for _, property := range properties.Properties {

				if val, ok := templatesMetrics[property.Name]; ok {
					fmt.Printf(val, controller.Name, agentType, property.Value)
				}
				if property.Name == "expiry-date" {
					value, err := strconv.ParseInt(property.Value, 10, 64)
					if err != nil {
						panic(err.Error())
					}
					fmt.Printf(templatesMetrics["hours-to-expire"], controller.Name, agentType, differenceFromNow(value))
				}

			}
		}

	}

}

func processLicenseModules(controller models.Controller, accID string) {
	urlTemplate := "%s://%s:%d/controller/api/accounts/%s/licensemodules"
	url := fmt.Sprintf(urlTemplate, controller.Protocol, controller.Host, controller.Port, accID)

	licenseModules := new(models.LicenseModules)

	fromJSONtoModel(controller, url, licenseModules)

	for _, licenseModule := range licenseModules.LicenseModules {
		wg.Add(len(licenseModule.Links))
		for _, link := range licenseModule.Links {

			go processLink(controller, licenseModule.Name, link)

		}

	}
	wg.Wait()
}

func process(controller models.Controller) {

	accID := getAccountID(controller)
	processLicenseModules(controller, accID)

}

func getControllersFromJSON() models.Controllers {
	file := "./conf.json"
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err.Error())
	}

	var controllers models.Controllers
	err = json.Unmarshal(raw, &controllers)
	if err != nil {
		panic(err.Error())
	}
	return controllers
}

func main() {

	controllers := getControllersFromJSON()

	for _, controller := range controllers.Controllers {
		process(controller)
	}
	os.Exit(0)
}
