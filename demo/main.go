package main

import (
	"os"
	"path/filepath"

	"github.com/rivo/tview"
	"github.com/zmnpl/goidgames"
)

func main() {
	home, _ := os.UserHomeDir()

	app := tview.NewApplication()

	idgamesbrowser := goidgames.NewIdgamesBrowser(app)
	idgamesbrowser.SetDownloadPath(filepath.Join(home, "Downloads"))

	if err := app.SetRoot(idgamesbrowser.GetRootLayout(), true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
