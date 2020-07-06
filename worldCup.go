package main

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/justinjudd/competition"
	"github.com/justinjudd/competition/models"
	"github.com/justinjudd/competition/tournament"
)

// WorldCup fulfills the Tournament interface. It is a Tournament that consists of up to 2 teams going against each other in a soccer match.
// Consists of two sub-tournaments, the first is a play-in games, similar to the real World Cup.
// The best teams from Group Play will advance to the double elimination stage where they will compete to  determine the very best team
type WorldCup struct {
	stage int
	models.Tournament
	LargerComp  *Tournament
	moveForward bool
}

func (w *WorldCup) MatchLength() time.Duration {
	return 7 * time.Minute
}

func (w *WorldCup) Name() string {
	return "World Cup"
}

func (w *WorldCup) StageID() int {
	return w.stage
}

func (w *WorldCup) StageName() string {
	var stage string
	switch w.stage {
	case 0:
		stage = "Double Elimination"
	}
	return fmt.Sprintf("World Cup: %s", stage)
}

var scoreDeductions = []int{0, 60, 45, 35, 25, 15, 5, 5}

func (w *WorldCup) NextRound() (models.Round, error) {
	rounds := w.GetAllRounds()
	roundCount := len(rounds)

	if roundCount > 0 {
		lastRound := rounds[roundCount-1]

		for _, game := range lastRound.GetGames() {
			if len(game.GetPlaces()) <= 1 {
				continue
			}
			if len(game.GetTeams()) > 1 {
				for i, team := range game.GetTeams() {
					p := competition.FlipTies(int(game.GetPlaces()[i]))
					if p >= 1 { // this team lost
						if models.IsByeTeam(team) {
							continue
						}
						switch w.stage {
						case 0: //Play in rounds
							for _, player := range team.GetPlayers() {
								w.LargerComp.Leaderboard.AdjustPlayerScore(player, -5)
							}

						case 1: // Double elimination
							for _, record := range team.GetRecords()[:len(team.GetRecords())-1] {
								teamIndex := -1
								for j, t := range record.GetTeams() {
									if team.Equals(t) {
										teamIndex = j
									}
								}
								p := competition.FlipTies(int(record.GetPlaces()[teamIndex]))

								if p >= 1 { // Team has lost twice
									deduction := -1 * scoreDeductions[len(team.GetRecords())]
									for _, player := range team.GetPlayers() {
										w.LargerComp.Leaderboard.AdjustPlayerScore(player, deduction)

										if len(team.GetRecords()) >= 3 {
											w.LargerComp.Leaderboard.AdjustPlayerScore(player, 5) // Player gets 5 points for surviving more than 2 games
										}

									}
								}
							}

						}

					}
				}

			}

		}

	}

	rnd, err := w.Tournament.NextRound()
	if err != nil && w.stage == 0 {
		w.NextStage()
		rnd, err = w.Tournament.NextRound()
	}

	return rnd, err
}

func (w *WorldCup) Complete() {
	w.moveForward = true
}

func (w *WorldCup) IsComplete() bool {
	return w.moveForward
}

type groupPlayTeamScore struct {
	Team         models.Team
	Wins         float64
	TotalGoals   int
	GoalsAgainst int
}

type groupPlayTeamScores []groupPlayTeamScore

func (t groupPlayTeamScores) Len() int      { return len(t) }
func (t groupPlayTeamScores) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t groupPlayTeamScores) Less(i, j int) bool {
	f := t[i]
	s := t[j]
	if f.Wins != s.Wins {
		return f.Wins < s.Wins
	}
	if f.TotalGoals != s.TotalGoals {
		return f.TotalGoals < s.TotalGoals
	}

	return f.GoalsAgainst < s.GoalsAgainst
}

func (w *WorldCup) NextStage() error {
	fmt.Println("World Cup Next stage")

	data, err := competition.GenerateTournamentHTML(w)
	if err == nil {
		w.LargerComp.SendEvent("main", string(data))
	}

	switch w.stage {
	case 0: // Play Play-in rounds

		// Top 2 teams from each group advance

		groups := map[string]map[string]groupPlayTeamScore{}

		for _, bracket := range w.Tournament.GetBracketOrder() {
			groups[bracket] = map[string]groupPlayTeamScore{}
		}

		for _, round := range w.Tournament.GetAllRounds() {
			for _, game := range round.GetGames() {
				group := game.GetBracket()
				for i, team := range game.GetTeams() {
					score := groups[group][team.GetName()]
					if score.Team == nil {
						score.Team = team
					}
					switch game.GetTeamPlace(team) {
					case 0: // Won
						score.Wins++
					case -1: // Tied
						score.Wins += .5
					}
					score.TotalGoals += int(game.GetTeamScore(team))
					otherTeamIndex := (i + 1) % 2
					if len(game.GetTeams()) > otherTeamIndex {
						score.GoalsAgainst += int(game.GetTeamScore(game.GetTeams()[otherTeamIndex]))
					}

					groups[group][team.GetName()] = score

				}

			}

		}

		advancing := []models.Team{}

		for _, grouped := range groups {
			scores := groupPlayTeamScores{}
			for _, team := range grouped {
				scores = append(scores, team)

			}
			sort.Sort(sort.Reverse(scores))
			for _, team := range scores[:2] {
				advancing = append(advancing, team.Team)
			}
		}

		ordered := []models.Team{}
		for i := 0; i < 2; i++ {
			for j := 0; j < len(advancing)/2; j++ {
				ordered = append(ordered, advancing[j*2+i])
			}
		}

		w.Tournament = tournament.NewDoubleElimination(w.LargerComp.AddTournament("World Cup: Elimination Rounds", models.TournamentType_DOUBLE_ELIMINATION, ordered, false, 2, 1, true))
		matchTime := 7 * time.Minute
		var buf bytes.Buffer
		gr := gob.NewEncoder(&buf)
		gr.Encode(matchTime)
		w.Tournament.SetMetadata(buf.Bytes())
		w.stage++

		return nil

	case 1:

		return fmt.Errorf("World Cup finished")
	}

	return nil

}

type Team struct {
	Players []models.Player
	Name    string
	Image   []byte
}

func (t *Tournament) WorldCup() models.Tournament {
	t.showScores = true

	groupComps := []models.Tournament{}
	matchTime := 7 * time.Minute
	var buf bytes.Buffer
	gr := gob.NewEncoder(&buf)
	gr.Encode(matchTime)

	teams := []Team{}
	f, err := os.Open("world_cup.csv")
	if err != nil {
		return nil
	}
	defer f.Close()
	r := csv.NewReader(f)
	teamRecords, err := r.ReadAll()
	lastGroup := ""
	for _, row := range teamRecords[1:] {

		if len(lastGroup) > 0 && lastGroup != row[3] {
			comp := tournament.NewRoundRobin(t.AddTournament("World Cup: Group Play:"+lastGroup, models.TournamentType_ROUND_ROBIN, nil, false, 2, 2, true))
			comp.SetMetadata(buf.Bytes())
			for _, team := range teams {
				comp.CreateTeam(team.Name, team.Players, team.Image)
			}
			groupComps = append(groupComps, comp)
			teams = []Team{}
		}
		var players []models.Player
		if len(row[1]) > 0 {
			players = append(players, t.GetPlayer(row[1]))
			if len(row[2]) > 0 { // Has second team as part of this country team
				players = append(players, t.GetPlayer(row[2]))
			}
		}
		team := Team{players, row[0], []byte(countryImage(row[0]))}
		teams = append(teams, team)
		lastGroup = row[3]

	}
	comp := tournament.NewRoundRobin(t.AddTournament("World Cup: Group Play:"+lastGroup, models.TournamentType_ROUND_ROBIN, nil, false, 2, 2, true))
	comp.SetMetadata(buf.Bytes())
	for _, team := range teams {
		comp.CreateTeam(team.Name, team.Players, team.Image)
	}
	groupComps = append(groupComps, comp)

	c := tournament.NewGroupCompetition(groupComps, t.AddTournament("World Cup: Group Play", models.TournamentType_GROUP_PLAY, nil, false, 2, 1, true))
	c.SetMetadata(buf.Bytes())

	fmt.Println("Starting World Cup")

	return &WorldCup{0, c, t, false}

}
