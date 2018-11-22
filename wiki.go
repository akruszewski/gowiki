/* Application Personal Wiki

web API:

	url: /api/wiki/
	description: get wiki pages list.
	authentication headers: required
	method: GET
	response:
	[{
		"title": string, // title used in urls; name of file
		"updated": datetime,
	}, ...]

	url: /api/wiki/<page_title>/
	description: get wiki page.
	authentication headers: required
	method: GET
	response:
	{
		"document": string,
		"updated": datetime,
	}

	url: /api/wiki/<page_title>/log/
	description: get wiki page git log list.
	authentication headers: required
	method: GET
	response:
	[{
		"commit_id": string,
		"commit_message": datetime,
		"updated": datetime,
	}, ...]

	url: /api/wiki/<page_title>/<commit_id>/
	description: get wiki page for given git commit.
	authentication headers: required
	method: GET
	response:
	{
		"document": string,
		"updated": datetime,
	}


	url: /api/wiki/<page_title>/
	description: create or update wiki page
	authentication headers: required
	method: POST
	request:
	{
		"document": string // required,
		"commit_message": string

	}
	response:
	{
		"document": string,
		"created": datetime,
		"updated": datetime,
	}

	url: /api/wiki/<page_title>/
	description: delete wiki page
	authentication headers: required
	method: DELETE

	url: /api/settings/
	description: get settings of wiki
	authentication headers: required
	method: GET
	response:
	{
		"git_url": string,
		"public_key": string
	}

	url: /api/settings/
	description: update wiki settings
	authentication headers: required
	method: POST
	request:
	{
		"git_url": string
	}
	response:
	{
		"git_url": string,
		"public_key": string
	}

	url: /api/sync_wiki/
	description: sync wiki with git repo.
	authentication headers: required
	method: POST
	response:
	{
		"message": string,
		"user": string,
		"password": string
	}

*/
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
	//    "gopkg.in/yaml.v2"
)

// TODO: move it to separate yaml file
var WikiPath = os.Getenv("WIKIPATH")

type Settings struct {
	GitUrl    string `json:"git_url"`
	PublicKey string `json:"public_key"`
}

// Page is singe wiki page consisting of Title and Document
type Page struct {
	Title    string `json:"title"`
	Document string `json:"document"`
	Updated  string `json:"updated"`
}

// Saves Document file in WikiPath with Page Title and '.wiki' suffix
func (p *Page) save() error {
	return ioutil.WriteFile(
		WikiPath+p.Title+".wiki",
		[]byte(p.Document),
		0600,
	)
}

//func (p *Page) commit() error {
//	return nil
//}
//
//func (p *Page) log() error {
//	return nil
//}

// Returns []bytes with Document in json format.
func (p *Page) toJSON() ([]byte, error) {
	js, err := json.Marshal(p)
	if err != nil {
		log.Printf("Page %v can't be decoded:  %v", p.Title, err)
		return nil, err
	}
	return js, nil
}

// Creates Page Docuemnt from given []bytes contains JSON serializable data.
func (p *Page) fromJSON(data []byte) error {
	err := json.Unmarshal(data, &p)
	if err != nil {
		log.Printf("Error decoding body: %v", err)
		return err
	}
	return nil
}

// Returns Page type variable loaded from WikiPath with given title.
func loadPage(title string) (*Page, error) {
	body, err := ioutil.ReadFile(WikiPath + title + ".wiki")
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Document: string(body)}, nil
}

// Removes Page Document file with given title.
func removePage(title string) error {
	return os.Remove(WikiPath + title + ".wiki")
}

/*
TODO:
    add initWiki function
    add listWiki function
    add cloneWiki function
    add syncWiki function
    add generatePrivateKey function
    add getPublicKey function
*/

// check if env vars are setup corectly, otherwise set defaults
func init() {
	if WikiPath == "" {
		WikiPath = "/Users/adriankruszewski/tmp/"
	}
}

func main() {
	r := httprouter.New()

	//    r.GET("/api/wiki/:title/log/", PageLogHandler)
	//    r.GET("/api/wiki/:title/:commit_id", PageCommitHandler)

	r.GET("/api/wiki/:title", PageGetHandler)
	r.POST("/api/wiki/:title", PageUpdateHandler)
	r.DELETE("/api/wiki/:title", PageDeleteHandler)

	r.GET("/api/wiki/", PageListHandler)
	//
	//    r.GET("/api/settings/", SettingsGetHandler)
	//    r.POST("/api/settings/", SettingsUpdateHandler)
	//
	//    r.GET("/api/sync_wiki/", SyncWikiHandler)

	log.Println("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}

//func PageLogHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
//}
//
//func PageCommitHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
//}
//

// Http view handler for retrieving wiki page
func PageGetHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	title := p.ByName("title")
	page, err := loadPage(title)
	if err != nil {
		http.Error(rw, "Page not found.", http.StatusNotFound)
		return
	}
	js, err := page.toJSON()
	if err != nil {
		http.Error(rw, "Decoding error", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)
}

// Http view handler for updating wiki page
func PageUpdateHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	title := p.ByName("title")
	reqBody, err := ioutil.ReadAll(r.Body)
	page := Page{Title: title}

	err = page.fromJSON(reqBody)
	if err != nil {
		http.Error(rw, "Decoding error", http.StatusInternalServerError)
		return
	}
	page.save()

	js, err := page.toJSON()
	if err != nil {
		http.Error(rw, "Encoding error", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)
}

// Http view handler for deleting wiki page
func PageDeleteHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := removePage(p.ByName("title"))
	if err != nil {
		log.Printf("Error deleting page: %v", err)
		http.Error(rw, "can't delete page", http.StatusBadRequest)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write([]byte("{\"message\": \"Page removed\"}"))
}

// Http view handler for listing wiki pages
func PageListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	files, err := ioutil.ReadDir(WikiPath)
	var pages []string
	if err != nil {
		log.Printf("Error listing pages: %v", err)
		http.Error(rw, "Can't load page list.", http.StatusInternalServerError)
	}
	// TODO: figure out why this range didn't work
	for _, file := range files {
		fileName := file.Name()
		if strings.HasSuffix(fileName, ".wiki") {
			pages = append(pages, strings.TrimSuffix(fileName, ".wiki"))
		}
	}
	js, err := json.Marshal(pages)
	if err != nil {
		log.Printf("Error marshaling pages: %v", err)
		http.Error(rw, "Can't load page list.", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)
}

//
//func SettingsGetHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
//}
//
//func SettingsUpdateHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
//}
//
//func SyncWikiHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
//}
