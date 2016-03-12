
package main

import (
    "bytes"
    "flag"
    "fmt"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "strings"
)

var help = flag.Bool("help", false, "Show Help")

var login = flag.String("login", "", "login to log into your account")
var password = flag.String("password", "", "password to log into your account")
var host = flag.String("host", "test.demo.gallery", "host to inspect / modify")
var ip = flag.String("ip", "10.10.10.10", "IP address to set to")
var verbose = flag.Bool("verbose", false, "Verbose output")
var debug = flag.Bool("debug", false, "Debug output")

func main() {
    flag.Parse()
    
    if *help {
        flag.PrintDefaults()
        return
    }
    
    //testGet("/zones")
    // TODO: er, this is only the first zone - FIXME
    resp, err := getZones(*host)
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
        return
    }
    zoneID := ""
    for _, v := range resp.Result {
        if strings.HasSuffix(*host, v.Name) {
            zoneID = v.ID
        }
    }

    if zoneID == "" {
        fmt.Printf("ERROR: %s not found in zones\n", *host)
        return
    }

    //Get list of existing records
    resp, err = get("/zones/"+zoneID+"/dns_records", nil)
    if err != nil {
        fmt.Println("ERROR: ", err)
    }
    
    recordID := ""
    for _, v := range resp.Result {
        if v.Name == *host {
            recordID = v.ID
            fmt.Printf("Currently set %s to %s\n", v.Name, v.Content)
        }
    }
        
    if recordID == "" {
        // create a new entry
        _, err = post("/zones/"+zoneID+"/dns_records", map[string]string{
            "type": "A",
            "name": *host,
            "content": *ip,
            "ttl": "120",
        })
    } else {
        // modify entry
        _, err = put("/zones/"+zoneID+"/dns_records/"+recordID, map[string]string{
            "type": "A",
            "name": *host,
            "content": *ip,
            "ttl": "120",
        })        
    }

    //Get list of existing records
    get("/zones/"+zoneID+"/dns_records", nil)

    fmt.Printf("Set %s to %s\n", *host, *ip)
    for _, v := range resp.Result {
        if strings.HasSuffix(*host, v.Name) {
            zoneID = v.ID
            // unfortuanatly, there appears to be a delay or the result is cached.
            //fmt.Printf("Now set %s to %s\n", v.Name, v.Content)
        }
    }
}


func getZones(zoneName string) (ResponseStruct, error) {
    var response ResponseStruct

    if *verbose {
        fmt.Println("testGet: /zones")
    }
    response, err := get("/zones", nil)
    if err != nil {
        return response, err
    }

    if *verbose {
        fmt.Println("Unmarshaled as")
        fmt.Printf("%+v\n", response)
    }

    // TODO: find the matching result
    return response, nil
}

func oldgetZoneInfo() {
    fmt.Println("About to make request to https://www.cloudflare.com/api_json.html")
    // Get zone_id
    // curl https://www.cloudflare.com/api_json.html  \
    //          -d 'a=rec_load_all' \
    //          -d 'tkn=asdf' \
    //          -d 'email=svendowideit' \
    //          -d 'z=demo.gallery'
    resp, err := http.PostForm("https://www.cloudflare.com/api_json.html",
            url.Values{
                    "a": {"rec_load_all"}, 
                    "tkn": {*password},
                    "email": {*login},
                    "z": {"demo.gallery"},
                    })
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
        return
    }
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
        return
    }

    //var out bytes.Buffer
	//json.Indent(&out, body, "", "  ")
	//out.WriteTo(os.Stdout)
    
    type RequestInfo struct {
        Act string `json:"act"`
    }
    type Obj struct {
        RecID string `json:"rec_id"`
        ZoneName string `json:"zone_name"`
    }
    type RecsStruct struct {
        Objs []Obj `json:"objs"`
    }
    type ResponseInfo struct {
        HasMore bool `json:"has_more"`
        Count int `json:"count"`
        Recs RecsStruct `json:"recs"`
    }
    type CloudFlareRecLoadAll struct {
        Request RequestInfo
        Response ResponseInfo
    }
    var domains CloudFlareRecLoadAll
	err = json.Unmarshal(body, &domains)
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
        return
    }
    
    //fmt.Println("Unmarshaled as")
    //fmt.Printf("%+v\n", domains)
    
    fmt.Printf("\nZone Name: %s, Id: %s\n", 
        domains.Response.Recs.Objs[0].ZoneName, 
        domains.Response.Recs.Objs[0].RecID)
}

func put(cmd string, options map[string]string) (ResponseStruct, error) {
    return callJSON("PUT", cmd, options)
}

func post(cmd string, options map[string]string) (ResponseStruct, error) {
    return callJSON("POST", cmd, options)
}
func callJSON(X, cmd string, options map[string]string) (ResponseStruct, error) {
    var response ResponseStruct

    if *verbose {
        fmt.Println(X, ": https://api.cloudflare.com/client/v4", cmd)
    }
    client := &http.Client{    }

    var buffer *bytes.Buffer
    if options != nil {
        buffer = new(bytes.Buffer)
        err := json.NewEncoder(buffer).Encode(options)
        if err != nil {
            return response, err
        }
    }
    
    req, err := http.NewRequest(X, "https://api.cloudflare.com/client/v4"+cmd, buffer)

    req.Header.Add("X-Auth-Email", *login)
    req.Header.Add("X-Auth-Key", *password)
//    req.Header.Add("Content-Type", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
        return response, err
    }
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)

    if *debug {
        var out bytes.Buffer
        json.Indent(&out, body, "", "  ")
        out.WriteTo(os.Stdout)
    }

    if *verbose {
        fmt.Printf("output: %+v\n------\n", string(body))    
    }

	err = json.Unmarshal(body, &response)
    if err != nil {
        return response, err
    }
    return response, err
}


func get(cmd string, options map[string]string) (ResponseStruct, error) {
    u, err := url.Parse("https://api.cloudflare.com/client/v4"+cmd)
    if *verbose {
        fmt.Println("Get: ", u.String())
    }
    client := &http.Client{    }
    
    if options != nil {
        params := url.Values{}
        for k, v := range options {
            params.Add(k, v)
        }
        u.RawQuery = params.Encode()
    }
    req, err := http.NewRequest("GET", u.String(), nil)

    req.Header.Add("X-Auth-Email", *login)
    req.Header.Add("X-Auth-Key", *password)
    req.Header.Add("Content-Type", "application/json")

    var response ResponseStruct

    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
        return response, err
    }
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)

    if *debug {
        var out bytes.Buffer
        json.Indent(&out, body, "", "  ")
        out.WriteTo(os.Stdout)
    }
    
    if *verbose {
        fmt.Printf("output: %+v\n------\n", string(body))    
    }

	err = json.Unmarshal(body, &response)
    if err != nil {
        return response, err
    }
    return response, err
}

type ResultInfoStruct struct {
    Page int `json:"page"`
    PerPage int `json:"per_page"`
    TotalPages int `json:"total_pages"`
    Count int `json:"count"`
    TotalCount int `json:"total_count"`
}
type ResultStruct struct {
    ID string `json:"id"`
    Type string `json:"type"`
    Name string `json:"name"`
    Content string `json:"content"`
    TTL int `json:"ttl"`
    Status string `json:"status"`
}
// ResponseStruct - general purpose json response
type ResponseStruct struct {
    Success bool  `json:"success"`
    // For now, there's enough overlap in the response json to use the same struct
    // later, may need to make an {}interface and then cast
    Result []ResultStruct `json:"result"`
    ResultInfo ResultInfoStruct `json:"result_info"`
    // Errors []interface
    // Messages []interface
}
