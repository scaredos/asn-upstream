package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type lookUp struct {
	Data struct {
		Neighbours []struct {
			Asn     int    `json:"asn"`
			Type    string `json:"type"`
			Power   int    `json:"power"`
			V4Peers int    `json:"v4_peers"`
			V6Peers int    `json:"v6_peers"`
		} `json:"neighbours"`
	} `json:"data"`
}

func ipFind(ip string) string {
	r, _ := regexp.Compile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	return r.FindString(ip)
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: ./main <ip>")
		os.Exit(0)
	}
	ip := ipFind(os.Args[1])
	if ip == "" {
		fmt.Println("no ip provided")
		os.Exit(0)
	}
	var client http.Client
	req, _ := http.NewRequest("GET", "https://ipinfo.io/"+ip+"/org", nil)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	body := string(bodyBytes)
	asn := strings.Split(body, " ")[0]
	fmt.Println("IP ASN:", strings.Replace(body, "\n", "", -1))
	req, _ = http.NewRequest("GET", "https://stat.ripe.net/data/asn-neighbours/data.json?resource="+asn, nil)
	resp, _ = client.Do(req)
	defer resp.Body.Close()
	bodyBytes, _ = ioutil.ReadAll(resp.Body)
	var lookupResp lookUp
	json.Unmarshal(bodyBytes, &lookupResp)
	neighbors := lookupResp.Data.Neighbours
	fmt.Println("Neighbors:")
	for i := 0; i < len(neighbors); i++ {
		if neighbors[i].Type == "left" {
			neighborAsn := "AS" + strconv.Itoa(neighbors[i].Asn)
			req, _ = http.NewRequest("GET", "https://ipinfo.io/"+neighborAsn, nil)
			req.Header.Set("User-Agent", "Mozilla/5.0")
			resp, _ = client.Do(req)
			defer resp.Body.Close()
			bodyBytes, _ = ioutil.ReadAll(resp.Body)
			body = string(bodyBytes)
			body = strings.Split(body, `<h6 class="font-weight-normal opacity-50">`)[1]
			body = strings.Split(body, "</h6>")[0]
			fmt.Println("    " + body + " (upstream)")
		} else if neighbors[i].Type == "right" {
			neighborAsn := "AS" + strconv.Itoa(neighbors[i].Asn)
			req, _ = http.NewRequest("GET", "https://ipinfo.io/"+neighborAsn, nil)
			req.Header.Set("User-Agent", "Mozilla/5.0")
			resp, _ = client.Do(req)
			defer resp.Body.Close()
			bodyBytes, _ = ioutil.ReadAll(resp.Body)
			body = string(bodyBytes)
			body = strings.Split(body, `<h6 class="font-weight-normal opacity-50">`)[1]
			body = strings.Split(body, "</h6>")[0]
			fmt.Println("    " + body + " (downstream)")
		}
	}
}
