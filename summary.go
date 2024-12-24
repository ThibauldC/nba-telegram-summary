package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Response struct {
	ResultSets []ResultSet `json:"resultSets"`
}

type ResultSet struct {
	Name    string          `json:"name"`
	Headers []string        `json:"headers"`
	RowSet  [][]interface{} `json:"rowSet"` // Use [][]interface{} for mixed data types in rows
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
	//fmt.Println(games_response.RowSet[0])
	//scores := []boxscore
	game_id := games_response.RowSet[0][4]
	client := http.Client{}
	game_url := fmt.Sprintf("https://stats.nba.com/stats/boxscoresummaryv2?GameID=%s", game_id)
	//fmt.Println(game_id)
	req, _ := http.NewRequest("GET", game_url, nil)

	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)

	if err != nil {

	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var parsedResponse Response
	_ = json.Unmarshal(body, &parsedResponse)
	//fmt.Println(parsedResponse.ResultSets)
	for _, set := range parsedResponse.ResultSets {
		if set.Name == "LineScore" {
			home_team_stats := set.RowSet[0]
			away_team_stats := set.RowSet[1]
			home_team, home_score := home_team_stats[4], int(home_team_stats[len(home_team_stats)-1].(float64))
			away_team, away_score := away_team_stats[4], int(away_team_stats[len(away_team_stats)-1].(float64))
			fmt.Printf("Score: %s vs %s: %d - %d\n", home_team, away_team, home_score, away_score)
		}
	}
	// for _, element := range games_response.RowSet {
	// 	game_id := element[4]
	// 	game_url := fmt.Sprintf("https://stats.nba.com/stats/boxscoresummaryv2?GameID=%s", game_id)
	// 	//fmt.Println(game_id)
	// 	req, _ := http.NewRequest("GET", game_url, nil)

	// 	for k, v := range headers {
	// 		req.Header.Add(k, v)
	// 	}
	// 	resp, err := client.Do(req)

	// 	if err != nil {

	// 	}
	// 	defer resp.Body.Close()
	// 	body, _ := io.ReadAll(resp.Body)
	// 	var parsedResponse Response
	// 	_ = json.Unmarshal(body, &parsedResponse)
	//}
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

	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var parsedResponse Response
	_ = json.Unmarshal(body, &parsedResponse)

	return parsedResponse.ResultSets[0]
}
