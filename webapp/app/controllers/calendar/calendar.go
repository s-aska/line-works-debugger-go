package calendar

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/luxifer/ical"
	"github.com/s-aska/line-works-debugger-go/contrlib/perm"
	"github.com/teambition/rrule-go"
	"google.golang.org/appengine/urlfetch"
)

var jst = time.FixedZone("Asia/Tokyo", 60*60*9)
var utc = time.FixedZone("UTC", 0)

type ScheduleListResp struct {
	Result      string     `json:"result"`
	ReturnValue []Schedule `json:"returnValue"`
}

type Schedule struct {
	UserID     string `json:"userId"`
	CalendarID string `json:"calendarId"`
	Ical       string `json:"ical"`
	ViewURL    string `json:"viewUrl"`
}

type MyEvent struct {
	UID          string
	Source       string
	Location     string
	Created      time.Time
	LastModified time.Time
	StartDate    time.Time
	EndDate      time.Time
	RuleDates    []time.Time
	SummaryDate  string
	Summary      string
	Description  string
	Organizer    string
	ViewURL      string
	Raw          string
}

func Events(c echo.Context) error {
	r := c.Request()
	ctx := r.Context()

	from := r.FormValue("from")
	to := r.FormValue("to")

	session := perm.LoadSession(r)
	if appID := r.FormValue("appID"); appID != "" {
		session.AppID = appID
	}
	if consumerKey := r.FormValue("consumerKey"); consumerKey != "" {
		session.ConsumerKey = consumerKey
	}
	if accessToken := r.FormValue("accessToken"); accessToken != "" {
		session.AccessToken = accessToken
	}

	name := "明日"
	if n := r.FormValue("name"); n != "" {
		name = n
	}

	values := url.Values{}
	values.Add("rangeDateFrom", from)
	values.Add("rangeDateUntil", to)
	if calendarId := r.FormValue("calendarId"); calendarId != "" {
		values.Add("calendarId", calendarId)
	}
	req, err := http.NewRequest("POST",
		fmt.Sprintf("https://apis.worksmobile.com/%s/calendar/getScheduleList/V3",
			session.AppID),
		strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("consumerKey", session.ConsumerKey)
	req.Header.Set("Authorization", "Bearer "+session.AccessToken)

	client := urlfetch.Client(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// dumpReq, _ := httputil.DumpRequest(req, true)
	// log.Debugf(ctx, "%s", dumpReq)
	//
	// dumpResp, _ := httputil.DumpResponse(resp, true)
	// log.Debugf(ctx, "%s", dumpResp)

	jsonResp := &ScheduleListResp{}
	err = json.NewDecoder(resp.Body).Decode(jsonResp)
	if err != nil {
		return err
	}

	events := make([]*MyEvent, 0)
	for _, schedule := range jsonResp.ReturnValue {
		calendar, err := ical.Parse(strings.NewReader(schedule.Ical), nil)
		if err != nil {
			return err
		}

		for _, event := range calendar.Events {
			me := &MyEvent{}
			me.ViewURL = schedule.ViewURL
			me.Summary = event.Summary
			me.Description = strings.Replace(event.Description, "\\n", "\n", -1)
			me.StartDate = event.StartDate
			me.EndDate = event.EndDate

			for _, prop := range event.Properties {
				switch prop.Name {
				case "UID":
					me.UID = prop.Value
				case "LOCATION":
					me.Location = prop.Value
				case "CREATED":
					if t, err := time.Parse("20060102T150405Z", prop.Value); err == nil {
						me.Created = t
					}
				case "LAST-MODIFIED":
					if t, err := time.Parse("20060102T150405Z", prop.Value); err == nil {
						me.LastModified = t
					}
				case "RRULE":
					option, err := rrule.StrToROptionInLocation(prop.Value, jst)
					if err != nil {
						return err
					}
					option.Dtstart = me.StartDate
					option.Until = option.Until.In(jst)
					rule, err := rrule.NewRRule(*option)
					if err != nil {
						return err
					}
					me.RuleDates = rule.All()
				case "ORGANIZER":
					for key, param := range prop.Params {
						if key == "CN" {
							for _, v := range param.Values {
								me.Organizer = v
							}
						}
					}
				}
			}

			if len(me.RuleDates) > 0 {
				for _, date := range me.RuleDates {
					if date.Format("20060102") >= from && date.Format("20060102") <= to {
						duration := me.EndDate.Sub(me.StartDate)
						sd := fmt.Sprintf("%sT%s00", date.Format("20060102"), me.StartDate.Format("1504"))
						me.StartDate, _ = time.ParseInLocation("20060102T150405", sd, jst)
						me.EndDate = me.StartDate.Add(duration)
						// log.Debugf(ctx, "    RuleDate:%s from:%s to:%s pass", date.Format("20060102"), from, to)
						break
					}
					// log.Debugf(ctx, "    RuleDate:%s from:%s to:%s skip", date.Format("20060102"), from, to)
				}
			}
			if me.StartDate.Format("20060102") == me.StartDate.Format("20060102") {
				me.SummaryDate = fmt.Sprintf("%s〜%s",
					me.StartDate.Format("15:04"),
					me.EndDate.Format("15:04"))
			} else {
				me.SummaryDate = fmt.Sprintf("%s〜\n%s",
					me.StartDate.Format("2006/01/02 15:04"),
					me.EndDate.Format("2006/01/02 15:04"))
			}

			events = append(events, me)
			break
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].StartDate.Before(events[j].StartDate)
	})

	// Send to LINE
	if lineToken := r.FormValue("lineToken"); lineToken != "" {
		message := name + "の予定\n"
		for _, me := range events {
			message += "■ " + me.SummaryDate + "\n"
			message += me.Summary + "\n"
			if me.Location != "" {
				message += me.Location + "\n"
			}
			message += "\n"
		}
		values := url.Values{}
		values.Add("message", message)
		req, err := http.NewRequest("POST",
			"https://notify-api.line.me/api/notify",
			strings.NewReader(values.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "Bearer "+lineToken)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return c.String(http.StatusOK, "OK")
	}

	data := map[string]interface{}{}
	data["req"] = r
	data["session"] = session
	data["events"] = events
	return c.Render(http.StatusOK, "calendar/index.html", data)
}
