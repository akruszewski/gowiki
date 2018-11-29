package webservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	page "github.com/akruszewski/awiki/page/git"
	"github.com/julienschmidt/httprouter"
	git "gopkg.in/src-d/go-git.v4"
)

var testCases = []struct {
	title    string
	document string
	json     []byte
}{
	{
		"test_index",
		"index",
		[]byte(`{"title":"test_index","document":"index"}`),
	},
	{
		"test_index 2",
		"index2",
		[]byte(`{"title":"test_index 2","document":"index2"}`),
	},
	{
		"test_index.3",
		"index3",
		[]byte(`{"title":"test_index.3","document":"index3"}`),
	},
}

// Prepares dir for repo and returns its path.
func prepareTestDir() (string, error) {
	dir, err := ioutil.TempDir("", "tmp_wiki")
	if err != nil {
		log.Fatal(err)
		return dir, err
	}
	return dir, nil
}

// Prepares test repository and returns its object.
func setupTestRepoWithFiles() (*git.Repository, string) {
	dir, _ := prepareTestDir()
	repo, err := page.Init(dir)
	if err != nil {
		log.Fatalf("Can't init wiki repo %s", err)
	}
	for _, pg := range testCases {
		fileName := path.Join(dir, pg.title+".wiki")
		_ = ioutil.WriteFile(
			fileName,
			[]byte(pg.document),
			0600,
		)
		_, err := page.CommitFile(
			repo,
			fileName,
			fmt.Sprint(fileName, " created."),
			"John Doe",
			"john@doe.com",
		)
		if err != nil {
			log.Fatalf("Can't commit file %s, %s", fileName, err)
		}
	}
	return repo, dir
}

// Removes tmp test directory
func tearDownTestRepo(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatalf("Can't remove git repository %s", dir)
	}
}

//Test pageUpdate http view function.
func TestPageUpdate(t *testing.T) {
	_, WikiPath := setupTestRepoWithFiles()
	defer tearDownTestRepo(WikiPath)

	//Test creating new page.
	for _, test := range testCases {
		ps := httprouter.Params{httprouter.Param{
			Key:   "title",
			Value: test.title,
		}}
		req, err := http.NewRequest(
			"POST",
			fmt.Sprintf("http://localhost:8080/%s", test.title),
			bytes.NewBuffer(test.json),
		)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		func() {
			WikiPath = WikiPath
			pageUpdate(w, req, ps)
		}()
		if w.Code != 200 {
			t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
		}
		pg := page.Page{}
		err = json.Unmarshal(w.Body.Bytes(), &pg)
		if err != nil {
			t.Error(err)
		}
		if !strings.Contains(pg.Document, test.document) {
			t.Error("Response body is different from expected one!")
		}
	}

	//Test updating new page.
	ps := httprouter.Params{httprouter.Param{
		Key:   "title",
		Value: testCases[0].title,
	}}
	newDocument := []byte(`{"document":"asd"}`)
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://localhost:8080/%s", testCases[0].title),
		bytes.NewBuffer(newDocument),
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	func() {
		WikiPath = WikiPath
		pageUpdate(w, req, ps)
	}()
	if w.Code != 200 {
		t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
	}

	pg := page.Page{}
	err = json.Unmarshal(w.Body.Bytes(), &pg)
	if err != nil {
		t.Errorf("Error occures: %s.", err)
	}

	if !strings.Contains(
		pg.Document,
		"asd",
	) {
		t.Errorf("Response body is different from expected one! %s", pg)
	}
}

//Test PageGetHandler http view function.
func TestPageGetHandler(t *testing.T) {
	_, WikiPath := setupTestRepoWithFiles()
	defer tearDownTestRepo(WikiPath)

	// Gets existing page.
	for _, test := range testCases {
		ps := httprouter.Params{httprouter.Param{
			Key:   "title",
			Value: test.title,
		}}
		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf("http://localhost:8080/%s", test.title),
			nil,
		)
		if err != nil {
			log.Fatal(err)
		}
		w := httptest.NewRecorder()

		func() {
			WikiPath = WikiPath
			pageGet(w, req, ps)
		}()
		if w.Code != 200 {
			t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
		}

		pg := page.Page{}
		err = json.Unmarshal(w.Body.Bytes(), &pg)
		if err != nil {
			t.Errorf("Error occures: %s.", err)
		}

		if !strings.Contains(pg.Document, test.document) {
			t.Error("Response document body is different from expected one!", pg.Document)
		}
	}

	// Gets not existing page.
	req, err := http.NewRequest(
		"GET",
		"http://localhost:8080/test_not_existing",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	w := httptest.NewRecorder()

	pageGet(
		w,
		req,
		httprouter.Params{httprouter.Param{
			Key:   "title",
			Value: "test_not_existing",
		}},
	)
	if w.Code != 404 {
		t.Errorf("Expected 404 HTTP status code, gets %d", w.Code)
	}
}

//Test PageListHandler http view function.
func TestPageListHandler(t *testing.T) {
	_, WikiPath := setupTestRepoWithFiles()
	defer tearDownTestRepo(WikiPath)
	req, err := http.NewRequest(
		"GET",
		"http://localhost:8080/",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	w := httptest.NewRecorder()
	func() {
		WikiPath = WikiPath
		pageList(w, req, httprouter.Params{})
	}()
	if w.Code != 200 {
		t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
	}
	bd := w.Body.Bytes()
	if !bytes.Contains(
		bd,
		[]byte(`["test_index 2","test_index.3","test_index"]`),
	) {
		t.Errorf("Response body is different (%s) from expected one!", bd)
	}
}

//Test PageDeleteHandler http view function.
func TestPageDeleteHandler(t *testing.T) {
	_, WikiPath := setupTestRepoWithFiles()
	defer tearDownTestRepo(WikiPath)
	// Deletes existing page.
	for _, test := range testCases {
		ps := httprouter.Params{httprouter.Param{
			Key:   "title",
			Value: test.title,
		}}
		req, err := http.NewRequest(
			"DELETE",
			fmt.Sprintf("http://localhost:8080/%s", test.title),
			nil,
		)
		if err != nil {
			log.Fatal(err)
		}

		w := httptest.NewRecorder()
		func() {
			WikiPath = WikiPath
			pageDelete(w, req, ps)
		}()
		if w.Code != 200 {
			t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
		}
		if !bytes.Contains(
			w.Body.Bytes(),
			[]byte("{\"message\": \"Page removed\"}"),
		) {
			t.Error("Response body is different from expected one!")
		}
	}
}

// Test LogHandler http view function
func TestLogHandler(t *testing.T) {
	_, WikiPath := setupTestRepoWithFiles()
	defer tearDownTestRepo(WikiPath)

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("http://localhost:8080/wiki_log/"),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	func() {
		WikiPath = WikiPath
		wikiLog(w, req, httprouter.Params{})
	}()
	if w.Code != 200 {
		t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
	}

	var ll page.Log
	err = json.Unmarshal(w.Body.Bytes(), &ll)

	//Log needs to be reversed to reflect testCases order
	if err != nil {
		t.Error("Can't encode response!")
	}
	llLen := len(ll)
	for i, test := range testCases {
		if !strings.Contains(ll[llLen-i-1].Message, test.title) {
			t.Error(
				"Incorrect log message!",
				test.title,
				ll[llLen-i-1].Message,
			)
		}
	}
}

// Test PageLogHandler http view function
func TestPageLogHandler(t *testing.T) {
	_, WikiPath := setupTestRepoWithFiles()
	defer tearDownTestRepo(WikiPath)

	for _, test := range testCases {
		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf("http://localhost:8080/wiki_log/"),
			nil,
		)
		if err != nil {
			log.Fatal(err)
		}

		w := httptest.NewRecorder()
		func() {
			WikiPath = WikiPath
			pageLog(w, req, httprouter.Params{httprouter.Param{
				Key:   "title",
				Value: test.title,
			}})
		}()
		if w.Code != 200 {
			t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
		}
		t.Error(string(w.Body.Bytes()))
	}
}
