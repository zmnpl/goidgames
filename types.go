package goidgames

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Idgame represents the metadata returned by the idgames api.
type Idgame struct {
	Id          int      `json:"id"`          // The file's id.
	Title       string   `json:"title"`       // The title of the file.
	Dir         string   `json:"dir"`         // The file's full directory path.
	Filename    string   `json:"filename"`    // The filename itself, no path.
	Size        int      `json:"size"`        // The size of the file in bytes.
	Age         int64    `json:"age"`         // The date that the file was added in seconds since the Unix Epoch (Jan. 1, 1970). Note: This is likely influenced by the time zone of the primary idGames Archive.
	Date        string   `json:"date"`        // A YYYY-MM-DD formatted date describing the date that this file was added to the archive.
	Author      string   `json:"author"`      // The file's author/uploader.
	Email       string   `json:"email"`       // The author's E-mail address.
	Description string   `json:"description"` // The file's description.
	Credits     string   `json:"credits"`     // The file's additional credits.
	Base        string   `json:"base"`        // The file's base (from another mod? made from scratch?).
	Buildtime   string   `json:"buildtime"`   // The file's/WAD's build time.
	Editors     string   `json:"editors"`     // The editors used to create this.
	Bugs        string   `json:"bugs"`        // Known bugs (if any).
	Textfile    string   `json:"textfile"`    // The file's text file contents.
	Rating      float32  `json:"rating"`      // The file's average rating, as rated by users.
	Votes       int      `json:"votes"`       // The number of votes that this file received.
	Url         string   `json:"url"`         // The URL for the idGames Archive page for this file.
	Idgamesurl  string   `json:"idgamesurl"`  // The idgames protocol URL for this file.
	Reviews     []Review `json:"reviews"`     // The element that contains all reviews for this file in review elements.
}

// Review represents a single review for one of the idgame files.
type Review struct {
	Text     string `json:"text"`     // The individual review's text, if any. Note: may be blank.
	Vote     int    `json:"vote"`     // The vote associated with the review.
	Username string `json:"username"` // The user name associated with the review, if any. Note: may be blank/null, which means "Anonymous". Since Version 3
}

type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %v complete", wc.Total)
}

// DownloadTo tries to download the game to given path and returns the full path of the downloaded file
func (g Idgame) DownloadTo(path string) (filePath string, err error) {
	success := false
	if err = os.MkdirAll(path, 0755); err != nil {
		return "", err
	}
	// try for all mirrors
	for _, mirror := range Mirrors {
		resp, err := http.Get(fmt.Sprintf("%s/%s/%s", mirror, g.Dir, g.Filename))
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		out, err := os.Create(filepath.Join(path, g.Filename))
		if err != nil {
			continue
		}
		defer out.Close()

		counter := &WriteCounter{}
		_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
		if err == nil {
			success = true
			break
		}
	}
	if !success {
		return "", fmt.Errorf("%s", "Unable to download.")
	}
	return filePath, nil
}
