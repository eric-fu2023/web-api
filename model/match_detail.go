package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type BetFbSport struct {
	ploutos.BetFb
}

type BetImsbSport struct {
	ploutos.BetImsb
}

type BetReport struct {
	ploutos.BetReport
}

type Bet interface {
	ploutos.Bet
}

type FbMatchDetailRequest struct {
	MatchId      string `json:"matchId"`
	CurrencyId   int64  `json:"currencyId"`
	OddsType     int64  `json:"oddsType"`
	LanguageType string `json:"languageType"`
}

type FbMatchAPIResponse struct {
	Code  int             `json:"code"`
	Data  FbMatchResponse `json:"data,omitempty"`
	Msg   string          `json:"msg"`
	Error string          `json:"error,omitempty"`
}

type FbMatchResponse struct {
	Ts []FbMatchTeam `json:"ts"`
	Lg FbLeague      `json:"lg"`
	Bt int64         `json:"bt"`
}

type FbMatchTeam struct {
	Id   int64  `json:"id"`
	Lurl string `json:"lurl"`
	Na   string `json:"na"`
}

type FbLeague struct {
	Na   string `json:"na"`
	Lurl string `json:"lurl"`
}

type ImsbMatchDetailRequest struct {
	MatchIds []int64 `json:"ids"`
}

type ImsbMatchAPIResponse struct {
	Code  int               `json:"code"`
	Data  ImsbMatchResponse `json:"data,omitempty"`
	Msg   string            `json:"msg"`
	Error string            `json:"error,omitempty"`
}

type ImsbMatchResponse struct {
	Matches []ImsbMatchDetailResponse `json:"matches"`
}

type ImsbMatchDetailResponse struct {
	Id          int64                   `json:"id"`
	Title       string                  `json:"title"`
	TeamA       ImsbTeamAResponse       `json:"teama"`
	TeamB       ImsbTeamBResponse       `json:"teamb"`
	Competition ImsbCompetitionResponse `json:"competition"`
}

type ImsbTeamAResponse struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	LogoUrl string `json:"logo_url"`
}
type ImsbTeamBResponse struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	LogoUrl string `json:"logo_url"`
}
type ImsbCompetitionResponse struct {
	Id     int64  `json:"id"`
	Title  string `json:"title"`
	Format string `json:"format"`
}

// func (a Bet) GetMoreBetReportDetails() {
// 	// Get Teams Icon, Name, League Name

// 	if s, ok := a.(BetFbSport); ok {
// 		fmt.Print("aa")
// 	}

// }

// BetFb
func GetFbMatchDetails(matchId int64) (match FbMatchResponse, err error) {

	request := FbMatchDetailRequest{
		MatchId:      fmt.Sprint(int(matchId)),
		CurrencyId:   2,
		LanguageType: "CMN",
		OddsType:     1,
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}
	tayaUrl, _ := GetAppConfig("taya_url", "apiServerAddress")
	url := tayaUrl + "/v1/match/getMatchDetail"
	log.Printf("GET MATCH DETAIL FROM TAYA URL=%v commonNoAuth.GetMatchDetail err - %v \n", url, err)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	var matchDetail FbMatchAPIResponse
	err = json.Unmarshal(body, &matchDetail)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	match = matchDetail.Data
	return
}

// BetImsb
func GetImsbMatchDetails(matchIds ...int64) (matches []ImsbMatchDetailResponse, err error) {

	request := ImsbMatchDetailRequest{
		MatchIds: matchIds,
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}
	bataceDomainUrl, _ := GetAppConfig("server_url", "domain")
	url := bataceDomainUrl + "index/v1/batace/im_match/fetch"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	var matchDetail ImsbMatchAPIResponse
	err = json.Unmarshal(body, &matchDetail)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	matches = matchDetail.Data.Matches
	return
}
