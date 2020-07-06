package main

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/justinjudd/competition-example/sse"
	"github.com/justinjudd/competition-example/sse/broker"

	"github.com/justinjudd/competition"
	"github.com/justinjudd/competition/models"
)

// Tournament is the main components acting as glue between the web components and the Tournement/Competition components
type Tournament struct {
	Arenas       []models.Arena
	ArenaUsage   map[models.Arena][]models.Game
	ActiveRounds []models.Round
	broker.Broker
	showScores         bool
	CurrentCompetition models.Tournament
	CompetitionCounter int
	models.Competition
	models.StorageEngine
	Leaderboard
}

// SendEvent sends an event notification to listening browser clients
func (t *Tournament) SendEvent(id string, message string) {
	evt := sse.NewEvent(id, []byte(strings.Replace(message, "\n", " ", -1)))
	err := t.Broadcast(evt)
	if err != nil {
		fmt.Println("Unable to broadcast event:", err)
	}
}

func NewTournament(broker broker.Broker, storage models.StorageEngine) *Tournament {
	t := Tournament{}
	t.Broker = broker
	t.StorageEngine = storage

	f, err := os.Open("players.csv")
	if err != nil {
		fmt.Println("Unable to open players csv file:", err)
		return nil
	}
	r := csv.NewReader(f)
	playerRecords, err := r.ReadAll()
	if err != nil {
		fmt.Println("Unable to read players csv file:", err)
		return nil
	}
	players := []models.Player{}

	for _, p := range playerRecords[1:] {
		player := t.CreatePlayer(p[1], []byte(p[2]))
		players = append(players, player)
	}

	t.Leaderboard = NewLeaderboard(players, 200)

	t.Competition = t.CreateCompetition("Robot Rumble", players)

	colors := []string{
		"Blue",
		"Red",
		"Yellow",
		"Green",
	}
	t.Arenas = make([]models.Arena, 4)
	for i := 0; i < len(t.Arenas); i++ {
		a := t.CreateArena(fmt.Sprintf("Arena %s", colors[i]))
		t.Arenas[i] = a
	}

	t.ArenaUsage = map[models.Arena][]models.Game{}
	t.showScores = false

	t.CurrentCompetition = t.RobotTag()

	t.AdvanceRound()

	return &t
}

func (t *Tournament) NextStage() {
	t.GetActiveTournament().SetFinal()
	if t.CurrentCompetition.GetStatus() == models.Status_COMPLETED {
		// Move to next competition - Wait for admin to move it in order to keep displaying results from this portion

	}

}

func (t *Tournament) CompleteCompetition() {
	t.GetActiveTournament().SetFinal()
	t.CurrentCompetition.SetFinal()
	t.NextCompetition()
}

func (t *Tournament) NextCompetition() {
	if t.CurrentCompetition.GetStatus() == models.Status_COMPLETED {
		compName := t.CurrentCompetition.GetName()
		switch {
		case strings.Contains(compName, "Tag"):
			t.CurrentCompetition = t.WorldCup()
			t.AdvanceRound()
		case strings.Contains(compName, "World Cup"):

		}
	}
}

func (t *Tournament) AddMatchPoints(game models.Game, team models.Team, points int) error {

	teams := game.GetTeams()
	teamIndex := -1
	for i, t := range teams {
		if team.Equals(t) {
			teamIndex = i
			break
		}
	}
	if teamIndex < 0 {
		return fmt.Errorf("Unable to find team in provided game.")
	}
	scores := game.GetScores()
	scores[teamIndex] += int64(points)
	game.SetScores(scores)

	t.SignalCurrentGames()
	arena := game.GetArena()
	if arena != nil {
		t.SignalArenaChange(arena)
	}

	return nil
}

func (t *Tournament) SubmitCompletedGame(game models.Game) error {

	if game.GetStatus() == models.Status_COMPLETED { // Game already marked as completed
		return nil
	}

	thisRound := t.CurrentCompetition.GetActiveRound()

	game.SetFinal()

	placesSet := false
	for _, place := range game.GetPlaces() {
		if place != 0 {
			placesSet = true
			break
		}
	}
	if !placesSet { // Set places based on the scores
		setPlaces(game)
	}

	// Remove game from arena schedule

	t.SignalCurrentGames()

	allGamesComplete := true
	for _, g := range thisRound.GetGames() {
		if g.GetStatus() != models.Status_COMPLETED {
			allGamesComplete = false
			break
		}
	}
	if allGamesComplete {
		thisRound.SetFinal()
		t.AdvanceRound()
	}

	return nil
}

func (t *Tournament) AdvanceRound() {
	t.ActiveRounds = []models.Round{}
	defer func() {
		data, err := competition.GenerateTournamentHTML(t.CurrentCompetition)
		if err == nil {
			t.SendEvent("main", string(data))
		} else {
			fmt.Println("Error getting main bracket:", err)
		}
	}()

	r, err := t.CurrentCompetition.NextRound()
	if err != nil {
		_, over := err.(competition.CompetitionOverError)
		if !over {
			fmt.Println(err)
			return
		} else {
			t.NextStage()
			return
		}
	}
	t.ActiveRounds = append(t.ActiveRounds, r)

	t.AssignArenas(false)

	out, err := json.Marshal(t.GetRankings())
	if err == nil {
		t.SendEvent("leaderboard", string(out))
	}

	for a := range t.ArenaUsage {
		t.SignalArenaChange(a)
	}

	t.SignalCurrentGames()

}

func (t *Tournament) AssignArenas(assignByes bool) {
	t.ArenaUsage = map[models.Arena][]models.Game{}
	counter := 0
	for _, round := range t.ActiveRounds {
		for _, game := range round.GetGames() {

			if models.IsByeGame(game, int(t.CurrentCompetition.GetAdvancing())) && !assignByes {
				game.SetFinal()
				continue
			}

			arena := t.Arenas[counter%len(t.Arenas)]
			game.SetArena(arena)
			t.ArenaUsage[arena] = append(t.ArenaUsage[arena], game)
			counter++
		}
	}

}

func (t *Tournament) CurrentGames() []models.Game {
	current := []models.Game{}
	for _, round := range t.ActiveRounds {
		current = append(current, round.GetGames()...)
	}
	return current
}

func (t *Tournament) SignalCurrentGames() {
	var gameData []string
	current := t.CurrentGames()
	for _, game := range current {
		if game.GetArena() == nil { // Only send scheduled events
			continue
		}

		data := map[string]interface{}{"game": prepGameForJson(game, t.CurrentCompetition), "scored": t.showScores}
		out, err := json.Marshal(data)
		if err == nil && game.GetArena().GetName() != "" {
			gameData = append(gameData, string(out))
		}

	}

	t.SendEvent("current", "["+strings.Join(gameData, ", ")+"]")
}

type jsonArena struct {
	Id   string `json:"id"`
	Name string `json:"Name"`
}
type jsonTeam struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}
type jsonGame struct {
	Scores   []int64       `json:"scores"`
	Places   []int64       `json:"places"`
	Status   models.Status `json:"status"`
	Teams    []jsonTeam    `json:"teams"`
	Arena    models.Arena  `json:"arena"`
	MaxTeams uint32        `json:"maxTeams"`
}

func prepGameForJson(game models.Game, tournament models.Tournament) jsonGame {

	g := jsonGame{}
	g.Scores = game.GetScores()
	g.Places = game.GetPlaces()
	g.Status = game.GetStatus()
	g.MaxTeams = tournament.GetGameSize()
	for _, team := range game.GetTeams() {
		g.Teams = append(g.Teams, jsonTeam{Id: team.GetName(), Name: team.GetName(), Image: string(team.GetMetadata())})
	}
	g.Arena = game.GetArena()

	return g
}

func (t *Tournament) SignalArenaChange(arena models.Arena) {
	arenaName := strings.ToLower(strings.Replace(arena.GetName(), " ", "", -1))
	queuedGames := arena.GetGames()
	if len(queuedGames) == 0 {
		return
	}

	gameData := []string{}
	for _, game := range queuedGames[1:] {

		data := map[string]interface{}{"game": prepGameForJson(game, t.CurrentCompetition), "scored": t.showScores}
		out, err := json.Marshal(data)
		if err == nil && game.GetArena().GetName() != "" {
			gameData = append(gameData, string(out))
		}

	}
	t.SendEvent(fmt.Sprintf("%s-next", arenaName), "["+strings.Join(gameData, ", ")+"]")

	currentGame := queuedGames[0]
	data := map[string]interface{}{"game": prepGameForJson(currentGame, t.CurrentCompetition), "scored": t.showScores}
	out, err := json.Marshal(data)

	if err == nil {
		t.SendEvent(fmt.Sprintf("%s-current", arenaName), string(out))
	} else {
		fmt.Println(err)
	}

}

func (t *Tournament) SignalGameStart(arena models.Arena) {
	arenaName := strings.ToLower(strings.Replace(arena.GetName(), " ", "", -1))
	queuedGames := arena.GetGames()
	if len(queuedGames) == 0 {
		return
	}
	md := t.CurrentCompetition.GetMetadata()
	buf := bytes.NewBuffer(md)
	gd := gob.NewDecoder(buf)
	var matchLength time.Duration
	err := gd.Decode(&matchLength)
	if err == nil {
		endTime := time.Now().Add(matchLength)

		t.SendEvent(fmt.Sprintf("%s-start", arenaName), strconv.FormatInt(endTime.UnixNano()/1000000, 10))
	}

	t.SignalArenaChange(arena)
	t.SignalCurrentGames()
}

func (t *Tournament) SignalGameEnd(arena models.Arena) {
	arenaName := strings.ToLower(strings.Replace(arena.GetName(), " ", "", -1))
	queuedGames := arena.GetGames()
	if len(queuedGames) == 0 {
		return
	}
	endTime := time.Now()
	t.SendEvent(fmt.Sprintf("%s-start", arenaName), strconv.FormatInt(endTime.UnixNano()/1000000, 10))

}

func countryImage(country string) string {
	return fmt.Sprintf("/static/images/flags/4x3/%s.svg", countryCodes[country])
}

var countryCodes = map[string]string{
	"United States":       "us",
	"Brazil":              "br",
	"United Kingdom":      "gb",
	"Sweden":              "se",
	"Switzerland":         "ch",
	"Japan":               "jp",
	"Italy":               "it",
	"Greece":              "gr",
	"France":              "fr",
	"Finland":             "fi",
	"Germany":             "de",
	"El Salvador":         "sv",
	"Chile":               "cl",
	"Australia":           "au",
	"Argentina":           "ar",
	"Canada":              "ca",
	"Costa Rica":          "cr",
	"Iceland":             "is",
	"China":               "cn",
	"South Korea":         "kr",
	"Trinidad and Tobago": "tt",
	"Mexico":              "mx",
	"Ireland":             "ie",
	"Puerto Rico":         "pr",
	"New Zealand":         "nz",
	"Uganda":              "ug",
	"Wales":               "gb-wls",
	"England":             "gb-eng",
	"Russia":              "ru",
	"Christmas Island":    "cx",
	"Philippines":         "ph",
	"Denmark":             "dk",
	"Fiji":                "fj",
	"India":               "in",
	"Belarus":             "by",
	"Chad":                "td",
	"Norway":              "no",
}

func setPlaces(g models.Game) {
	if g.GetStatus() != models.Status_COMPLETED {
		return
	}

	byeTeams := map[int]bool{}
	for i, team := range g.GetTeams() {
		if models.IsByeTeam(team) {
			byeTeams[i] = true
			g.GetPlaces()[i] = int64(len(g.GetTeams()))
			g.GetScores()[i] = 0
		}
	}

	scores := map[models.Team]int{}
	rankedTeams := make([]models.Team, len(g.GetTeams())-len(byeTeams))
	placeCounter := 0
	for i, team := range g.GetTeams() {
		if byeTeams[i] {
			continue
		}
		rankedTeams[placeCounter] = team
		placeCounter++
	}
	//copy(rankedTeams, g.Teams)
	places := map[models.Team]int{}
	for i, team := range g.GetTeams() {
		if byeTeams[i] {
			continue
		}
		scores[team] = int(g.GetScores()[i])
		places[team] = i
	}

	sort.Slice(rankedTeams, func(i, j int) bool {
		t1, t2 := rankedTeams[i], rankedTeams[j]

		return scores[t2] < scores[t1] // flip so ranked highest first
	})

	nextPlace := 0
	for i, team := range rankedTeams {
		score := scores[team]
		var tie bool
		thisPlace := nextPlace
		if i+1 < len(rankedTeams) {
			t2 := rankedTeams[i+1]
			peek := scores[t2]
			if peek == score { // there has been a tie
				tie = true

			}
		}
		if i != 0 {
			t2 := rankedTeams[i-1]
			prev := scores[t2]
			if prev == score { // there has been a tie
				tie = true
				thisPlace = nextPlace - 1
				nextPlace++
			}
		}
		place := places[team]

		if !tie {
			g.GetPlaces()[place] = int64(thisPlace)

		} else {
			g.GetPlaces()[place] = int64(thisPlace)*-1 - 1
		}
		nextPlace++
	}

}
