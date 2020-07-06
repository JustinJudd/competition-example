package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/justinjudd/competition/models"
)

type PlayerScore struct {
	models.Player
	Score int
	Place int
}

type playerScore struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	Score int    `json:"score"`
	Place int    `json:"place"`
}

func (ps PlayerScore) MarshalJSON() ([]byte, error) {
	p := playerScore{ps.GetName(), fmt.Sprintf("%s", ps.GetMetadata()), ps.Score, ps.Place}
	return json.Marshal(p)
}

type Rankings []*PlayerScore

type Leaderboard struct {
	Scores map[string]*PlayerScore
}

func NewLeaderboard(players []models.Player, baseScore int) Leaderboard {
	l := Leaderboard{}
	l.Scores = make(map[string]*PlayerScore)
	for _, player := range players {
		l.Scores[player.GetName()] = &PlayerScore{player, baseScore, 0}
	}

	return l
}

func (l Leaderboard) GetRankings() Rankings {
	rankings := make(Rankings, 0, len(l.Scores))
	for _, score := range l.Scores {
		rankings = append(rankings, score)
	}

	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].Score > rankings[j].Score
	})

	totalScoreCounts := map[int]int{}

	for _, score := range rankings {
		totalScoreCounts[score.Score]++
	}

	prevPlace := 0
	prevScore := -1
	placeJump := 1

	for _, place := range rankings {
		if totalScoreCounts[place.Score] > 1 && prevScore == place.Score {
			// This player has a tied score
			place.Place = prevPlace
			placeJump = totalScoreCounts[place.Score]
		} else {
			prevPlace += placeJump
			place.Place = prevPlace
			prevScore = place.Score
			placeJump = 1
		}
		l.Scores[place.Player.GetName()] = place
	}

	return rankings
}

func (l Leaderboard) GetPlayerScore(p models.Player) *PlayerScore {
	return l.Scores[p.GetName()]
}

func (l Leaderboard) AdjustPlayerScore(p models.Player, change int) {
	score, ok := l.Scores[p.GetName()]
	if ok {
		score.Score += change
	} else {
		fmt.Println("Unable to find player", p.GetName(), "to update score")
	}
	l.Scores[p.GetName()] = score
}
