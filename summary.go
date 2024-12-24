package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Response struct {
	ResultSets []ResultSet `json:"resultSets"`
}

type ResultSet struct {
	Name    string          `json:"name"`
	Headers []string        `json:"headers"`
	RowSet  [][]interface{} `json:"rowSet"`
}

type boxscore struct {
	hometeam  string
	awayteam  string
	homescore uint8
	awayscore uint8
}

var headers = map[string]string{
	"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"x-nba-stats-origin": "stats",
	"x-nba-stats-token":  "true",
	"Referer":            "https://www.nba.com",
	"Origin":             "https://www.nba.com",
	"Accept-Language":    "en-US,en;q=0.9",
}

//Array [String("SEASON_ID"), String("TEAM_ID"), String("TEAM_ABBREVIATION"), String("TEAM_NAME"), String("GAME_ID"), String("GAME_DATE"), String("MATCHUP"), String("WL"), String("MIN"), String("PTS"), String("FGM"), String("FGA"), String("FG_PCT"), String("FG3M"), String("FG3A"), String("FG3_PCT"), String("FTM"), String("FTA"), String("FT_PCT"), String("OREB"), String("DREB"), String("REB"), String("AST"), String("STL"), String("BLK"), String("TOV"), String("PF"), String("PLUS_MINUS")]
// Array [String("22024"), Number(1610612754), String("IND"), String("Indiana Pacers"), String("0022400247"), String("2024-11-18"), String("IND @ TOR"), String("L"), Number(240), Number(119), Number(44), Number(99), Number(0.444), Number(15), Number(40), Number(0.375), Number(16), Number(20), Number(0.8), Number(12), Number(20), Number(32), Number(29), Number(17), Number(5), Number(10), Number(19), Number(-11.0)]

func main() {
	yesterday := time.Now().Add(time.Duration(-24) * time.Hour).Format("01/02/2006")
	games_response := get_games(yesterday)
	var scores []boxscore
	var game_ids []string
	for _, game := range games_response.RowSet {
		game_id := game[4].(string)
		if !slices.Contains(game_ids, game_id) {
			game_ids = append(game_ids, game_id)
		}
	}
	for _, game_id := range game_ids {
		box_score := get_game_score(game_id)
		scores = append(scores, box_score)
	}

	var score_strings []string
	for _, score := range scores {
		score_strings = append(score_strings, fmt.Sprintf("%s vs %s: %d - %d", score.hometeam, score.awayteam, score.homescore, score.awayscore))
	}

	all_scores := strings.Join(score_strings, "\n")
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	chat_id, _ := strconv.Atoi(os.Getenv("CHAT_ID"))
	msg := tgbotapi.NewMessage(int64(chat_id), all_scores)

	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

}

func get_game_score(game_id string) boxscore {
	client := http.Client{}
	game_url := fmt.Sprintf("https://stats.nba.com/stats/boxscoresummaryv2?GameID=%s", game_id)
	req, _ := http.NewRequest("GET", game_url, nil)

	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var parsedResponse Response
	_ = json.Unmarshal(body, &parsedResponse)

	var home_team, away_team string
	var home_score, away_score uint8
	for _, set := range parsedResponse.ResultSets {
		if set.Name == "LineScore" {
			home_team_stats := set.RowSet[0]
			away_team_stats := set.RowSet[1]
			home_team, home_score = home_team_stats[4].(string), uint8(home_team_stats[len(home_team_stats)-1].(float64))
			away_team, away_score = away_team_stats[4].(string), uint8(away_team_stats[len(away_team_stats)-1].(float64))
		}
	}
	return boxscore{hometeam: home_team, awayteam: away_team, homescore: home_score, awayscore: away_score}
}

func get_games(date string) ResultSet {
	client := http.Client{}
	url := fmt.Sprintf("https://stats.nba.com/stats/leaguegamefinder?playerOrTeam=T&LeagueID=00&dateFrom=%s", date)
	fmt.Println(url)
	req, _ := http.NewRequest("GET", url, nil)

	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var parsedResponse Response
	_ = json.Unmarshal(body, &parsedResponse)

	return parsedResponse.ResultSets[0]
}
