package tui

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/zmnpl/twad/cfg"
)

const (
	setupOkHint = "Hit [red]Ctrl+O[white] when you are done."

	setupPathExplain = `For [orange]twad[white] to function correctly, your *.wad files /  DOOM mod files need to be organized in one central directory. Put your doom.wad and doom2.wad in that central directory and create subdirectories per mod for the respective files. This folder will be set as [red]DOOMWADDIR[white] environment variable for the current terminal session when [orange]twad[white] runs.

Navigate with arrow keys or Vim bindings. [red]Enter[white] or [red]Space[white] expand the directory. Highlight the righ one and hit [red]Ctrl+O[white]`

	setupPathExample = `[red]->[white]/home/slayer/doomwaddir            [red]# put doom.wad and doom2.wad in here
  [white]/home/slayer/doomwaddir[orange]/BrutalDoom [grey]# sub dir for Sigil
  [white]/home/slayer/doomwaddir[orange]/QCDE       [grey]# sub dir for QCDE`
)

// settings page
func makeFirstTimeSetup() *tview.Flex {
	basePathPreview := tview.NewTextView()
	basePathPreview.SetBackgroundColor(previewBackgroundColor)
	fmt.Fprintf(basePathPreview, "mods path: ")
	pathSelector := makePathSelectionTree(basePathPreview)

	explanation := tview.NewTextView().SetRegions(true).SetWrap(true).SetWordWrap(true).SetDynamicColors(true)
	fmt.Fprintf(explanation, "%s\n\nPoint me to the highlighted directory:\n", setupPathExplain)
	fmt.Fprintf(explanation, "%s", setupPathExample)
	fmt.Fprintf(explanation, "\n\n%s", setupOkHint)

	settingsFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	settingsFlex.SetBorder(true)
	settingsFlex.SetTitle("Setup")
	settingsFlex.SetBorderColor(accentColor)
	settingsFlex.SetTitleColor(accentColor)

	settingsPage := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(settingsFlex, 90, 0, true).
		AddItem(tview.NewBox().SetBorder(false), 0, 1, false)

	settingsFlex.AddItem(explanation, 12, 0, false).
		AddItem(basePathPreview, 1, 0, false).
		AddItem(pathSelector, 0, 1, true)

	return settingsPage
}

// tree view for selecting additional mods TODO
func makePathSelectionTree(preview *tview.TextView) *tview.TreeView {
	rootDir := "/"
	root := tview.NewTreeNode(rootDir).SetColor(tview.Styles.TitleColor)
	modFolderTree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	// A helper function which adds the files and directories of the given path
	// to the given target node.
	add := func(target *tview.TreeNode, path string) {
		files, err := ioutil.ReadDir(path)
		sort.Slice(files, func(i, j int) bool {
			return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name())
		})

		if err != nil {
			//panic(err)
		}
		for _, file := range files {
			if !file.IsDir() {
				continue
			}
			node := tview.NewTreeNode(file.Name()).
				SetReference(filepath.Join(path, file.Name())).
				SetSelectable(true)
			node.SetColor(tview.Styles.PrimaryTextColor)

			target.AddChild(node)
		}
	}

	// Add the current directory to the root node.
	add(root, rootDir)

	// If a directory was selected, open it.
	modFolderTree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()

		if reference == nil {
			return // Selecting the root node does nothing.
		}
		children := node.GetChildren()
		if len(children) == 0 {
			// Load and show files in this directory.
			path := reference.(string)

			fi, err := os.Stat(path)
			switch {
			case err != nil:
				// handle the error and return
			case fi.IsDir():
				// it's a directory
				add(node, path)
			}
		} else {
			// Collapse if visible, expand if collapsed.
			node.SetExpanded(!node.IsExpanded())
		}
	})

	modFolderTree.SetChangedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return // Selecting the root node does nothing.
		}
		preview.Clear()
		fmt.Fprintf(preview, "mod path: %s", reference.(string))
		config.WadDir = reference.(string)
	})

	modFolderTree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := event.Key()

		switch k {
		case tcell.KeyCtrlO:
			config.Configured = true
			err := cfg.Persist()
			if err != nil {
				// TODO - handle this
			}
			cfg.EnableBasePath()
			appModeNormal()
			return nil
		}

		return event
	})

	return modFolderTree
}
