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
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/cli"
	git "gopkg.in/src-d/go-git.v4"
	gitObject "gopkg.in/src-d/go-git.v4/plumbing/object"
	//    "gopkg.in/yaml.v2"
)

// TODO: move it to separate yaml file
var WikiPath = os.Getenv("WIKIPATH")

type Settings struct {
	GitUrl    string `json:"git_url"`
	PublicKey string `json:"public_key"`
}

// Entry for wiki log functionality
type WikiLogEntry struct {
	Commit  string    `json:"commit"`
	Message string    `json:"message"`
	Date    time.Time `json:"date"`
}

// Page is singe wiki page consisting of Title and Document
type Page struct {
	Title    string         `json:"title"`
	Document string         `json:"document"`
	Updated  time.Time      `json:"updated"`
	Message  string         `json:"message"`
	Log      []WikiLogEntry `json:"log"`
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
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	w.Add(p.Title + ".wiki")
	p.Updated = time.Now()
	commit, err := w.Commit(mssg, &git.CommitOptions{
		Author: &gitObject.Signature{
			Name:  "John Doe",     //TODO: fix name to gets real data
			Email: "john@doe.org", //TODO: fix email to gets real data
			When:  p.Updated,
		},
	})
	if err != nil {
		log.Print(err)
		return err
	}
	p.Log = append(p.Log, WikiLogEntry{
		Commit:  commit.String(),
		Message: mssg,
		Date:    p.Updated},
	)
	return nil
}

// Load Page file commits log
func (p *Page) loadLog(r *git.Repository) error {
	cIter, err := GetLogForFileWikiRepo(r, p.Title+".wiki")
	if err != nil {
		log.Printf("Can't load commits log for file %s", WikiPath+p.Title)
		return err
	}
	err = cIter.ForEach(func(c *gitObject.Commit) error {
		p.Log = append(p.Log, WikiLogEntry{
			Commit:  c.Hash.String(),
			Message: c.Message,
			Date:    c.Author.When,
		})
		return nil
	})
	if err != nil {
		return err
	}
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
func removePage(title string) error {
	return os.Remove(WikiPath + title + ".wiki")
}

// Init git repo for wiki if doesn't exists.
func InitWikiRepo() (*git.Repository, error) {
	repo, err := git.PlainInit(WikiPath, false)
	if err != nil {
		log.Println("Wiki already exists!")
		return nil, err
	}
	return repo, nil
}

// Get git repository for wiki.
func GetWikiRepo() (*git.Repository, error) {
	repo, err := git.PlainOpen(WikiPath)
	if err != nil {
		log.Println("Repository doesn't exists.!")
		return nil, err
	}
	return repo, nil
}

// Get log for given repository
func GetLogWikiRepo(r *git.Repository) (gitObject.CommitIter, error) {
	ref, err := r.Head()
	if err != nil {
		log.Println("Can't fetch repository HEAD")
		return nil, err
	}
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		log.Println("Can't get repository log")
		return nil, err
	}
	return cIter, nil
}

// Get log for given repository for given file
func GetLogForFileWikiRepo(r *git.Repository, s string) (gitObject.CommitIter, error) {
	ref, err := r.Head()
	if err != nil {
		log.Println("Can't fetch repository HEAD")
		return nil, err
	}
	log.Print(s)
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash(), FileName: &s})
	if err != nil {
		log.Println("Can't get repository log for given file.")
		return nil, err
	}
	return cIter, nil
}

func RunServer() {
	r := httprouter.New()

	r.GET("/api/wiki_log/", LogHandler)
	r.GET("/api/wiki/:title/log/", PageLogHandler)
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

// check if env vars are setup corectly, otherwise set defaults
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
				_, err := InitWikiRepo()
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

func LogHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rep, err := GetWikiRepo()
	if err != nil {
		http.Error(rw, "Wiki isn't configured!", http.StatusInternalServerError)
		return
	}
	log, err := GetLogWikiRepo(rep)
	if err != nil {
		http.Error(rw, "Can't fetch Wiki log", http.StatusInternalServerError)
	}

	var res []WikiLogEntry
	err = log.ForEach(func(c *gitObject.Commit) error {
		res = append(res, WikiLogEntry{
			Commit:  c.Hash.String(),
			Message: c.Message,
			Date:    c.Author.When,
		})
		return nil
	})
	js, err := json.Marshal(res)
	rw.Write(js)
}

func PageLogHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rep, err := GetWikiRepo()
	if err != nil {
		http.Error(rw, "Wiki isn't configured!", http.StatusInternalServerError)
		return
	}
	lg, err := GetLogForFileWikiRepo(rep, p.ByName("title")+".wiki")
	if err != nil {
		http.Error(rw, "Can't fetch log for Page.", http.StatusInternalServerError)
	}

	var res []WikiLogEntry
	err = lg.ForEach(func(c *gitObject.Commit) error {
		res = append(res, WikiLogEntry{
			Commit:  c.Hash.String(),
			Message: c.Message,
			Date:    c.Author.When,
		})
		return nil
	})
	js, err := json.Marshal(res)
	rw.Write(js)
}

//
//func PageCommitHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
//}
//

// Http view handler for retrieving wiki page
func PageGetHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	title := p.ByName("title")
	rep, err := GetWikiRepo()
	if err != nil {
		http.Error(rw, "Wiki improperly configured.", http.StatusInternalServerError)
	}
	page, err := loadPage(title, rep)
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

	rep, err := GetWikiRepo()
	if err != nil {
		http.Error(rw, "Wiki improperly configured.", http.StatusInternalServerError)
	}
	page.save()
	page.commit(page.Message, rep)
	page.loadLog(rep)

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
