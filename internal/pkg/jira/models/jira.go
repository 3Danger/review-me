package models

import (
	"fmt"
	"time"
)

type Jira struct {
	Expand string `json:"expand"`
	Id     string `json:"id"`
	Self   string `json:"self"`
	Key    string `json:"key"`
	Fields struct {
		Issuetype struct {
			Self        string `json:"self"`
			Id          string `json:"id"`
			Description string `json:"description"`
			IconUrl     string `json:"iconUrl"`
			Name        string `json:"name"`
			Subtask     bool   `json:"subtask"`
			AvatarId    int    `json:"avatarId"`
		} `json:"issuetype"`
		Timespent interface{} `json:"timespent"`
		Project   struct {
			Self           string `json:"self"`
			Id             string `json:"id"`
			Key            string `json:"key"`
			Name           string `json:"name"`
			ProjectTypeKey string `json:"projectTypeKey"`
			AvatarUrls     struct {
				X48 string `json:"48x48"`
				X24 string `json:"24x24"`
				X16 string `json:"16x16"`
				X32 string `json:"32x32"`
			} `json:"avatarUrls"`
		} `json:"project"`
		Customfield10230   interface{}   `json:"customfield_10230"`
		FixVersions        []interface{} `json:"fixVersions"`
		Customfield10110   interface{}   `json:"customfield_10110"`
		Customfield10231   interface{}   `json:"customfield_10231"`
		Customfield10111   interface{}   `json:"customfield_10111"`
		Aggregatetimespent interface{}   `json:"aggregatetimespent"`
		Resolution         interface{}   `json:"resolution"`
		Customfield10233   interface{}   `json:"customfield_10233"`
		Customfield10105   string        `json:"customfield_10105"`
		Customfield10106   float64       `json:"customfield_10106"`
		Customfield10107   struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"customfield_10107"`
		Customfield10503 struct {
			Self     string `json:"self"`
			Value    string `json:"value"`
			Id       string `json:"id"`
			Disabled bool   `json:"disabled"`
		} `json:"customfield_10503"`
		Customfield10108 interface{} `json:"customfield_10108"`
		Customfield10229 interface{} `json:"customfield_10229"`
		Customfield10900 interface{} `json:"customfield_10900"`
		Customfield10109 interface{} `json:"customfield_10109"`
		Resolutiondate   interface{} `json:"resolutiondate"`
		Workratio        int         `json:"workratio"`
		LastViewed       string      `json:"lastViewed"`
		Watches          struct {
			Self       string `json:"self"`
			WatchCount int    `json:"watchCount"`
			IsWatching bool   `json:"isWatching"`
		} `json:"watches"`
		Created          Time     `json:"created"`
		Customfield10100 []string `json:"customfield_10100"`
		Priority         struct {
			Self    string `json:"self"`
			IconUrl string `json:"iconUrl"`
			Name    string `json:"name"`
			Id      string `json:"id"`
		} `json:"priority"`
		Customfield10221 struct {
			Self     string `json:"self"`
			Value    string `json:"value"`
			Id       string `json:"id"`
			Disabled bool   `json:"disabled"`
		} `json:"customfield_10221"`
		Customfield10101              string        `json:"customfield_10101"`
		Labels                        []interface{} `json:"labels"`
		Customfield10216              interface{}   `json:"customfield_10216"`
		Customfield11900              interface{}   `json:"customfield_11900"`
		Customfield10217              interface{}   `json:"customfield_10217"`
		Customfield11902              interface{}   `json:"customfield_11902"`
		Timeestimate                  interface{}   `json:"timeestimate"`
		Aggregatetimeoriginalestimate interface{}   `json:"aggregatetimeoriginalestimate"`
		Versions                      []interface{} `json:"versions"`
		Customfield11901              interface{}   `json:"customfield_11901"`
		Customfield11903              interface{}   `json:"customfield_11903"`
		Issuelinks                    []interface{} `json:"issuelinks"`
		Assignee                      struct {
			Self         string `json:"self"`
			Name         string `json:"name"`
			Key          string `json:"key"`
			EmailAddress string `json:"emailAddress"`
			AvatarUrls   struct {
				X48 string `json:"48x48"`
				X24 string `json:"24x24"`
				X16 string `json:"16x16"`
				X32 string `json:"32x32"`
			} `json:"avatarUrls"`
			DisplayName string `json:"displayName"`
			Active      bool   `json:"active"`
			TimeZone    string `json:"timeZone"`
		} `json:"assignee"`
		Updated string `json:"updated"`
		Status  struct {
			Self           string `json:"self"`
			Description    string `json:"description"`
			IconUrl        string `json:"iconUrl"`
			Name           string `json:"name"`
			Id             string `json:"id"`
			StatusCategory struct {
				Self      string `json:"self"`
				Id        int    `json:"id"`
				Key       string `json:"key"`
				ColorName string `json:"colorName"`
				Name      string `json:"name"`
			} `json:"statusCategory"`
		} `json:"status"`
		Components           []interface{} `json:"components"`
		Timeoriginalestimate interface{}   `json:"timeoriginalestimate"`
		Description          string        `json:"description"`
		Customfield10211     string        `json:"customfield_10211"`
		Timetracking         struct {
		} `json:"timetracking"`
		Archiveddate          interface{}   `json:"archiveddate"`
		Customfield10600      interface{}   `json:"customfield_10600"`
		Attachment            []interface{} `json:"attachment"`
		Customfield10207      interface{}   `json:"customfield_10207"`
		Aggregatetimeestimate interface{}   `json:"aggregatetimeestimate"`
		Summary               string        `json:"summary"`
		Creator               struct {
			Self         string `json:"self"`
			Name         string `json:"name"`
			Key          string `json:"key"`
			EmailAddress string `json:"emailAddress"`
			AvatarUrls   struct {
				X48 string `json:"48x48"`
				X24 string `json:"24x24"`
				X16 string `json:"16x16"`
				X32 string `json:"32x32"`
			} `json:"avatarUrls"`
			DisplayName string `json:"displayName"`
			Active      bool   `json:"active"`
			TimeZone    string `json:"timeZone"`
		} `json:"creator"`
		Subtasks []struct {
			Id     string `json:"id"`
			Key    string `json:"key"`
			Self   string `json:"self"`
			Fields struct {
				Summary string `json:"summary"`
				Status  struct {
					Self           string `json:"self"`
					Description    string `json:"description"`
					IconUrl        string `json:"iconUrl"`
					Name           string `json:"name"`
					Id             string `json:"id"`
					StatusCategory struct {
						Self      string `json:"self"`
						Id        int    `json:"id"`
						Key       string `json:"key"`
						ColorName string `json:"colorName"`
						Name      string `json:"name"`
					} `json:"statusCategory"`
				} `json:"status"`
				Priority struct {
					Self    string `json:"self"`
					IconUrl string `json:"iconUrl"`
					Name    string `json:"name"`
					Id      string `json:"id"`
				} `json:"priority"`
				Issuetype struct {
					Self        string `json:"self"`
					Id          string `json:"id"`
					Description string `json:"description"`
					IconUrl     string `json:"iconUrl"`
					Name        string `json:"name"`
					Subtask     bool   `json:"subtask"`
					AvatarId    int    `json:"avatarId"`
				} `json:"issuetype"`
			} `json:"fields"`
		} `json:"subtasks"`
		Reporter struct {
			Self         string `json:"self"`
			Name         string `json:"name"`
			Key          string `json:"key"`
			EmailAddress string `json:"emailAddress"`
			AvatarUrls   struct {
				X48 string `json:"48x48"`
				X24 string `json:"24x24"`
				X16 string `json:"16x16"`
				X32 string `json:"32x32"`
			} `json:"avatarUrls"`
			DisplayName string `json:"displayName"`
			Active      bool   `json:"active"`
			TimeZone    string `json:"timeZone"`
		} `json:"reporter"`
		Customfield10000  string `json:"customfield_10000"`
		Aggregateprogress struct {
			Progress int `json:"progress"`
			Total    int `json:"total"`
		} `json:"aggregateprogress"`
		Customfield10200 interface{} `json:"customfield_10200"`
		Customfield10201 interface{} `json:"customfield_10201"`
		Environment      interface{} `json:"environment"`
		Duedate          string      `json:"duedate"`
		Progress         struct {
			Progress int `json:"progress"`
			Total    int `json:"total"`
		} `json:"progress"`
		Comment struct {
			Comments   []interface{} `json:"comments"`
			MaxResults int           `json:"maxResults"`
			Total      int           `json:"total"`
			StartAt    int           `json:"startAt"`
		} `json:"comment"`
		Votes struct {
			Self     string `json:"self"`
			Votes    int    `json:"votes"`
			HasVoted bool   `json:"hasVoted"`
		} `json:"votes"`
		Worklog struct {
			StartAt    int           `json:"startAt"`
			MaxResults int           `json:"maxResults"`
			Total      int           `json:"total"`
			Worklogs   []interface{} `json:"worklogs"`
		} `json:"worklog"`
		Archivedby interface{} `json:"archivedby"`
	} `json:"fields"`
}

type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(data []byte) error {
	tm, err := time.ParseInLocation(`"2006-01-02T15:04:05.000-0700"`, string(data), time.Local)
	if err != nil {
		return fmt.Errorf("unmarshalling json: %w", err)
	}

	t.Time = tm

	return nil
}
