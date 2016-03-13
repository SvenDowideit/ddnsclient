
package cloudflare

import (
    "bytes"
    "fmt"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "strings"
    
    "github.com/SvenDowideit/ddnsclient/protocols"
)

// TODO: can I get the package name for this?
const driverName = "cloudflare"

func init() {
    fmt.Printf("Registering %s\n", driverName)
    protocols.RegisterDriver(driverName, New)
}

func New(host, ip, login, password string, verbose, debug bool) (protocols.Driver, error) {
    c := Cloudflare{
        host:host,
        ip:ip, 
        login:login,
        password:password,
        verbose:verbose,
        debug:debug,
        
        apiUrl: "https://api.cloudflare.com/client/v4",
        
        headers: map[string]string{
            "X-Auth-Email": login,
            "X-Auth-Key": password,
        },
    }
    
    return c, nil
}

type Cloudflare struct {
    host, ip, login, password string
    
    verbose, debug bool
    
    apiUrl string
    headers map[string]string
}

func (c Cloudflare) Set() {
    zoneID := c.getZoneID()
    if zoneID == "" {
        fmt.Printf("ERROR: %s not found in zones\n", c.host)
        return
    }

    //Get list of existing records
    resp, err := get(c.apiUrl+"/zones/"+zoneID+"/dns_records", nil, c.headers, c.verbose, c.debug)
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
            "type": "A",
            "name": c.host,
            "content": c.ip,
            "ttl": "120",
        }, c.headers, c.verbose, c.debug)
    } else {
        // modify entry
        _, err = callJSON("PUT", c.apiUrl+"/zones/"+zoneID+"/dns_records/"+recordID, map[string]string{
            "type": "A",
            "name": c.host,
            "content": c.ip,
            "ttl": "120",
        }, c.headers, c.verbose, c.debug)        
    }

    //Get list of existing records
    get(c.apiUrl+"/zones/"+zoneID+"/dns_records", nil, c.headers, c.verbose, c.debug)

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

    if c.verbose {
        fmt.Println("testGet: /zones")
    }
    response, err := get(c.apiUrl+"/zones", nil, c.headers, c.verbose, c.debug)
    if err != nil {
        return response, err
    }

    if c.verbose {
        fmt.Println("Unmarshaled as")
        fmt.Printf("%+v\n", response)
    }

    // TODO: find the matching result
    return response, nil
}

func callJSON(X, cmd string, options, headers map[string]string, verbose, debug bool) (ResponseStruct, error) {
    var response ResponseStruct

    if verbose {
        fmt.Println(X, ": ", cmd)
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
    
    req, err := http.NewRequest(X, cmd, buffer)

    if headers != nil {
        for k, v := range headers {
            req.Header.Add(k, v)
        }
    }

//    req.Header.Add("Content-Type", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
        return response, err
    }
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)

    if debug {
        var out bytes.Buffer
        json.Indent(&out, body, "", "  ")
        out.WriteTo(os.Stdout)
    }

    if verbose {
        fmt.Printf("output: %+v\n------\n", string(body))    
    }

	err = json.Unmarshal(body, &response)
    if err != nil {
        return response, err
    }
    return response, err
}


func get(cmd string, options, headers map[string]string, verbose, debug bool) (ResponseStruct, error) {
    u, err := url.Parse(cmd)
    if verbose {
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

    if headers != nil {
        for k, v := range headers {
            req.Header.Add(k, v)
        }
    }
    req.Header.Add("Content-Type", "application/json")

    var response ResponseStruct

    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
        return response, err
    }
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)

    if debug {
        var out bytes.Buffer
        json.Indent(&out, body, "", "  ")
        out.WriteTo(os.Stdout)
    }
    
    if verbose {
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
