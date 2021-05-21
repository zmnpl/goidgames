package goidgames

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// IdgamesBrowser holds all fields of the module
type IdgamesBrowser struct {
	app         *tview.Application
	layout      *tview.Grid
	list        *tview.Table
	fileDetails *tview.TextView
	reviews     *tview.TextView
	search      *tview.InputField
	idgames     []Idgame

	enterCallback func(idgamesurul string)
}

// NewIdgamesBrowser is the modules constructor
// Must be initialized with a *tview.Application in which it is drawn
func NewIdgamesBrowser(app *tview.Application) *IdgamesBrowser {
	browser := &IdgamesBrowser{app: app}

	layout := tview.NewGrid()
	layout.SetRows(5, -1, 5)
	layout.SetColumns(-1, -1)

	browser.layout = layout

	browser.initList()
	browser.initDetails()
	browser.initSearchForm()

	return browser
}

// SetEnterCallback sets a callback function that receives the idgames url of a row on which "ENTER" is pressed by the user
// This callbak function could, for example, launch a download of given file
func (b *IdgamesBrowser) SetEnterCallback(f func(idgamesurl string)) {
	b.enterCallback = f
}

// init search form ui component
func (b *IdgamesBrowser) initSearchForm() {
	searchForm := tview.NewForm()
	searchForm.SetHorizontal(true).SetBorder(true)

	search := tview.NewInputField().SetLabel("Search Idgames (leave empty for latest)").SetText("").SetFieldWidth(25)
	searchForm.AddFormItem(search)

	searchForm.AddButton("Search", func() {
		query := search.GetText()
		if len(query) == 0 {
			b.UpdateLatest()
		} else {
			types := []string{
				SEARCH_TYPE_TITLE,
				SEARCH_TYPE_AUTHOR,
			}
			b.UpdateSearch(search.GetText(), types)
		}
		b.app.SetFocus(b.list)
	})

	b.layout.AddItem(searchForm, 0, 0, 1, 2, 0, 0, true)

	b.search = search
}

// init details ui component
func (b *IdgamesBrowser) initDetails() {
	details := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)
	details.SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)

	b.layout.AddItem(details, 1, 1, 1, 1, 0, 0, false)

	details.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := event.Key()
		if k == tcell.KeyTAB {
			b.app.SetFocus(b.search)
			return nil
		}
		if k == tcell.KeyBacktab {
			b.app.SetFocus(b.list)
			return nil
		}
		return event
	})

	b.fileDetails = details
}

// init list ui component
func (b *IdgamesBrowser) initList() {
	list := tview.NewTable().
		SetFixed(1, 2).
		SetSelectable(true, false).
		SetBorders(false).SetSeparator('|')
	list.SetBorder(true)

	b.layout.AddItem(list, 1, 0, 1, 1, 0, 0, false)

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := event.Key()
		if k == tcell.KeyTAB {
			b.app.SetFocus(b.fileDetails)
			return nil
		}
		if k == tcell.KeyBacktab {
			b.app.SetFocus(b.search)
			return nil
		}
		return event
	})

	list.SetSelectedFunc(func(r int, c int) {
		switch {
		case r > 0:
			if b.enterCallback != nil {
				b.enterCallback(b.idgames[r-1].Idgamesurl)
			}
		}
	})

	b.list = list
}

// updateGameDetails iterates the given slice and fetches the detail data from Idgames via the api's get function
func updateGameDetails(idgames []Idgame) {
	for i := range idgames {
		g, err := Get(idgames[i].Id, "")
		if err != nil {
			continue
		}
		idgames[i] = g
	}
}

// UpdateSearch triggers an API call with given search query and types and populates the UI with the results
func (b *IdgamesBrowser) UpdateSearch(query string, types []string) {
	go func() {
		b.app.QueueUpdateDraw(func() {
			idgames, _ := SearchMultipleTypes(query, types, SEARCH_SORT_RATING, SEARCH_SORT_DESC)

			go func() {
				updateGameDetails(idgames)
			}()

			b.populateList(idgames)
		})
	}()
}

// UpdateLatest triggers an API call for the latest entries and populates the UI with the results
func (b *IdgamesBrowser) UpdateLatest() {
	go func() {
		b.app.QueueUpdateDraw(func() {
			idgames, _ := LatestFiles(50, 0)

			go func() {
				updateGameDetails(idgames)
			}()

			b.populateList(idgames)
		})
	}()
}

// populate the UIs list
func (b *IdgamesBrowser) populateList(idgames []Idgame) {
	b.list.Clear()
	b.idgames = idgames

	// header
	b.list.SetCell(0, 0, tview.NewTableCell("Rating").SetTextColor(tview.Styles.SecondaryTextColor))
	b.list.SetCell(0, 1, tview.NewTableCell("Title").SetTextColor(tview.Styles.SecondaryTextColor))
	b.list.SetCell(0, 2, tview.NewTableCell("Author").SetTextColor(tview.Styles.SecondaryTextColor))
	b.list.SetCell(0, 3, tview.NewTableCell("Date").SetTextColor(tview.Styles.SecondaryTextColor))

	b.list.SetSelectionChangedFunc(func(r int, c int) {
		switch r {
		case 0:
			return
		default:
			b.populateDetails(idgames[r-1])
		}
	})

	fixRows := 1
	cols := 4
	rows := len(idgames)
	for r := 1; r < rows+fixRows; r++ {
		var f Idgame
		if r > 0 {
			f = idgames[r-fixRows]
		}
		for c := 0; c < cols; c++ {
			var cell *tview.TableCell

			switch c {
			case 0:
				cell = tview.NewTableCell(ratingString(f.Rating)).SetTextColor(tview.Styles.PrimaryTextColor)
			case 1:
				cell = tview.NewTableCell(f.Title).SetTextColor(tview.Styles.PrimaryTextColor)
			case 2:
				cell = tview.NewTableCell(f.Author).SetTextColor(tview.Styles.PrimaryTextColor)
			case 3:
				cell = tview.NewTableCell(f.Date).SetTextColor(tview.Styles.PrimaryTextColor)
			default:
				cell = tview.NewTableCell("").SetTextColor(tview.Styles.PrimaryTextColor)
			}

			b.list.SetCell(r, c, cell)
		}
	}
	b.list.ScrollToBeginning()
}

// populate the detail pane
func (b *IdgamesBrowser) populateDetails(idgame Idgame) {
	b.fileDetails.Clear()
	fmt.Fprintf(b.fileDetails, "%s", idgame.Textfile)
}

// helper to make a string from the games rating
func ratingString(rating float32) string {
	return strings.Repeat("*", int(rating)) + strings.Repeat("-", 5-int(rating))
}
