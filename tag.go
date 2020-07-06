package main

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/justinjudd/competition"
	"github.com/justinjudd/competition/models"
	"github.com/justinjudd/competition/tournament"
)

// RobotTag fulfills the Tournament interface. It is a Tournament that consists of up to 4 teams going against each other in a free-for-all tag battle.
// Consists of two sub-tournaments, the first is a compass draw to allow teams to play against each other and determine rankings.
// Next there is a double elimination tournament to determine the very best team
type RobotTag struct {
	stage int
	models.Tournament
	LargerComp  *Tournament
	moveForward bool
}

func (r *RobotTag) MatchLength() time.Duration {
	return 270 * time.Second
}

func (r *RobotTag) Name() string {
	return "Tag"
}

func (r *RobotTag) StageID() int {
	return r.stage
}

func (r *RobotTag) StageName() string {
	var stage string
	switch r.stage {
	case 0:
		stage = "Round Robin"
	case 1:
		stage = "Double Elimination"
	}
	return fmt.Sprintf("Tag Battle: %s", stage)
}

var pointModifier = 2
var startingPoints = 130

var roundScores = []int{13, 8, 4}

func (r *RobotTag) NextRound() (models.Round, error) {
	scoreDeductions = []int{0, 80, 65, 50, 35, 25, 15, 10, 10}

	switch r.stage {
	case 0:
		rounds := r.GetAllRounds()
		roundCount := len(rounds)

		if roundCount > 0 {
			lastRound := rounds[roundCount-1]
			if len(rounds) < len(roundScores) {
				for _, game := range lastRound.GetGames() {
					for _, team := range game.GetTeams() {
						var points int
						if !competition.IsWinner(team, game, r.GetAdvancing()) {
							points -= roundScores[roundCount] * pointModifier
						}

						for _, player := range team.GetPlayers() {

							r.LargerComp.Leaderboard.AdjustPlayerScore(player, points)
						}
					}
				}
			} else { // extra round for the East and West brackets

				teams := r.GetTeams()
				sort.Slice(teams, func(i, j int) bool {
					teamA, teamB := teams[i], teams[j]
					AGames := teamA.GetRecords()
					lastAGame := AGames[len(AGames)-1]
					BGames := teamB.GetRecords()
					lastBGame := BGames[len(BGames)-1]
					lastADivision := 0
					lastBDivision := 0
					for i, div := range tournament.CompassDivisionNames {
						if lastAGame.GetBracket() == div {
							lastADivision = i
							break
						}
						if lastBGame.GetBracket() == div {
							lastBDivision = i
							break
						}
					}
					if lastADivision < lastBDivision { // Team A ended in a higher division
						return true
					}
					for _, recordA := range teamA.GetRecords() {
						aPlace := recordA.GetTeamPlace(teamA)
						bPlace := recordA.GetTeamPlace(teamB)
						aWin := competition.FlipTies(int(aPlace)) < 2
						bWin := competition.FlipTies(int(bPlace)) < 2

						if aWin && !bWin {
							return true
						}
					}
					return competition.FlipTies(int(lastAGame.GetTeamPlace(teamA))) < competition.FlipTies(int(lastBGame.GetTeamPlace(teamB)))

				})
				pointDeduction := 0
				rankedTeams := rankOrderTeams(r.GetTeams())
				for _, team := range rankedTeams {
					for _, player := range team.GetPlayers() {

						r.LargerComp.Leaderboard.AdjustPlayerScore(player, -1*int(pointDeduction*pointModifier))

					}
					pointDeduction++
				}
			}
		}

	case 1:
		rounds := r.GetAllRounds()
		roundCount := len(rounds)

		if roundCount > 1 {
			lastRound := rounds[roundCount-1]
			for _, game := range lastRound.GetGames() {
				if len(game.GetPlaces()) <= 1 {
					continue
				}
				if len(game.GetTeams()) > 2 { // at least one team lost
					for _, team := range game.GetTeams() {

						if !competition.IsWinner(team, game, r.GetAdvancing()) {
							if models.IsByeTeam(team) {
								continue
							}
							for _, record := range team.GetRecords()[:len(team.GetRecords())-1] {
								teamIndex := -1
								for j, t := range record.GetTeams() {
									if team.Equals(t) {
										teamIndex = j
									}
								}
								p := competition.FlipTies(int(record.GetPlaces()[teamIndex]))

								if p >= 1 { // Team has lost twice
									for _, player := range team.GetPlayers() {

										r.LargerComp.Leaderboard.AdjustPlayerScore(player, -1*int(scoreDeductions[len(team.GetRecords())]))
									}
								}
							}
						}
					}

				}

			}

		}

	}

	rnd, err := r.Tournament.NextRound()
	if err != nil && r.stage == 0 {
		r.NextStage()
		rnd, err = r.Tournament.NextRound()
	}

	return rnd, err
}

func (r *RobotTag) Complete() {
	r.moveForward = true
	for _, round := range r.GetAllRounds() {
		round.SetFinal()
	}
}

func (r *RobotTag) IsComplete() bool {
	return r.moveForward
}

func rankOrderTeams(teams []models.Team) []models.Team {
	scores := map[models.Team]int{}
	rankedTeams := make([]models.Team, len(teams))

	for i, team := range teams {
		scores[team] = 0
		rankedTeams[i] = team
	}

	for _, team := range teams {
		for i, record := range team.GetRecords() {
			teamIndex := -1
			for j, t2 := range record.GetTeams() {
				if team.Equals(t2) {
					teamIndex = j
				}
			}
			if i == len(team.GetRecords())-1 {

				scores[team] += 4 - competition.FlipTies(int(record.GetPlaces()[teamIndex]))
			} else {

				switch competition.FlipTies(int(record.GetPlaces()[teamIndex])) {
				case 0, 1:
					scores[team] += 4 * int(math.Pow(float64(4-i), 2.0))
				case 2, 3:
					//scores[team] -= 20
				}
			}
		}

	}

	sort.Slice(rankedTeams, func(i, j int) bool {
		t1, t2 := rankedTeams[i], rankedTeams[j]

		return scores[t2] < scores[t1] // flip so ranked highest first
	})

	return rankedTeams
}

func (r *RobotTag) NextStage() error {
	fmt.Println("Tag: Next Stage")

	data, err := competition.GenerateTournamentHTML(r)
	if err == nil {
		r.LargerComp.SendEvent("main", string(data))
	}
	switch r.stage {
	case 0:

		rankedTeams := rankOrderTeams(r.GetTeams())

		// Set up next competition

		r.Tournament = tournament.NewDoubleElimination(r.LargerComp.AddTournament("Tag", models.TournamentType_DOUBLE_ELIMINATION, rankedTeams, true, 4, 1, false))
		matchTime := 270 * time.Second
		var buf bytes.Buffer
		gr := gob.NewEncoder(&buf)
		gr.Encode(matchTime)
		r.Tournament.SetMetadata(buf.Bytes())
		r.stage++

		return nil

	case 1:

		for _, team := range r.GetTeams() {
			lastGame := team.GetRecords()[len(team.GetRecords())-1]
			p := lastGame.GetTeamPlace(team)
			place := competition.FlipTies(int(p))
			for _, player := range team.GetPlayers() {
				r.LargerComp.Leaderboard.AdjustPlayerScore(player, -1*int(place))
			}
		}

		return fmt.Errorf("Tag finished")

	}

	return nil

}

func (t *Tournament) RobotTag() models.Tournament {
	t.showScores = false

	cd := tournament.NewCompassDraw(t.AddTournament("Tag: Compass Draw", models.TournamentType_COMPASS_DRAW, nil, false, 4, 2, false))
	matchTime := 270 * time.Second
	var buf bytes.Buffer
	gr := gob.NewEncoder(&buf)
	gr.Encode(matchTime)
	cd.SetMetadata(buf.Bytes())

	var teams []models.Team
	f, err := os.Open("tag.csv")
	if err != nil {
		fmt.Println("Unable to open teams csv file:", err)
		return nil
	}
	defer f.Close()
	r := csv.NewReader(f)
	teamRecords, err := r.ReadAll()
	for _, row := range teamRecords[1:] {
		var players []models.Player
		if len(row[0]) > 0 {
			players = append(players, t.GetPlayer(row[0]))

		}
		team := cd.CreateTeam(row[0], players, []byte{})
		teams = append(teams, team)
	}

	competition.RandomizeTeams(teams)

	rt := RobotTag{0, cd, t, false}

	fmt.Println("Starting Tag battle")

	return &rt

}
