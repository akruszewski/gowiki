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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/urfave/cli"
	git "gopkg.in/src-d/go-git.v4"
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
	Title    string     `json:"title"`
	Document string     `json:"document"`
	Updated  time.Time  `json:"updated"`
	Message  string     `json:"message"`
	Log      []LogEntry `json:"log"`
}

// Saves Document file in WikiPath with Page Title and '.wiki' suffix
func (p *Page) save() error {
	return ioutil.WriteFile(
		WikiPath+p.Title+".wiki",
		[]byte(p.Document),
		0600,
	)
}

// Commits Page to repo
func (p *Page) commit(mssg string, r *git.Repository) error {
	logEntry, err := CommitFile(
		r,
		p.Title,
		mssg,
		os.Getenv("GIT_USERNAME"),
		os.Getenv("GIT_EMAIL"),
	)
	if err != nil {
		return err
	}
	p.Log = append(p.Log, *logEntry)
	return nil
}

// Load Page file commits log
func (p *Page) loadLog(r *git.Repository) error {
	lg, err := GetFileLog(r, p.Title+".wiki")
	if err != nil {
		log.Printf("Can't load commits log for file %s", WikiPath+p.Title)
		return err
	}
	p.Log = append(p.Log, lg...)
	return nil
}

// Returns []bytes with Document in json format.
func (p *Page) toJSON() ([]byte, error) {
	js, err := json.Marshal(p)
	if err != nil {
		log.Printf("Page %v can't be decoded:  %v", p.Title, err)
		return nil, err
	}
	return js, nil
}

// Creates Page Document from given []bytes contains JSON serializable data.
func (p *Page) fromJSON(data []byte) error {
	err := json.Unmarshal(data, &p)
	if err != nil {
		log.Printf("Error decoding body: %v", err)
		return err
	}
	return nil
}

// Returns Page type variable loaded from WikiPath with given title.
func loadPage(title string, r *git.Repository) (*Page, error) {
	filePath := WikiPath + title + ".wiki"
	body, err := ioutil.ReadFile(filePath)
	page := Page{Title: title, Document: string(body)}
	page.loadLog(r)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// Removes Page Document file with given title.
func removePage(title string, r *git.Repository) error {
	err := os.Remove(WikiPath + title + ".wiki")
	if err != nil {
		return err
	}
	_, err = CommitFile(
		r,
		title,
		fmt.Sprintf("Wiki page %s removed.", title),
		os.Getenv("GIT_USERNAME"),
		os.Getenv("GIT_EMAIL"),
	)
	return err
}

// check if env vars are setup correctly, otherwise set defaults
func init() {
	if WikiPath == "" {
		WikiPath = "/Users/adriankruszewski/tmp/"
	}
}

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Init Wiki git repository.",
			Action: func(c *cli.Context) error {
				log.Print("Initializing git repo in ", WikiPath)
				_, err := InitWiki()
				if err != nil {
					os.Exit(-1)
				}
				return nil
			},
		},
		{
			Name:    "runserver",
			Aliases: []string{"r"},
			Usage:   "Run Wiki server.",
			Action: func(c *cli.Context) error {
				RunServer()
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
