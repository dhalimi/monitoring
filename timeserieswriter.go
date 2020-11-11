package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
)

//router.HandleFunc("/metrics", metricsStaticFile)

type metricTags struct {
	MulticloudID string `json:"multicloud_id"`
	ProjectID    string `json:"project_id"`
	CloudID      string `json:"cloud_id"`
	Region       string `json:"region"`
	Provider     string `json:"provider"`
	Service      string `json:"service"`
	Action       string `json:"action"`
	Outcome      string `json:"outcome"`
}

func timeseriesWriter(r metricTags) error {

	// see example - r := metricTags{"multicloudid", "projectid", "cloudid", "region", "domain", "provider", "service", "action", "outcome2"}

	// Creating the timeseries event file if it doesn't exist
	if _, err := os.Stat(eventFile); os.IsNotExist(err) {
		message := []byte("# Nebula events time series metrics")
		err := ioutil.WriteFile(eventFile, message, 0644)
		if err != nil {
			return err
		}
	}

	// Build the tags string and seperatly the outcome tag for later use
	tags := reflect.ValueOf(r)
	typeOfTags := tags.Type()
	data := ""
	outcome := ""
	for i := 0; i < tags.NumField(); i++ {
		if strings.ToLower(typeOfTags.Field(i).Name) == "outcome" {
			outcome = fmt.Sprintf(`%s="%v"`, strings.ToLower(typeOfTags.Field(i).Name), tags.Field(i).Interface())
		} else {
			data += fmt.Sprintf(`%s="%v",`, strings.ToLower(typeOfTags.Field(i).Name), tags.Field(i).Interface())
		}

	}

	//
	input, err := ioutil.ReadFile(eventFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")

	// check if result changed
	newEvent := true
	for i, line := range lines {
		if strings.Contains(line, data) {
			newEvent = false
			fmt.Println("timeseries exist")
			if !strings.Contains(line, outcome) {
				fmt.Println("outcome value changed")
				fmt.Println(lines[i])
				lines[i] = nebulaAudit + "{" + data + outcome + "}"
			}
		}
	}
	fmt.Println(newEvent)
	if newEvent {
		lines = append(lines, nebulaAudit+"{"+data+outcome+"}")
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(eventFile, []byte(output), 0644)
	if err != nil {
		return err
	}

	return nil

}

func metricsStaticFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	p := "." + r.URL.Path
	if p == "./metrics" {
		p = "./static"
	}
	http.ServeFile(w, r, p)
}
