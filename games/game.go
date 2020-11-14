package games

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zmnpl/twad/cfg"
	st "github.com/zmnpl/twad/games/savesStats"
	"github.com/zmnpl/twad/helper"
)

// Game represents one game configuration
type Game struct {
	Name             string         `json:"name"`
	SourcePort       string         `json:"source_port"`
	Iwad             string         `json:"iwad"`
	Config           string         `json:"config"`
	Environment      []string       `json:"environment"`
	Mods             []string       `json:"mods"`
	CustomParameters []string       `json:"custom_parameters"`
	ConsoleStats     map[string]int `json:"stats"`
	Playtime         int64          `json:"playtime"`
	LastPlayed       string         `json:"last_played"`
	SaveGameCount    int            `json:"save_game_count"`
	Rating           int            `json:"rating"`
	Stats            []st.MapStats
	StatsTotal       st.MapStats
	Savegames        []*st.Savegame
}

// NewGame creates new instance of a game
func NewGame(name, sourceport, iwad, cnfg string) Game {
	config := cfg.Instance()

	game := Game{
		Name:             name,
		SourcePort:       "gzdoom",
		Iwad:             "doom2.wad",
		Config:           cfg.Instance().DefaultConfigLabel,
		Environment:      make([]string, 0),
		CustomParameters: make([]string, 0),
		Mods:             make([]string, 0),
		ConsoleStats:     make(map[string]int),
	}

	// replace with given or first list entry
	if sourceport != "" {
		game.SourcePort = sourceport
	} else {
		if len(config.SourcePorts) > 0 {
			game.SourcePort = config.SourcePorts[0]
		}
	}

	// replace with given or first list entry
	if iwad != "" {
		game.Iwad = iwad
	} else {
		if len(config.IWADs) > 0 {
			game.Iwad = config.IWADs[0]
		}
	}
	
	// replace with given or source port default
	if cnfg != "" {
		game.Config = cnfg
    } else {
		game.Config = cfg.Instance().DefaultConfigLabel
    }

	return game
}

// Run executes given configuration and launches the mod
// Just a wrapper for game.run
func (g *Game) Run() (err error) {
	g.run(*newRunConfig())
	return
}

// Quickload starts the game from it's last savegame
// Just a wrapper for game.run
func (g *Game) Quickload() (err error) {
	g.run(*newRunConfig().quickload())
	return
}

// Warp lets you select episode and level to start in
// Just a wrapper for game.run
func (g *Game) Warp(episode, level, skill int) (err error) {
	g.run(*newRunConfig().warp(episode, level).setSkill(g.spAdjustedSkill(skill)))
	return
}

// WarpRecord lets you select episode and level to start in
// Just a wrapper for game.run
func (g *Game) WarpRecord(episode, level, skill int, demoName string) (err error) {
	g.run(*newRunConfig().warp(episode, level).setSkill(g.spAdjustedSkill(skill)).recordDemo(demoName))
	return
}

// PlayDemo replays the given demo file
// Wrapper for game.run
func (g *Game) PlayDemo(name string) {
	g.run(*newRunConfig().playDemo(name))
}

// AddMod adds mod
func (g *Game) AddMod(modFile string) {
	g.Mods = append(g.Mods, modFile)
	InformChangeListeners()
	Persist()
}

// RemoveMod removes mod at the given index
func (g *Game) RemoveMod(i int) {
	g.Mods = append(g.Mods[0:i], g.Mods[i+1:]...)
}

func (g *Game) run(rcfg runconfig) (err error) {
	start := time.Now()

	// change working directory to redirect stat file output for boom
	wd, wdChangeError := os.Getwd()
	if sourcePortFamily(g.SourcePort) == boom {
		os.Chdir(g.getSaveDir())
	}

	// rip and tear!
	doom := g.composeProcess(g.getLaunchParams(rcfg))
	output, err := doom.CombinedOutput()
	if err != nil {
		ioutil.WriteFile("twad.log", []byte(fmt.Sprintf("%v\n\n%v\n\n%v\n\n%v", string(output), err.Error(), g.getLaunchParams(rcfg), doom)), 0755)
		return err
	}

	// change back working directory to where it was
	wdNow, _ := os.Getwd()
	if wd != wdNow && wdChangeError == nil {
		os.Chdir(wd)
	}

	playtime := time.Since(start).Milliseconds()
	g.Playtime = g.Playtime + playtime
	g.LastPlayed = time.Now().Format("2006-01-02 15:04:05MST")

	// could take a while ...
	go g.processOutput(string(output))
	go g.ReadLatestStats()

	return
}

func (g *Game) composeProcess(params []string) (cmd *exec.Cmd) {
	// create process object
	cmd = exec.Command(g.SourcePort, params...)
	// add environment variables; use os environment as basis
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, g.Environment...)
	return
}

func (g *Game) getLaunchParams(rcfg runconfig) []string {
	params := make([]string, 0, 10)

	// IWAD
	if g.Iwad != "" {
		params = append(params, "-iwad", g.Iwad) // -iwad seems to be universal across zdoom, boom and chocolate doom
	}

	// config
	if g.Config != cfg.Instance().DefaultConfigLabel && g.Config != "" {
		// Gets full path, as shell working directory seems to be where gzdoom searches, but not for mods?
		// I tried to dig into the source for gzdoom, but had no luck finding a difference.
		
		// Nevermind, figured it out. Source ports are searching all of the Paths for the relative folder/wad files.
		// But -config only searches in the current working dir. I would change the dir of the program, but run() just above does
		// that for boom stuff. So, this is my solution.
		// configPath := "$DOOMWADDIR/config/"+g.Config; did not work
		configPath := cfg.Instance().WadDir+"/config/"+g.Config
		params = append(params, "-config", configPath) // -config seems to be universal across zdoom, boom and chocolate doom
	}
	
	
	// mods
	if len(g.Mods) > 0 {
		params = append(params, "-file") // -file seems to be universal across zdoom, boom and chocolate doom
		params = append(params, g.Mods...)
	}

	// custom game save directory
	// making dir seems to be redundant, since engines do that already
	// still keeping it to possibly keep track of it / handle errors
	// only use separate save dir if directory has been craeted or path exists already
	if err := os.MkdirAll(g.getSaveDir(), 0755); err == nil {
		params = append(params, g.spSaveDirParam())
		params = append(params, g.getSaveDir())
	}

	// stats for chocolate doom and ports
	if sourcePortFamily(g.SourcePort) == chocolate {
		params = append(params, "-statdump")
		params = append(params, path.Join(g.getSaveDir(), "statdump.txt"))
	}

	// stats for chocolate doom and ports
	if sourcePortFamily(g.SourcePort) == boom {
		params = append(params, "-levelstat")
	}

	// quickload
	if rcfg.loadLastSave {
		params = append(params, g.getLastSaveLaunchParams()...)
	}

	// warp
	if rcfg.beam && (rcfg.warpEpisode > 0 || rcfg.warpLevel > 0) {
		params = append(params, "-warp")
		if rcfg.warpEpisode > 0 {
			params = append(params, strconv.Itoa(rcfg.warpEpisode))
		}
		if rcfg.warpLevel > 0 {
			params = append(params, strconv.Itoa(rcfg.warpLevel))
		}

		// add skill
		params = append(params, "-skill")
		params = append(params, strconv.Itoa(rcfg.skill))
	}

	// demo recording
	if rcfg.recDemo {
		if err := os.MkdirAll(g.getDemoDir(), 0755); err == nil {
			params = append(params, "-record") // TODO: Does -record behave equally across ports?
			params = append(params, g.getDemoDir()+"/"+rcfg.demoName)
		}
	}

	// play demo
	if rcfg.plyDemo {
		params = append(params, "-playdemo")
		params = append(params, g.getDemoDir()+"/"+rcfg.demoName)
	}

	return append(params, g.CustomParameters...)
}

func (g *Game) getLastSaveLaunchParams() (params []string) {
	params = []string{}

	if lastSave, err := g.lastSave(); err == nil {
		params = append(params, []string{"-loadgame", lastSave}...) // -loadgame seems to be universal across zdoom, boom and chocolate doom
	}
	return
}

// CommandList returns the full slice of strings in order to launch the game
func (g *Game) CommandList() (command []string) {
	command = g.Environment
	command = append(command, g.SourcePort)
	command = append(command, g.getLaunchParams(*newRunConfig().quickload())...)
	return
}

// SaveCount returns the number of savegames existing for this game
func (g *Game) SaveCount() int {
	if saves, err := g.savegameFiles(); err == nil {
		return len(saves)
	}
	return 0
}

// savegameFiles returns a slice of os.FileInfo with all savegmes for this game
func (g *Game) savegameFiles() ([]os.FileInfo, error) {
	saves, err := ioutil.ReadDir(g.getSaveDir())
	if err != nil {
		return nil, err
	}
	saves = helper.FilterExtensions(saves, g.spSaveFileExtension(), false)

	sort.Slice(saves, func(i, j int) bool {
		return saves[i].ModTime().After(saves[j].ModTime())
	})

	return saves, nil
}

// LoadSavegames returns a slice of Savegames for the game
func (g *Game) LoadSavegames() []*st.Savegame {
	saveDir := g.getSaveDir()
	savegames := make([]*st.Savegame, 0)
	savegameFiles, _ := g.savegameFiles()

	for i, s := range savegameFiles {
		savegame := st.NewSavegame(s, saveDir)
		g.loadSaveMeta(&savegame)

		// load stats
		// after 3 load parallel
		if i <= 2 {
			g.loadSaveStats(&savegame)
		} else {
			go g.loadSaveStats(&savegame)
		}
		savegames = append(savegames, &savegame)
	}
	g.Savegames = savegames
	return savegames
}

func (g *Game) loadSaveMeta(s *st.Savegame) {
	s.Meta = g.GetSaveMeta(path.Join(s.Directory, s.FI.Name()))
}

func (g *Game) loadSaveStats(s *st.Savegame) {
	s.Levels = g.GetStats(path.Join(s.Directory, s.FI.Name()))
}

// GetSaveMeta reads meta information for the given savegame
func (g *Game) GetSaveMeta(savePath string) st.SaveMeta {
	if sourcePortFamily(g.SourcePort) == chocolate {
		meta, _ := st.ChocolateMetaFromBinary(savePath)
		return meta
	} else if sourcePortFamily(g.SourcePort) == boom {
		meta, _ := st.ChocolateMetaFromBinary(savePath)
		return meta
	}

	return st.GetZDoomSaveMeta(savePath)
}

// GetStats reads stats from the given savegame path for zdoom ports
// If the port is boom or chocolate, their respective dump-files are used
func (g *Game) GetStats(savePath string) []st.MapStats {
	var stats []st.MapStats
	if sourcePortFamily(g.SourcePort) == chocolate {
		stats, _ = st.GetChocolateStats(path.Join(g.getSaveDir(), "statdump.txt"))
	} else if sourcePortFamily(g.SourcePort) == boom {
		stats, _ = st.GetBoomStats(path.Join(g.getSaveDir(), "levelstat.txt"))
	} else {
		stats = st.GetZDoomStats(savePath)
	}

	return stats
}

// ReadLatestStats tries to read stats from the newest existing savegame
func (g *Game) ReadLatestStats() {
	lastSavePath, _ := g.lastSave()
	g.Stats = g.GetStats(lastSavePath)
	g.StatsTotal = st.SummarizeStats(g.Stats)
}

// DemoCount returns the number of demos existing for this game
func (g *Game) DemoCount() int {
	if demos, err := ioutil.ReadDir(g.getDemoDir()); err == nil {
		return len(demos)
	}
	return 0
}

// Rate increases or decreases the games rating
func (g *Game) Rate(increment int) {
	g.Rating += increment
	switch {
	case g.Rating > 5:
		g.Rating = 5
	case g.Rating < 0:
		g.Rating = 0
	}

}

// SwitchMods switches both entries within the mod slice
func (g *Game) SwitchMods(a, b int) {
	if a < len(g.Mods) && b < len(g.Mods) {
		modA := g.Mods[a]
		modB := g.Mods[b]
		g.Mods[a] = modB
		g.Mods[b] = modA
	}
}

func (g *Game) getSaveDir() string {
	return filepath.Join(cfg.GetSavegameFolder(), g.cleansedName())
}

// lastSave returns the the file name or slotnumber (depending on source port) for the game
func (g *Game) lastSave() (save string, err error) {
	saveDir := g.getSaveDir()
	saves, err := ioutil.ReadDir(saveDir)
	if err != nil {
		return
	}

	// assume zdoom
	portSaveFileExtension := g.spSaveFileExtension()

	// find the newest file
	newestTime, _ := time.Parse(time.RFC3339, "1900-01-01T00:00:00+00:00")
	for _, file := range saves {
		extension := strings.ToLower(filepath.Ext(file.Name()))
		if file.Mode().IsRegular() && file.ModTime().After(newestTime) && extension == portSaveFileExtension {
			save = filepath.Join(saveDir, file.Name())
			newestTime = file.ModTime()
		}
	}

	// adjust for different souce ports
	save = g.spSaveGameName(save)

	if save == "" {
		err = os.ErrNotExist
	}

	return
}

// Demos returns the demo files existing for the game
func (g *Game) Demos() ([]os.FileInfo, error) {
	demos, err := ioutil.ReadDir(g.getDemoDir())
	if err != nil {
		return nil, err
	}
	sort.Slice(demos, func(i, j int) bool {
		return demos[i].ModTime().After(demos[j].ModTime())
	})
	return demos, err
}

// DemoExists checks if a file with the same name already exists in the default demo dir
// Doesn't use standard library to ignore file ending; design decision
func (g *Game) DemoExists(name string) bool {
	if files, err := ioutil.ReadDir(g.getDemoDir()); err == nil {
		for _, f := range files {
			nameWithouthExt := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			if nameWithouthExt == name {
				return true
			}
		}
	}
	return false
}

// RemoveDemo removes the demo file with the given name
// and returns the new set of demos
func (g *Game) RemoveDemo(name string) ([]os.FileInfo, error) {
	err := os.Remove(filepath.Join(g.getDemoDir(), name))
	if err != nil {
		return nil, err
	}
	return g.Demos()
}

func (g *Game) getDemoDir() string {
	return filepath.Join(cfg.GetDemoFolder(), g.cleansedName())
}

// cleansedName removes all but alphanumeric characters from name
// used for directory names
func (g *Game) cleansedName() string {
	cleanser, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return cleanser.ReplaceAllString(g.Name, "")
}

// processOutput processes the terminal output of the zdoom port
func (g *Game) processOutput(output string) {
	if g.ConsoleStats == nil {
		g.ConsoleStats = make(map[string]int)
	}
	for _, v := range strings.Split(output, "\n") {
		if stat, increment := parseStatline(v, g); stat != "" {
			g.ConsoleStats[stat] = g.ConsoleStats[stat] + increment
		}
	}

	Persist()
}

// parseStatLine receives each line from processOutput()
// if the line matches a known pattern it will be added to the games stats
func parseStatline(line string, g *Game) (string, int) {
	line = strings.TrimSpace(line)
	switch {

	case strings.HasPrefix(line, "Picked up a "):
		return strings.TrimSuffix(strings.TrimPrefix(line, "Picked up a "), "."), 1

	case strings.HasPrefix(line, "You got the "):
		return strings.TrimSuffix(strings.TrimPrefix(line, "You got the "), "!"), 1

	case strings.HasPrefix(line, "Level map01 - Kills: 10/19 - Items: 8/9 - Secrets: 0/5 - Time: 0:35"):
		return "", 1

	default:
		return "", 0
	}
}

// Printing Methods
// String returns the string which is run when running

// RatingString returns the string resulting from the games rating
func (g *Game) RatingString() string {
	return strings.Repeat("*", g.Rating) + strings.Repeat("-", 5-g.Rating)
}

// EnvironmentString returns a join of all prefix parameters
func (g *Game) EnvironmentString() string {
	return strings.TrimSpace(strings.Join(g.Environment, " "))
}

// ParamsString returns a join of all prefix parameters
func (g *Game) ParamsString() string {
	return strings.TrimSpace(strings.Join(g.CustomParameters, " "))
}
