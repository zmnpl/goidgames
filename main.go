package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/tidwall/gjson"
)

const (
	API_URL       = "https://www.doomworld.com/idgames/api/api.php?"
	ACTION        = "action"
	ACTION_GET    = "get"
	ACTION_SEARCH = "search"
	ACTION_LATEST = "latestfiles"
	OUT           = "out"
	OUT_JSON      = "json"
	OUT_XML       = "xml"

	SEARCH_TYPE_FILENAME    = "filename"
	SEARCH_TYPE_TITLE       = "title"
	SEARCH_TYPE_AUTHOR      = "author"
	SEARCH_TYPE_EMAIL       = "email"
	SEARCH_TYPE_DESCRIPTION = "description"
	SEARCH_TYPE_CREDITS     = "credits"
	SEARCH_TYPE_EDITORS     = "editors"
	SEARCH_TYPE_TEXTFILE    = "textfile"

	SEARCH_SORT_DATE     = "date"
	SEARCH_SORT_FILENAME = "filename"
	SEARCH_SORT_SIZE     = "size"
	SEARCH_SORT_RATING   = "rating"

	SEARCH_SORT_ASC  = "asc"
	SEARCH_SORT_DESC = "desc"
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

// Get gets the data for a game specified by id or filepath.
// Pass an empyt string for not used paramters.
func Get(id int, filepath string) (g Idgame, err error) {
	u, _ := url.Parse(API_URL)
	q := u.Query()
	q.Set(ACTION, ACTION_GET)
	q.Set(OUT, OUT_JSON)

	if id > 0 {
		q.Set("id", fmt.Sprint(id))
	}

	if len(filepath) > 0 {
		q.Set("file", filepath)
	}

	u.RawQuery = q.Encode()

	responseData, err := getResponseData(u)
	if err != nil {
		return g, err
	}

	gameJson := gjson.Get(string(responseData), "content")
	json.Unmarshal([]byte(gameJson.String()), &g)

	reviews := gjson.Get(gameJson.String(), "reviews.review")
	json.Unmarshal([]byte(reviews.String()), &g.Reviews)

	return
}

// Search searches for games based on the query. It returns a slice with results. Types can be found as constant.
// searchType, sort and sortOrder can be one of the constants or an empty string. If empty, the API uses it's default.
func Search(query, searchType, sort, sortOrder string) (idgames []Idgame, err error) {
	if len(query) < 3 {
		return nil, fmt.Errorf("Query must at least be 3 characters.")
	}

	u, _ := url.Parse(API_URL)
	q := u.Query()
	q.Set(ACTION, ACTION_SEARCH)
	q.Set(OUT, OUT_JSON)
	q.Set("query", query)

	if len(searchType) > 0 {
		q.Set("type", searchType)
	}
	if len(sort) > 0 {
		q.Set("sort", sort)
	}
	if len(sortOrder) > 0 {
		q.Set("dir", sortOrder)
	}

	u.RawQuery = q.Encode()

	responseData, err := getResponseData(u)
	if err != nil {
		return nil, err
	}

	// try unmarshaling into slice assuming we get multiple here
	games := gjson.Get(string(responseData), "content.file")
	if err = json.Unmarshal([]byte(games.String()), &idgames); err != nil {
		switch err.(type) {
		default:
			return
		// if the error was an UnmarshalTypeError we think that there is only one result (not a slice but a scalar)
		// and unmarshal into just one object
		case *json.UnmarshalTypeError:
			idgames = make([]Idgame, 1)
			err = json.Unmarshal([]byte(games.String()), &idgames[0])
		}
	}
	return
}

// LatestFiles returns a slice of the latest additions to idgames. Limit the number or start from a specific Id. Pass 0 as startid if you want to see the hottest and newest files. If the limit is higher then the APIs max, then this is silently ignored.
func LatestFiles(limit, startid int) (idgames []Idgame, err error) {
	u, _ := url.Parse(API_URL)
	q := u.Query()
	q.Set(ACTION, ACTION_LATEST)
	q.Set(OUT, OUT_JSON)

	if limit > 0 {
		q.Set("limit", fmt.Sprint(limit))
	}
	if startid > 0 {
		q.Set("startid", fmt.Sprint(startid))
	}

	u.RawQuery = q.Encode()

	responseData, err := getResponseData(u)
	if err != nil {
		return nil, err
	}

	// try unmarshaling into slice assuming we get multiple here
	games := gjson.Get(string(responseData), "content.file")
	if err = json.Unmarshal([]byte(games.String()), &idgames); err != nil {
		switch err.(type) {
		default:
			return
		// if the error was an UnmarshalTypeError we think that there is only one result (not a slice but a scalar)
		// and unmarshal into just one object
		case *json.UnmarshalTypeError:
			idgames = make([]Idgame, 1)
			err = json.Unmarshal([]byte(games.String()), &idgames[0])
		}
	}
	return
}

func getResponseData(url *url.URL) ([]byte, error) {
	response, err := http.Get(url.String())
	if err != nil {
		return nil, fmt.Errorf("Could not connect to idgames: %s", err.Error())
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Could not understand idgame's response: %s", err.Error())
	}

	return responseData, nil
}
