package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Account : struct used to get the Account details
type Account struct {
	ID string `json:"id"`
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

// Controllers : List of controllers to monitor
type Controllers struct {
	Controllers []Controller `json:"controllers"`
}

// Controller : struct used for the Controller connection
type Controller struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Account  string `json:"account"`
	Protocol string `json:"protocol"`
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func getAccountID(controller Controller) string {

	urlTemplate := "%s://%s:%d/controller/api/accounts/myaccount"
	url := fmt.Sprintf(urlTemplate, controller.Protocol, controller.Host, controller.Port)

	acc := new(Account)
	getJSON(controller, url, acc)

	return acc.ID
}

func getJSON(controller Controller, url string, target interface{}) error {
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
	if resp.StatusCode > 400 {
		panic(fmt.Sprintf("Error accessing the API %d", resp.StatusCode))
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func differenceFromNow(timeToCompare int64) int64 {
	now := time.Now().Unix() * 1000
	differenceHours := (timeToCompare - now) / 1000 / 60 / 60
	return differenceHours
}

func process(controller Controller) {

	accID := getAccountID(controller)

	urlTemplate := "%s://%s:%d/controller/api/accounts/%s/licensemodules/java/properties"
	url := fmt.Sprintf(urlTemplate, controller.Protocol, controller.Host, controller.Port, accID)

	target := new(Properties)

	getJSON(controller, url, target)

	for _, element := range target.Properties {
		if element.Name == "expiry-date" {
			value, err := strconv.ParseInt(element.Value, 10, 64)
			if err != nil {
				panic(err.Error())
			}
			diff := differenceFromNow(value)
			fmt.Printf("name=Custom Metrics|Licensing|%s|Hours Remaining,value=%d\n", controller.Name, diff)
		}
	}
}

func getControllersFromJSON() Controllers {
	file := "./conf.json"
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err.Error())
	}

	var controllers Controllers
	json.Unmarshal(raw, &controllers)
	return controllers
}

func main() {

	controllers := getControllersFromJSON()

	for _, controller := range controllers.Controllers {
		process(controller)
	}
	os.Exit(0)
}
