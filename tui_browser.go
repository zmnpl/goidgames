package goidgames

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	pageMain   = "main"
	pageDLSure = "dlsure"
)

// IdgamesBrowser holds all fields of the module
type IdgamesBrowser struct {
	app           *tview.Application
	canvas        *tview.Pages
	layout        *tview.Grid
	list          *tview.Table
	fileDetails   *tview.TextView
	reviews       *tview.TextView
	dlPathPreview *tview.TextView
	search        *tview.InputField
	idgames       []Idgame
	downloadPath  string

	confirmCallback      func(idgame Idgame)
	postDownloadCallback func(archivePath string)
}

// NewIdgamesBrowser is the modules constructor
// Must be initialized with a *tview.Application in which it is drawn
func NewIdgamesBrowser(app *tview.Application) *IdgamesBrowser {
	browser := &IdgamesBrowser{app: app}

	layout := tview.NewGrid()
	browser.layout = layout
	layout.SetRows(5, -1, 3)
	layout.SetColumns(-1, -1)

	canvas := tview.NewPages()
	canvas.AddPage(pageMain, layout, true, true)
	browser.canvas = canvas

	browser.initList()
	browser.initDetails()
	browser.initSearchForm()
	browser.initDlPathPreview()

	return browser
}

// SetConfirmCallback sets a callback function that receives the Idgame instance of a row on which "ENTER" is pressed by the user
// This callbak function could, for example, launch a download of given file
func (b *IdgamesBrowser) SetConfirmCallback(f func(idgame Idgame)) {
	b.confirmCallback = f
}

// SetDownloadPath sets the path where the browser can download game files to
func (b *IdgamesBrowser) SetDownloadPath(path string) {
	b.downloadPath = path
	b.populatedlPathPreview()
}

// SetDownloadDoneCallback set a callback function which gets exectuted when the download has been finished
// the callback receives the local file path of the downloaded archive
func (b *IdgamesBrowser) SetPostDownloadCallback(doWhenDownloadDone func(archivePath string)) {
	b.postDownloadCallback = doWhenDownloadDone
}

// GetRootLayout returns the root layout
func (b *IdgamesBrowser) GetRootLayout() *tview.Pages {
	return b.canvas
}

// GetRootLayout returns the root layout
func (b *IdgamesBrowser) GetSelectedRowNumber() int {
	r, _ := b.list.GetSelection()
	return r
}

// UpdateSearch triggers an API call with given search query and types and populates the UI with the results
func (browser *IdgamesBrowser) UpdateSearch(query string, types []string) {
	go func() {
		browser.app.QueueUpdateDraw(func() {
			idgames, _ := SearchMultipleTypes(query, types, SEARCH_SORT_RATING, SEARCH_SORT_DESC)

			go func() {
				updateGameDetails(idgames)
			}()

			browser.populateList(idgames)
		})
	}()
}

// UpdateLatest triggers an API call for the latest entries and populates the UI with the results
func (browser *IdgamesBrowser) UpdateLatest() {
	go func() {
		browser.app.QueueUpdateDraw(func() {
			idgames, _ := LatestFiles(50, 0)

			go func() {
				updateGameDetails(idgames)
			}()

			browser.populateList(idgames)
		})
	}()
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
			b.app.SetFocus(b.list)
			return nil
		}
		if k == tcell.KeyBacktab {
			b.app.SetFocus(b.search)
			return nil
		}
		return event
	})

	b.fileDetails = details
}

func (b *IdgamesBrowser) initDlPathPreview() {
	dlPathPreview := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)
	dlPathPreview.SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)
	b.layout.AddItem(dlPathPreview, 2, 0, 1, 2, 0, 0, false)

	b.dlPathPreview = dlPathPreview
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
		if r > 0 {
			g := b.idgames[r-1]

			// custom callback when enter is hit on a selection
			if b.confirmCallback != nil {
				b.confirmCallback(g)
			} else {
				// if there is no custom callback, a download is initiated
				b.canvas.AddPage(pageDLSure, sureDownloadBox(fmt.Sprintf("Download %v?", g.Title),
					func() {
						g.DownloadTo(b.downloadPath)
						b.canvas.RemovePage(pageDLSure)
						b.app.SetFocus(b.list)
					},
					func() {
						b.canvas.RemovePage(pageDLSure)
						b.app.SetFocus(b.list)
					},
					8,
					r+1,
					b.list.Box),
					true, true)
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

// populateList populates the UIs list
func (browser *IdgamesBrowser) populateList(idgames []Idgame) {
	browser.list.Clear()
	browser.idgames = idgames

	// header
	browser.list.SetCell(0, 0, tview.NewTableCell("Rating").SetTextColor(tview.Styles.SecondaryTextColor))
	browser.list.SetCell(0, 1, tview.NewTableCell("Title").SetTextColor(tview.Styles.SecondaryTextColor))
	browser.list.SetCell(0, 2, tview.NewTableCell("Author").SetTextColor(tview.Styles.SecondaryTextColor))
	browser.list.SetCell(0, 3, tview.NewTableCell("Date").SetTextColor(tview.Styles.SecondaryTextColor))

	browser.list.SetSelectionChangedFunc(func(r int, c int) {
		switch r {
		case 0:
			return
		default:
			browser.populateDetails(idgames[r-1])
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

			browser.list.SetCell(r, c, cell)
		}
	}
	browser.list.ScrollToBeginning()
}

// populate the detail panelayout
func (browser *IdgamesBrowser) populateDetails(idgame Idgame) {
	browser.fileDetails.Clear()

	// stylize the text file a bit
	re := regexp.MustCompile(`^(\S.*?):(.*)`)
	for _, line := range strings.Split(idgame.Textfile, "\n") {
		line = re.ReplaceAllString(line, fmt.Sprintf("%s$1:%s$2", hexStringFromColor(tview.Styles.MoreContrastBackgroundColor), hexStringFromColor(tview.Styles.PrimaryTextColor)))
		line = strings.Replace(line, "===========================================================================",
			hexStringFromColor(tview.Styles.MoreContrastBackgroundColor)+"==========================================================================="+hexStringFromColor(tview.Styles.PrimaryTextColor),
			1)

		fmt.Fprintf(browser.fileDetails, "%s\n", line)
	}

	browser.fileDetails.ScrollToBeginning()
}

// populate the detail panelayout
func (browser *IdgamesBrowser) populatedlPathPreview() {
	browser.dlPathPreview.Clear()
	fmt.Fprintf(browser.dlPathPreview, "%sDownload to:%s %s", hexStringFromColor(tview.Styles.MoreContrastBackgroundColor), hexStringFromColor(tview.Styles.PrimaryTextColor), browser.downloadPath)
}

// helper to make a string from the games rating
func ratingString(rating float32) string {
	return strings.Repeat("*", int(rating)) + strings.Repeat("-", 5-int(rating))
}

// help for navigation
func sureDownloadBox(title string, onOk func(), onCancel func(), xOffset int, yOffset int, container *tview.Box) *tview.Flex {
	okbtn := tview.NewButton("foo")
	okbtn.SetSelectedFunc(onOk)
	youSureForm := tview.NewForm()
	youSureForm.
		SetBorder(true).
		SetTitle(title)
	youSureForm.SetFocus(1)

	youSureForm.
		AddButton("Download", func() {
			youSureForm.GetButton(0).SetLabel("Hold on...")
			onOk()
		}).
		AddButton("Cancel", onCancel)

	height := 5
	width := 75

	// surrounding layout
	_, _, _, containerHeight := container.GetRect()
	helpHeight := 5

	// default: right below the selected game
	// though, if it flows out of the screen, then on top of the game
	if yOffset+height > containerHeight+helpHeight {
		yOffset = yOffset - height - 1
	}

	youSureLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, yOffset, 0, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, xOffset, 0, false).
			AddItem(youSureForm, width, 0, true).
			AddItem(nil, 0, 1, false),
			height, 1, true).
		AddItem(nil, 0, 1, false)

	return youSureLayout
}

func hexStringFromColor(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("[#%02x%02x%02x]", r, g, b)
}
