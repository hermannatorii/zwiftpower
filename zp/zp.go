package zp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type club struct {
	Data []Rider
}

// Rider shows data about a rider
type Rider struct {
	Name             string
	Zwid             int
	LatestEventDate  time.Time
	Rides            int
	Races            int
	Races90          int
	Races30          int
	Ftp90            float64
	Ftp60            float64
	Ftp30            float64
	LatestRace       string
	LatestRaceDate   time.Time
	LatestEvent      string
	LatestRaceAvgWkg float64
	LatestRaceWkgFtp float64
}

type riderData struct {
	Data []Event
}

// Event is a ZwiftPower event
type Event struct {
	EventType     string `json:"f_t"`
	EventDateSecs int64  `json:"event_date"`
	EventDate     time.Time
	EventTitle    string      `json:"event_title"`
	AvgWkg        interface{} `json:"avg_wkg"`
	WkgFtp        interface{} `json:"wkg_ftp"`
}

// ImportZP imports data about the club with this ID
func ImportZP(clubID int) []Rider {
	data, err := getJSON(fmt.Sprintf("https://www.zwiftpower.com/api3.php?do=team_riders&id=%d", clubID))
	if err != nil {
		log.Fatal(err)
	}

	var c club
	err = json.Unmarshal(data, &c)
	if err != nil {
		log.Fatal(err)
	}

	return c.Data
}

// ImportRider imports data about the rider with this ID
func ImportRider(riderID int) (rider Rider, err error) {
	// I think hitting the profile URL loads the data into the cache
	_, _ = http.Get(fmt.Sprintf("https://www.zwiftpower.com/profile.php?z=%d", riderID))
	data, err := getJSON(fmt.Sprintf("https://www.zwiftpower.com/cache3/profile/%d_all.json", riderID))
	if err != nil {
		return rider, err
	}

	var r riderData
	err = json.Unmarshal(data, &r)
	if err != nil {
		log.Print(string(data))
		return rider, err
	}

	rider.Zwid = riderID
	if len(r.Data) < 1 {
		log.Printf("No event data for rider %d", riderID)
		return rider, nil
	}

	var latestEventDate time.Time
	var latestRaceDate time.Time
	for _, e := range r.Data {
		e.EventDate = time.Unix(e.EventDateSecs, 0)
		daysAgo := int(time.Now().Sub(e.EventDate).Hours() / 24)
		// log.Printf("date %v, from %v is %d days ago\n", e.EventDate, e.EventDateSecs, daysAgo)
		isRace := strings.Contains(e.EventType, "RACE")

		rider.Rides++
		if isRace {
			rider.Races++
		}

		var wkgFtp float64
		var avgWkg float64

		eventWkgFtp := e.WkgFtp.([]interface{})
		wkgFtp, ok := eventWkgFtp[0].(float64)
		if !ok {
			wkgFtp, err = strconv.ParseFloat(eventWkgFtp[0].(string), 64)
			if err != nil {
				log.Fatal(err)
			}
		}

		avgWkg, err = strconv.ParseFloat(e.AvgWkg.([]interface{})[0].(string), 64)
		if err != nil {
			log.Fatal(err)
		}

		// Last three months?
		if daysAgo <= 90 {
			if wkgFtp > rider.Ftp90 {
				rider.Ftp90 = wkgFtp
			}

			if isRace {
				rider.Races90++
			}
		}

		// Last two months?
		if daysAgo <= 60 {
			if wkgFtp > rider.Ftp60 {
				rider.Ftp60 = wkgFtp
			}
		}

		// Last month?
		if daysAgo <= 30 {
			if isRace {
				rider.Races30++
			}

			if wkgFtp > rider.Ftp30 {
				rider.Ftp30 = wkgFtp
			}
		}

		if e.EventDate.After(latestEventDate) {
			latestEventDate = e.EventDate
			rider.LatestEvent = e.EventTitle
		}

		if isRace && e.EventDate.After(latestRaceDate) {
			latestRaceDate = e.EventDate
			rider.LatestRace = e.EventTitle
			rider.LatestRaceAvgWkg = avgWkg
			rider.LatestRaceWkgFtp = wkgFtp
		}
	}

	rider.LatestEventDate = latestEventDate
	rider.LatestRaceDate = latestRaceDate
	return rider, nil
}

func getJSON(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return []byte{}, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

// MonthsAgo describes how many months since the rider's latest event
func (r Rider) MonthsAgo() string {
	if r.LatestEventDate.IsZero() {
		return "No latest event"
	}

	if time.Now().Sub(r.LatestEventDate) > (time.Hour * 24 * 365) {
		return "Over a year ago"
	}

	monthDiff := time.Now().Month() - r.LatestEventDate.Month()
	if monthDiff < 0 {
		monthDiff += 12
	}

	switch monthDiff {
	case 0:
		return "This month"
	case 1:
		return "Last month"
	default:
		return fmt.Sprintf("%d months ago", monthDiff)
	}
}

// Strings turns a rider struct into []string
func (r Rider) Strings() []string {
	output := make([]string, 14)
	output[0] = r.Name
	output[1] = strconv.Itoa(r.Zwid)
	output[2] = r.LatestEventDate.Format("2006-01-02")
	output[3] = r.MonthsAgo()
	output[4] = r.LatestEvent
	output[5] = strconv.Itoa(r.Rides)
	output[6] = fmt.Sprintf("https://www.zwiftpower.com/profile.php?z=%d", r.Zwid)
	output[7] = strconv.FormatFloat(r.Ftp30, 'f', 1, 64)
	output[8] = strconv.FormatFloat(r.Ftp90, 'f', 1, 64)
	output[9] = strconv.Itoa(r.Races30)
	output[10] = strconv.Itoa(r.Races90)
	output[11] = strconv.Itoa(r.Races)
	output[12] = r.LatestRace
	output[13] = r.LatestRaceDate.Format("2006-01-02")
	return output
}
