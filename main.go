package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/justinjudd/competition-example/sse"
	"github.com/justinjudd/competition-example/sse/broker"

	"github.com/justinjudd/competition"
	"github.com/justinjudd/competition/models"
	"github.com/justinjudd/competition/models/storm"

	"github.com/gorilla/mux"
)

// Server manages the whole tournament and web server
type Server struct {
	broker.Broker
	*Tournament
	models.StorageEngine
	Router *mux.Router
	*template.Template
}

// NewServer creates and returns a new Server, containing the underlying StorageEngine and HTTP Server components
func NewServer() *Server {
	s := Server{}

	s.Broker = NewBroker()

	storage, err := storm.NewStorageEngine("record.db")
	if err != nil {
		panic("Unable to open storage engine:" + err.Error())
	}
	s.StorageEngine = storage
	s.Tournament = NewTournament(s.Broker, s.StorageEngine)

	funcMap := template.FuncMap{

		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"last": func(x int, a interface{}) bool {
			return x == reflect.ValueOf(a).Len()-1
		},
		"winner": func(game models.Game, team models.Team) bool {
			if team == nil { // team was a bye team
				return false
			}
			if game.GetStatus() != models.Status_COMPLETED {
				return false
			}
			var located int
			for i, t := range game.GetTeams() {
				if t == nil { // bye team
					continue
				}
				if t.Equals(team) {
					located = i
					break
				}
			}
			place := competition.FlipTies(int(game.GetPlaces()[located]))
			return place < len(game.GetTeams())/2
		},
		"score": func(game models.Game, team models.Team) int {
			var located int
			for i, t := range game.GetTeams() {
				if t.Equals(team) {
					located = i
					break
				}
			}
			return int(game.GetScores()[located])

		},
		"complete": func(game models.Game) bool {
			return game.GetStatus() == models.Status_COMPLETED
		},
		"width": func(game models.Game) int {
			return 12 / len(game.GetTeams())
		},
		"tabletWidth": func(game models.Game) int {
			return 24 / len(game.GetTeams())
		},
		"mobileWidth": func(game models.Game) int {
			return 36 / len(game.GetTeams())
		},
		"backgroundColor": func(game models.Game, n int) string {

			teams := game.GetTeams()
			if n > len(teams) {
				return ""
			}

			team := teams[n]
			if len(team.GetMetadata()) > 0 { // Don't show background if there is a team image
				return ""
			}

			switch n {
			case 0:
				return "blue"
			case 1:
				return "red"
			case 2:
				return "yellow"
			case 3:
				return "green"
			}
			return ""
		},
		"slice": func(n int) []int {
			return make([]int, n)
		},
		"add": func(a int, b int) int {
			return a + b
		},
		"addUint": func(a uint32, b int) int {
			return int(a) + b
		},
		"urlprep": func(s string) string {
			s = strings.Replace(s, " ", "", -1)
			return strings.ToLower(s)
		},
		"place": func(t models.Team, g models.Game) int {
			return competition.FlipTies(int(g.GetTeamPlace(t)))

		},
		"containingTeam": func(p models.Player, g models.Game) models.Team {
			for _, t := range g.GetTeams() {
				for _, player := range t.GetPlayers() {
					if p.GetName() == player.GetName() {
						return t
					}
				}
			}
			return nil
		},
	}
	t, err := template.New("templates").Funcs(funcMap).ParseGlob("./templates/*.tmpl")
	if err != nil {
		fmt.Println(err)
	}
	s.Template = t
	return &s
}

func (s *Server) ArenaHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	arenaName, ok := vars["arena"]
	if !ok {
		return
	}
	var arena models.Arena
	for _, a := range s.Arenas {
		if arenaName == strings.ToLower(strings.Replace(a.GetName(), " ", "", -1)) {
			arena = a
		}
	}
	if arena == nil {
		return
	}

	colors := map[int]string{1: "blue", 2: "red", 3: "yellow", 4: "green"}
	err := s.ExecuteTemplate(w, "arena.tmpl", map[string]interface{}{"arenaName": arenaName, "arena": arena, "colors": colors})
	if err != nil {
		fmt.Println(err)
	}

}

func (s *Server) TeamHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	teamNameString, ok := vars["teamId"]
	if !ok {
		fmt.Println("Unable to get team name")
		return
	}

	player := s.GetPlayer(teamNameString)
	if player == nil {
		fmt.Println("Unable to find matching team.")
		return
	}

	games := player.GetRecords()

	colors := map[int]string{1: "blue", 2: "red", 3: "yellow", 4: "green"}
	err := s.ExecuteTemplate(w, "team.tmpl", map[string]interface{}{
		"player":        player,
		"games":         games[:len(games)-1],
		"currentGame":   games[len(games)-1],
		"colors":        colors,
		"currentScored": s.showScores,
	})
	if err != nil {
		fmt.Println(err)
	}

}

func (s *Server) BracketHandler(w http.ResponseWriter, r *http.Request) {

	data := map[string]string{}
	err := s.ExecuteTemplate(w, "bracket.tmpl", data)
	if err != nil {
		fmt.Println(err)
	}

}

func (s *Server) AdminArenaHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	arenaName, ok := vars["arena"]
	if !ok {
		return
	}
	arenas := s.GetArenas()
	var arena models.Arena
	for _, a := range arenas {
		if arenaName == strings.ToLower(strings.Replace(a.GetName(), " ", "", -1)) {
			arena = a
			break
		}
	}
	if arena == nil {
		fmt.Println("Unable to find arena:", arenaName)
		return
	}
	queuedGames := arena.GetGames()
	if len(queuedGames) == 0 {

		err := s.ExecuteTemplate(w, "blank.tmpl", map[string]interface{}{"content": "No current game at this arena. Please refresh the page."})
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	game := queuedGames[0]

	ctx := r.Context()
	admin := ctx.Value(adminKey)

	var queued []models.Game
	if len(queuedGames) > 1 {
		queued = queuedGames[1:]
	}

	colors := map[int]string{1: "blue", 2: "red", 3: "yellow", 4: "green"}
	err := s.ExecuteTemplate(w, "admin-arena.tmpl", map[string]interface{}{"arenaName": arenaName, "game": game, "arena": arena, "scored": s.showScores, "admin": admin, "colors": colors, "queuedGames": queued})
	if err != nil {
		fmt.Println(err)
	}

}

func (s *Server) AdminUpdateScore(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		arenaName := r.PostFormValue("arena")
		teamName := r.PostFormValue("team")
		score := r.PostFormValue("score")
		points, err := strconv.ParseInt(score, 10, 64)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, a := range s.GetArenas() {
			if strings.ToLower(strings.Replace(a.GetName(), " ", "", -1)) != arenaName {
				continue
			}
			games := a.GetGames()
			if len(games) == 0 {
				return
			}
			game := games[0]
			for _, team := range game.GetTeams() {
				if team.GetName() == teamName {
					s.AddMatchPoints(game, team, int(points))

					http.Redirect(w, r, fmt.Sprintf("/admin/arena/%s", arenaName), 302)
					return
				}
			}
			break
		}

	} else {
		fmt.Println("Who is trying to make a get request to the score updating page?")
	}
}

func fixPlaces(places []int64) (fixed []int64) {

	fixed = make([]int64, len(places))

	placeMap := map[int64]int{}
	for _, place := range places {
		placeMap[place]++
	}
	if placeMap[0] == len(places) { // All tied at first place
		for i := range places {
			fixed[i] = -1
		}

	} else {
		placesPerPlace := map[int64]int64{}
		currentPlace := int64(0)
		total := int64(len(places))
		for i := int64(0); i <= total; i++ {
			switch placeMap[i] {
			case 0: //None match
			case 1: // Just one team had this place (No tie for this spot)
				placesPerPlace[i] = currentPlace
				currentPlace++
			default: // There were ties
				placesPerPlace[i] = currentPlace * -1
				currentPlace++
			}
		}
		for i, place := range places {
			fixed[i] = placesPerPlace[place]
		}
	}

	return fixed
}

func (s *Server) AdminUpdatePlaces(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		arenaName := r.PostFormValue("arena")

		for _, a := range s.GetArenas() {
			if strings.ToLower(strings.Replace(a.GetName(), " ", "", -1)) != arenaName {
				continue
			}
			games := a.GetGames()
			if len(games) == 0 {
				return
			}
			game := games[0]
			places := game.GetPlaces()
			for i := 0; i < 4; i++ {
				p := r.PostFormValue(fmt.Sprintf("place-%d", i))
				if len(p) == 0 {
					continue
				}
				place, err := strconv.Atoi(p)
				if err != nil {
					fmt.Println("Received invalid place when trying to set place")
				}
				places[i] = int64(place)

			}

			game.SetPlaces(places)

			s.SignalCurrentGames()
			if a != nil {
				s.SignalArenaChange(a)
			}
			break
		}

		http.Redirect(w, r, fmt.Sprintf("/admin/arena/%s", arenaName), 302)

	} else {
		fmt.Println("Who is trying to make a get request to the place updating page?")
	}
}

func (s *Server) AdminGameStart(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		arenaName := r.PostFormValue("arena")

		for _, a := range s.GetArenas() {
			if strings.ToLower(strings.Replace(a.GetName(), " ", "", -1)) != arenaName {
				continue
			}
			games := a.GetGames()
			if len(games) == 0 {
				return
			}
			game := games[0]
			game.Start()
			s.SignalGameStart(a)

			http.Redirect(w, r, fmt.Sprintf("/admin/arena/%s", arenaName), 302)
			break
		}

	} else {
		fmt.Println("Who is trying to make a get request to the game start page?")
	}
}

func (s *Server) AdminGameComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		arenaName := r.PostFormValue("arena")
		placeStringsArray := r.PostForm["place"]
		placeStrings := strings.Split(placeStringsArray[0], ",")
		places := make([]int64, len(placeStrings))
		placeMap := map[int64][]int{}
		for i, place := range placeStrings {
			p, err := strconv.Atoi(place)
			if err != nil {
			}

			places[i] = int64(p)
			placeMap[int64(p)] = append(placeMap[int64(p)], i)
		}
		for place, indexes := range placeMap {
			if len(indexes) > 1 {
				for _, index := range indexes {
					places[index] = -1 * place
				}
			} else {
				places[indexes[0]] = place - 1
			}
		}
		for _, a := range s.GetArenas() {
			if strings.ToLower(strings.Replace(a.GetName(), " ", "", -1)) != arenaName {
				continue
			}
			games := a.GetGames()
			if len(games) == 0 {
				return
			}
			game := games[0]
			game.SetPlaces(places)
			s.SubmitCompletedGame(game)
			s.SignalGameEnd(a)

			http.Redirect(w, r, fmt.Sprintf("/admin/arena/%s", arenaName), 302)
			break
		}

	} else {
		fmt.Println("Who is trying to make a get request to the game completion page?")
	}
}

func (s *Server) AdminCompetitionComplete(w http.ResponseWriter, r *http.Request) {
	s.CompleteCompetition()

	http.Redirect(w, r, "/admin/", 302)
}

func NewBroker() (broker broker.Broker) {
	config := sse.Config{
		Timeout:      time.Second * 3,
		Tolerance:    3,
		ErrorHandler: nil,
	}
	return sse.NewBroker(config)
}

func main() {

	rand.Seed(20190316)

	s := NewServer()

	r := mux.NewRouter()

	admin := r.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			username := r.PostFormValue("username")
			password := r.PostFormValue("password")

			fmt.Println(username, password)
		} else {
			tmpl, err := template.New("login").ParseFiles("templates/login.tmpl")
			if err != nil {
				fmt.Println(err)
			}
			err = tmpl.ExecuteTemplate(w, "login.tmpl", nil)
			if err != nil {
				fmt.Println(err)
			}
		}

	})

	admin.HandleFunc("/arena/{arena}", basicAuthWrapper(s.AdminArenaHandler)).Name("arena")
	admin.HandleFunc("/updateScore", basicAuthWrapper(s.AdminUpdateScore)).Name("updateScore")
	admin.HandleFunc("/updatePlaces", basicAuthWrapper(s.AdminUpdatePlaces)).Name("updatePlaces")
	admin.HandleFunc("/gameStart", basicAuthWrapper(s.AdminGameStart)).Name("gameStart")
	admin.HandleFunc("/gameComplete", basicAuthWrapper(s.AdminGameComplete)).Name("gameConplete")
	admin.HandleFunc("/competitionComplete", basicAuthWrapper(s.AdminCompetitionComplete)).Name("competitionComplete")

	admin.HandleFunc("/", basicAuthWrapper(func(w http.ResponseWriter, r *http.Request) {

		colors := map[int]string{0: "blue", 1: "red", 2: "yellow", 3: "green"}
		err := s.ExecuteTemplate(w, "admin.tmpl", map[string]interface{}{"arenas": s.Arenas, "colors": colors})
		if err != nil {
			fmt.Println(err)
		}
	})).Name("admin")

	r.HandleFunc("/events", s.Broker.ClientHandler)
	r.HandleFunc("/arena/{arena}", s.ArenaHandler)
	r.HandleFunc("/team/{teamId}", s.TeamHandler)
	r.HandleFunc("/bracket", s.BracketHandler)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		err := s.ExecuteTemplate(w, "main.tmpl", nil)
		if err != nil {
			fmt.Println(err)
		}
	})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/", r)
	s.Router = r

	err := http.ListenAndServe(":10000", nil)
	if err != nil {
		panic(err)
	}

}

var credentials = map[string]string{
	"admin":  "exampleadmin",
	"admin1": "arena1",
	"admin2": "arena2",
	"admin3": "arena3",
	"admin4": "arena4",
}

var adminPaths = map[string][]string{
	"admin": []string{
		"/admin", // allowed to access all admin pages
	},
	"admin1": []string{
		"/admin/arena/arenablue",
		"/admin/updateScore",
		"/admin/gameStart",
		"/admin/updatePlaces",
		"/admin/gameComplete",
	},
	"admin2": []string{
		"/admin/arena/arenared",
		"/admin/updateScore",
		"/admin/gameStart",
		"/admin/updatePlaces",
		"/admin/gameComplete",
	},
	"admin3": []string{
		"/admin/arena/arenayellow",
		"/admin/updateScore",
		"/admin/gameStart",
		"/admin/updatePlaces",
		"/admin/gameComplete",
	},
	"admin4": []string{
		"/admin/arena/arenagreen",
		"/admin/updateScore",
		"/admin/gameStart",
		"/admin/updatePlaces",
		"/admin/gameComplete",
	},
}

type key int

const adminKey key = 0

func checkBasicAuthCredentials(r *http.Request) bool {
	username, pass, ok := r.BasicAuth()
	if !ok {
		return false
	}
	password, ok := credentials[username]
	if !ok {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
		return false
	}
	if username == "admin" {
		ctx := context.WithValue(r.Context(), adminKey, true)
		// Update request with new context-enhanced request
		*r = *r.WithContext(ctx)

	}
	url := r.URL.String()
	validPath := false
	for i, path := range adminPaths[username] {
		if strings.Contains(url, path) {
			validPath = true
			if i == 0 { // Base path can be accessed directly
				return true
			}
		}
	}
	if validPath && username == "admin" {
		return true
	}
	// All other users, check from path to make sure an arena admin isn't trying to make changes in a different arena
	validFromPath := false
	for _, path := range adminPaths[username] {
		if strings.Contains(r.Referer(), path) {
			validFromPath = true
		}
	}

	return validFromPath
}

func basicAuthWrapper(wrapped http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if checkBasicAuthCredentials(r) {
			wrapped(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, "Please login"))
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
	}
}
