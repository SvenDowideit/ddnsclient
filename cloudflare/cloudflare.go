package cloudflare

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SvenDowideit/ddnsclient/protocols"
)

// TODO: can I get the package name for this?
const driverName = "cloudflare"

// see https://api.cloudflare.com/#dns-records-for-a-zone-create-dns-record

func init() {
	protocols.RegisterDriver(driverName, New)
}

func New(host, ip, login, password string) (protocols.Driver, error) {
	c := Cloudflare{
		host:     host,
		ip:       ip,
		login:    login,
		password: password,

		apiUrl: "https://api.cloudflare.com/client/v4",

		headers: map[string]string{
			"X-Auth-Email": login,
			"X-Auth-Key":   password,
		},
	}

	return c, nil
}

type Cloudflare struct {
	host, ip, login, password string

	apiUrl  string
	headers map[string]string
}

func (c Cloudflare) Set() {
	zoneID := c.getZoneID()
	if zoneID == "" {
		fmt.Printf("ERROR: %s not found in zones\n", c.host)
		return
	}

	//Get list of existing records
	resp, err := get(c.apiUrl+"/zones/"+zoneID+"/dns_records", nil, c.headers)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}

	recordID := ""
	for _, v := range resp.Result {
		if v.Name == c.host {
			recordID = v.ID
			fmt.Printf("Currently set %s to %s\n", v.Name, v.Content)
		}
	}

	if recordID == "" {
		// create a new entry
		_, err = callJSON("POST", c.apiUrl+"/zones/"+zoneID+"/dns_records", map[string]string{
			"type":    "A",
			"name":    c.host,
			"content": c.ip,
			"ttl":     "120",
		}, c.headers)
	} else {
		// modify entry
		_, err = callJSON("PUT", c.apiUrl+"/zones/"+zoneID+"/dns_records/"+recordID, map[string]string{
			"type":    "A",
			"name":    c.host,
			"content": c.ip,
			"ttl":     "120",
		}, c.headers)
	}

	//Get list of existing records
	get(c.apiUrl+"/zones/"+zoneID+"/dns_records", nil, c.headers)

	fmt.Printf("Set %s to %s\n", c.host, c.ip)
	for _, v := range resp.Result {
		if strings.HasSuffix(c.host, v.Name) {
			zoneID = v.ID
			// unfortuanatly, there appears to be a delay or the result is cached.
			//fmt.Printf("Now set %s to %s\n", v.Name, v.Content)
		}
	}
}

func (c Cloudflare) getZoneID() string {
	resp, err := c.getZones()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return ""
	}
	zoneID := ""
	for _, v := range resp.Result {
		if strings.HasSuffix(c.host, v.Name) {
			zoneID = v.ID
		}
	}
	return zoneID
}
func (c Cloudflare) getZones() (ResponseStruct, error) {
	var response ResponseStruct

	//    if protocols.verbose {
	//        fmt.Println("testGet: /zones")
	//    }
	response, err := get(c.apiUrl+"/zones", nil, c.headers)
	if err != nil {
		return response, err
	}

	//    if protocols.verbose {
	//        fmt.Println("Unmarshaled as")
	//        fmt.Printf("%+v\n", response)
	//    }

	// TODO: find the matching result
	return response, nil
}

func callJSON(X, cmd string, options, headers map[string]string) (ResponseStruct, error) {
	body, err := protocols.CallJSON(X, cmd, options, headers)

	var response ResponseStruct
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}
	return response, err
}
func get(cmd string, options, headers map[string]string) (ResponseStruct, error) {
	body, err := protocols.Get(cmd, options, headers)

	var response ResponseStruct
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}
	return response, err
}

type ResultInfoStruct struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}
type ResultStruct struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Status  string `json:"status"`
}

// ResponseStruct - general purpose json response
type ResponseStruct struct {
	Success bool `json:"success"`
	// For now, there's enough overlap in the response json to use the same struct
	// later, may need to make an {}interface and then cast
	Result     []ResultStruct   `json:"result"`
	ResultInfo ResultInfoStruct `json:"result_info"`
	// Errors []interface
	// Messages []interface
}
