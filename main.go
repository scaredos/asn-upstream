package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

// type for API response (As Neighbors)
// https://stat.ripe.net/data/asn-neighbours/data.json?resource=
type asNeighbors struct {
	Data struct {
		Neighbours []struct {
			Asn  int    `json:"asn"`  // Neighbour's asn
			Type string `json:"type"` // Neighbour's type (left or right)
		} `json:"neighbours"`
	} `json:"data"`
}

// type for API response (IP overview)
// https://stat.ripe.net/data/prefix-overview/data.json?resource=
type ipToAs struct {
	Data struct {
		Asns []struct {
			Asn    int    `json:"asn"`    // ASN of IP
			Holder string `json:"holder"` // Holder of ASN
		} `json:"asns"`
	} `json:"data"`
}

// type for API response (AS Data)
// https://stat.ripe.net/data/as-overview/data.json?resource=
type asData struct {
	Data struct {
		Holder string `json:"holder"` // Holder of ASN
	} `json:"data"`
}

// Function to return an IPv4 address from a string
func ipFind(ip string) string {
	r := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	return r.FindString(ip)
}

func main() {
	if len(os.Args) == 1 { // show usage if no os arguments are given
		fmt.Println("usage: ./main <ip>") // e.x.: ./main 1.1.1.1
		os.Exit(0)
	}
	ip := ipFind(os.Args[1]) // return Ipv4 from os argument
	if ip == "" {            // if there's no ip in the argument, exit with error
		fmt.Println("no ip provided")
		os.Exit(0)
	}
	var client http.Client
	req, _ := http.NewRequest("GET", "https://stat.ripe.net/data/prefix-overview/data.json?resource="+ip, nil)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	var asName ipToAs
	json.Unmarshal(bodyBytes, &asName)
	asn := strconv.Itoa(asName.Data.Asns[0].Asn)
	asn = "AS" + asn
	fmt.Println("IP ASN:", asn, asName.Data.Asns[0].Holder)
	req, _ = http.NewRequest("GET", "https://stat.ripe.net/data/asn-neighbours/data.json?resource="+asn, nil)
	resp, _ = client.Do(req)
	defer resp.Body.Close() // Close body later
	bodyBytes, _ = ioutil.ReadAll(resp.Body)
	var asNeighbor asNeighbors
	json.Unmarshal(bodyBytes, &asNeighbor)
	neighbors := asNeighbor.Data.Neighbours
	fmt.Println("Neighbors:")
	for i := 0; i < len(neighbors); i++ {
		// if the ASN type is uncertain, skip
		if neighbors[i].Type == "uncertain" {
			break
		}
		neighborAsn := "AS" + strconv.Itoa(neighbors[i].Asn)
		req, _ = http.NewRequest("GET", "https://stat.ripe.net/data/as-overview/data.json?resource="+neighborAsn, nil)
		resp, _ = client.Do(req)
		defer resp.Body.Close()
		bodyBytes, _ = ioutil.ReadAll(resp.Body)
		var neighbor asData
		json.Unmarshal(bodyBytes, &neighbor)
		switch neighbors[i].Type {
		case "left":
			fmt.Printf("\t%s - %s (upstream)\n", neighborAsn, neighbor.Data.Holder)
		case "right":
			fmt.Printf("\t%s - %s (downstream)\n", neighborAsn, neighbor.Data.Holder)
		}
	}
}
