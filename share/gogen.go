package share

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	config "github.com/coccyx/gogen/internal"
	"github.com/kr/pretty"
)

type GogenInfo struct {
	Gogen       string `json:"gogen"`
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
	SampleEvent string `json:"sampleEvent"`
	GistID      string `json:"gistID"`
}

// GogenList is returned by the /v1/list and /v1/search APIs for Gogen
type GogenList struct {
	Gogen       string
	Description string
}

// List calls /v1/list
func List() []GogenList {
	return listsearch("https://api.gogen.io/v1/list")

}

// Search calls /v1/search
func Search(q string) []GogenList {
	return listsearch("https://api.gogen.io/v1/search?q=" + url.QueryEscape(q))
}

func listsearch(url string) (ret []GogenList) {
	c := config.NewConfig()
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if resp.StatusCode != 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			c.Log.Fatalf("Non 200 response code searching for Gogen: %s", string(body))
		} else {
			c.Log.Fatalf("Error retrieving list of Gogens: %s", err)
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Log.Fatalf("Error reading body from response: %s", err)
	}
	var list map[string]interface{}
	err = json.Unmarshal(body, &list)
	// c.Log.Debugf("List body: %s", string(body))
	// c.Log.Debugf("list: %s", fmt.Sprintf("%# v", pretty.Formatter(list)))
	items := list["Items"].([]interface{})
	for _, item := range items {
		tempitem := item.(map[string]interface{})
		if _, ok := tempitem["gogen"]; !ok {
			continue
		}
		if _, ok := tempitem["description"]; !ok {
			continue
		}
		li := GogenList{Gogen: tempitem["gogen"].(string), Description: tempitem["description"].(string)}
		ret = append(ret, li)
	}
	c.Log.Debugf("List: %# v", pretty.Formatter(ret))
	return ret
}

// Get calls /v1/get
func Get(q string) (g GogenInfo) {
	c := config.NewConfig()
	client := &http.Client{}
	resp, err := client.Get("https://api.gogen.io/v1/get/" + q)
	if err != nil || resp.StatusCode != 200 {
		if resp.StatusCode != 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			c.Log.Fatalf("Non 200 response code retrieving Gogen: %s", string(body))
		} else {
			c.Log.Fatalf("Error retrieving Gogen %s: %s", q, err)
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Log.Fatalf("Error reading body from response: %s", err)
	}
	// c.Log.Debugf("Body: %s", body)
	var gogen map[string]interface{}
	err = json.Unmarshal(body, &gogen)
	if err != nil {
		c.Log.Fatalf("Error unmarshaling body: %s", err)
	}
	tmp, err := json.Marshal(gogen["Item"])
	if err != nil {
		c.Log.Fatalf("Error converting Item to JSON: %s", err)
	}
	err = json.Unmarshal(tmp, &g)
	if err != nil {
		c.Log.Fatalf("Error unmarshaling item: %s", err)
	}
	c.Log.Debugf("Gogen: %# v", pretty.Formatter(g))
	return g
}

// Upsert calls /v1/upsert
func Upsert(g GogenInfo) {
	c := config.NewConfig()
	gh := NewGitHub()
	client := &http.Client{}

	b, err := json.Marshal(g)
	if err != nil {
		c.Log.Fatalf("Error marshaling Gogen %#v: %s", g, err)
	}

	req, _ := http.NewRequest("POST", "https://api.gogen.io/v1/upsert", bytes.NewReader(b))
	req.Header.Add("Authorization", "token "+gh.token)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp.StatusCode != 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			c.Log.Fatalf("Non 200 response code Upserting: %s", string(body))
		} else {
			c.Log.Fatalf("Error POSTing to upsert: %s", err)
		}
	}
	c.Log.Debugf("Upserted: %# v", pretty.Formatter(g))
}
