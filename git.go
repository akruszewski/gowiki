package main

import (
	"log"
	"time"

	git "gopkg.in/src-d/go-git.v4"
	gitObject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Entry for wiki log functionality
type LogEntry struct {
	Commit  string    `json:"commit"`
	Message string    `json:"message"`
	Date    time.Time `json:"date"`
}

// Init git repo for wiki if doesn't exists.
func InitWiki() (*git.Repository, error) {
	repo, err := git.PlainInit(WikiPath, false)
	if err != nil {
		log.Println("Wiki already exists!")
		return nil, err
	}
	return repo, nil
}

// Get git repository for wiki.
func GetWiki(wikiPath string) (*git.Repository, error) {
	repo, err := git.PlainOpen(wikiPath)
	if err != nil {
		log.Println("Repository doesn't exists.!")
		return nil, err
	}
	return repo, nil
}

// Get log for given repository
func GetWikiLog(r *git.Repository) ([]LogEntry, error) {
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
	var wikiLog []LogEntry
	err = cIter.ForEach(func(c *gitObject.Commit) error {
		wikiLog = append(wikiLog, LogEntry{
			Commit:  c.Hash.String(),
			Message: c.Message,
			Date:    c.Author.When,
		})
		return nil
	})
	return wikiLog, nil
}

// Gets log for given file.
func GetFileLog(r *git.Repository, s string) ([]LogEntry, error) {
	ref, err := r.Head()
	if err != nil {
		log.Println("Can't fetch repository HEAD")
		return nil, err
	}
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash(), FileName: &s})
	if err != nil {
		log.Println("Can't get repository log for given file.")
		return nil, err
	}
	var lg []LogEntry
	cIter.ForEach(func(c *gitObject.Commit) error {
		lg = append(lg, LogEntry{
			Commit:  c.Hash.String(),
			Message: c.Message,
			Date:    c.Author.When,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return lg, nil
}

// Commits file with given fileName (without extension) to given repository
// with commit message as userName with email
func CommitFile(
	r *git.Repository,
	fileName string,
	message string,
	userName string,
	email string,
) (*LogEntry, error) {
	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}
	le := LogEntry{Message: message}
	w.Add(fileName + ".wiki")
	le.Date = time.Now()
	commit, err := w.Commit(message, &git.CommitOptions{
		Author: &gitObject.Signature{
			Name:  userName,
			Email: email,
			When:  le.Date,
		},
	})
	le.Commit = commit.String()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return &le, nil
}
