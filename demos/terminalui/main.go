package main

import (
	"github.com/rivo/tview"
	"github.com/zmnpl/goidgames"
)

func main() {
	app := tview.NewApplication()
	idgamesbrowser := goidgames.NewIdgamesBrowser(app)
	if err := app.SetRoot(idgamesbrowser.GetRootLayout(), false).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
