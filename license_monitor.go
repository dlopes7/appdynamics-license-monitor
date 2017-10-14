package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/dlopes7/go-appdynamics-rest-api/appdrest"
)

var templatesMetrics = map[string]string{
	"units-used":                     "name=Custom Metrics|Licensing|%s|%s|Agents|%s|Units Used,value=%d\n",
	"maximum-allowed-licenses":       "name=Custom Metrics|Licensing|%s|%s|Agents|%s|Units Allowed,value=%s\n",
	"number-of-provisioned-licenses": "name=Custom Metrics|Licensing|%s|%s|Agents|%s|Units Provisioned,value=%s\n",
	"hours-to-expire":                "name=Custom Metrics|Licensing|%s|%s|Agents|%s|Hours to Expire,value=%d\n",
}

func getControllersFromJSON() []*appdrest.Controller {
	file := "./conf.json"
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err.Error())
	}

	var controllers []*appdrest.Controller
	err = json.Unmarshal(raw, &controllers)
	if err != nil {
		panic(err.Error())
	}
	return controllers
}

func differenceFromNow(timeToCompare int64) int64 {
	now := time.Now().Unix() * 1000
	differenceHours := (timeToCompare - now) / 1000 / 60 / 60

	if differenceHours < 0 {
		differenceHours = 0
	}
	return differenceHours
}

func report(client *appdrest.Client) {
	defer wg.Done()
	acc, err := client.Account.GetMyAccount()
	if err != nil {
		panic(err.Error())
	}

	licenseModules, err := client.Account.GetLicenseModules(acc.ID)
	if err != nil {
		panic(err.Error())
	}

	wg.Add(len(licenseModules))
	for _, licenseModule := range licenseModules {
		defer wg.Done()

		go func(licenseModule *appdrest.LicenseModule) {
			wg.Add(1)
			defer wg.Done()
			properties, err := client.Account.GetLicenseProperties(acc.ID, licenseModule.Name)
			if err != nil {
				if err.(*appdrest.APIError).Code != 404 {
					panic(err.Error())
				}
			}

			for _, property := range properties {

				if val, ok := templatesMetrics[property.Name]; ok {
					fmt.Printf(val, client.Controller.Host, acc.Name, licenseModule.Name, property.Value)
				}
				if property.Name == "expiry-date" {
					value, err := strconv.ParseInt(property.Value, 10, 64)
					if err != nil {
						panic(err.Error())
					}
					hoursRemaining := differenceFromNow(value)
					fmt.Printf(templatesMetrics["hours-to-expire"], client.Controller.Host, acc.Name, licenseModule.Name, hoursRemaining)
				}
			}

		}(licenseModule)

		go func(licenseModule *appdrest.LicenseModule) {
			wg.Add(1)
			defer wg.Done()
			usages, err := client.Account.GetLicenseUsages(acc.ID, licenseModule.Name)
			if err != nil {
				if err.(*appdrest.APIError).Code != 404 {
					panic(err.Error())
				}
			}
			if len(usages) == 0 {
				fmt.Printf(templatesMetrics["units-used"], client.Controller.Host, acc.Name, licenseModule.Name, 0)
			} else {
				lastUsage := usages[len(usages)-1]
				fmt.Printf(templatesMetrics["units-used"], client.Controller.Host, acc.Name, licenseModule.Name, lastUsage.TotalUnitsUsed)
			}

		}(licenseModule)

	}
}

var wg sync.WaitGroup

func main() {
	controllers := getControllersFromJSON()
	wg.Add(len(controllers))

	for _, controller := range controllers {
		client := appdrest.NewClient(controller.Protocol, controller.Host, controller.Port, controller.User, controller.Password, controller.Account)
		go report(client)
	}
	wg.Wait()
	os.Exit(0)
}
