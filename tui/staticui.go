package tui

import (
	"fmt"

	"github.com/rivo/tview"
)

const (
	subtitle  = "[orange]twad[white] - [orange]t[white]erminal [orange]wad[white] manager and launcher[orange]"
	subtitle2 = "twad - terminal wad manager and launcher"

	tviewHeader = "[orange]tview"
	creditTview = `The terminal user interface is build with tview:
https://github.com/rivo/tview`

	doomLogoCreditHeader = "[orange]DOOM Logo"
	creditDoomLogo       = `DOOM and Quake are registered trademarks of id Software, Inc. The DOOM, Quake and id logos are trademarks of id Software, Inc.

The ASCII version of the DOOM logo is Copyright © 1994 by F.P. de Vries.

This logo is work from Frans P. de Vries who originally made it and nicely granted me permission to use it here

Details can be found in this little piece of video game history:
http://www.gamers.org/~fpv/doomlogo.html`

	licenseHeader = "[orange]License"
	mitLicense    = `MIT License

Copyright (c) 2019 Simon Paul

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`
)

//  header
func makeHeader() *tview.TextView {
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)
	fmt.Fprintf(header, "%s", doomLogo)

	return header
}

//  license
func makeLicense() *tview.TextView {
	disclaimer := tview.NewTextView().SetDynamicColors(true).SetRegions(true)
	fmt.Fprintf(disclaimer, "%s\n", doomLogoCreditHeader)
	fmt.Fprintf(disclaimer, "%s\n\n", creditDoomLogo)
	fmt.Fprintf(disclaimer, "%s\n", tviewHeader)
	fmt.Fprintf(disclaimer, "%s\n\n", creditTview)
	fmt.Fprintf(disclaimer, "%s\n", licenseHeader)
	fmt.Fprintf(disclaimer, "%s", mitLicense)
	disclaimer.SetBorder(true).SetTitle("Credits / License")

	return disclaimer
}

// button bar showing keys
func makeButtonBar() *tview.Flex {
	btnHome := tview.NewButton("(ESC) Reset UI")
	btnRun := tview.NewButton("(Enter) Run Game")
	btnInsert := tview.NewButton("(i) Add Game")
	btnAddMod := tview.NewButton("(a) Add Mods To Game")
	btnRemoveMod := tview.NewButton("(r) Remove Last Mod From Game")
	btnDelete := tview.NewButton("(Delete) Remove Game")
	btnLicenseAndCredits := tview.NewButton("(c) Credits/License")
	btnQuit := tview.NewButton("(q) Quit")
	buttonBar := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(btnHome, 0, 1, false).
		AddItem(btnRun, 0, 1, false).
		AddItem(btnInsert, 0, 1, false).
		AddItem(btnAddMod, 0, 1, false).
		AddItem(btnRemoveMod, 0, 1, false).
		AddItem(btnDelete, 0, 1, false).
		AddItem(btnLicenseAndCredits, 0, 1, false).
		AddItem(btnQuit, 0, 1, false)

	return buttonBar
}

// help for navigation
func makeHelpPane() *tview.Flex {
	home := tview.NewTextView().SetDynamicColors(true).SetText(" [orange](ESC)[white]   - Reset UI")
	run := tview.NewTextView().SetDynamicColors(true).SetText(" [orange](Enter)[white] - Run Game")
	insert := tview.NewTextView().SetDynamicColors(true).SetText(" [orange](i)[white]     - Add Game")
	add := tview.NewTextView().SetDynamicColors(true).SetText(" [orange](a)[white]     - Add Mod To Game")
	remove := tview.NewTextView().SetDynamicColors(true).SetText(" [orange](r)[white]     - Remove Last Mod From Game")
	delet := tview.NewTextView().SetDynamicColors(true).SetText(" [orange](Del)[white]   - Remove Game")
	license := tview.NewTextView().SetDynamicColors(true).SetText(" [orange](c)[white]     - Credits/License")
	quit := tview.NewTextView().SetDynamicColors(true).SetText(" [orange](q)[white]     - Quit")

	spacer := tview.NewTextView().SetDynamicColors(true).SetText("")

	helpArea := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(spacer, 1, 0, false).
			AddItem(home, 1, 0, false).
			AddItem(run, 1, 0, false).
			AddItem(insert, 1, 0, false).
			AddItem(add, 1, 0, false).
			AddItem(spacer, 1, 0, false),
			0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(spacer, 1, 0, false).
			AddItem(remove, 1, 0, false).
			AddItem(delet, 1, 0, false).
			AddItem(license, 1, 0, false).
			AddItem(quit, 1, 0, false).
			AddItem(spacer, 1, 0, false),
			0, 1, false)
	helpArea.SetBorder(true)
	helpArea.SetTitle("Help")

	helpPage := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(helpArea, 8, 0, true)

	return helpPage
}