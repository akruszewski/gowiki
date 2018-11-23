package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

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

// Prepares test repository and returns its object.
func setupTestRepo() *git.Repository {
	repo, err := InitWiki()
	if err != nil {
		log.Fatalf("Can't init wiki repo %s", err)
	}
	for _, page := range testCases {
		fileName := WikiPath + page.title + ".wiki"
		_ = ioutil.WriteFile(
			fileName,
			[]byte(page.document),
			0600,
		)
		_, err := CommitFile(
			repo,
			fileName,
			fileName+" created.",
			"John Doe",
			"john@doe.com",
		)
		if err != nil {
			log.Fatalf("Can't commit file %s, %s", fileName, err)
		}
	}
	return repo
}

func tearDownTestRepo() {
	for _, page := range testCases {
		_ = os.Remove(WikiPath + page.title + ".wiki")
	}
	err := os.RemoveAll(WikiPath + ".git/")
	if err != nil {
		log.Fatalf("Can't remove git repository %s", WikiPath)
	}
}

// Test .save Page type method (for Page file creation)
func TestSavePage(t *testing.T) {
	for _, test := range testCases {
		tmp := Page{Title: test.title, Document: test.document}
		err := tmp.save()
		if err != nil {
			t.Errorf("%+v.save() - error occurs: %v", tmp, err)
		}
	}
	tearDownTestRepo()
}

// Test loadPage function (for Page file read)
func TestLoadPage(t *testing.T) {
	rep := setupTestRepo()
	defer tearDownTestRepo()
	for _, test := range testCases {
		tmp, err := loadPage(test.title, rep)
		if err != nil {
			t.Errorf("loadPage(\"%v\") - error occurs: %v", test.title, err)
		}
		if strings.Compare(tmp.Document, test.document) != 0 {
			t.Errorf(
				"loadPage(%v) - expected document: %s, found %s",
				test.title,
				test.document,
				tmp.Document,
			)
		}
	}
}

// Test .toJSON Page type method
func TestToJsonPage(t *testing.T) {
	rep := setupTestRepo()
	defer tearDownTestRepo()
	for _, test := range testCases {
		tmp, _ := loadPage(test.title, rep)
		_, err := tmp.toJSON()
		if err != nil {
			t.Errorf("Page%v(\"%v\") - error occurs: %v", tmp, test.title, err)
		}
	}
}

// Test .fromJSON Page type method
func TestFromJsonPage(t *testing.T) {
	rep := setupTestRepo()
	defer tearDownTestRepo()
	for _, test := range testCases {
		tmp, _ := loadPage(test.title, rep)
		err := tmp.fromJSON(test.json)
		if err != nil || strings.Compare(tmp.Document, test.document) != 0 {
			t.Errorf("Page%v(\"%v\") - error occurs: %v", tmp, test.title, err)
		}
	}
}

// Test removePage function (for page file remove)
func TestRemovePage(t *testing.T) {
	rep := setupTestRepo()
	defer tearDownTestRepo()
	for _, test := range testCases {
		err := removePage(test.title, rep)
		if err != nil {
			t.Errorf(
				"removePage(%v) != nil, error message: %v",
				test.title,
				err,
			)
		}
	}
}
