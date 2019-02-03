package iarapi

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
)

var (
	// ErrIarAuthentication is an error thrown when IAR does not authenticate
	// a query
	ErrIarAuthentication = errors.New("Unable to authenticate to IAR due to incorrect credentials")
)

// IamRespondingAPI represents an API access object
type IamRespondingAPI struct {
	// Debug sets whether debugging output is enabled
	Debug bool

	browserObject   *browser.Browser
	initialized     bool
	orgCrypted      string
	member          int64
	memberCrypted   string
	subscriber      int64
	adminCrypted    string
	agency          int64
	agencyCrypted   string
	apiToken        string
	clientIarAPIURL string
	tokenForAPI     string
}

func (c *IamRespondingAPI) Login(agency, user, pass string) error {
	b := surf.NewBrowser()
	c.browserObject = b

	b.SetUserAgent(agent.Chrome())

	// Required to not have ASP.NET garbage yak all over me
	b.AddRequestHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")

	loginParams := map[string]interface{}{
		"memberLogin":     "true",
		"agencyName":      agency,
		"memberfname":     user,
		"memberpwd":       pass,
		"rememberPwd":     false,
		"urlTo":           "",
		"overrideSession": true,
	}
	post, err := json.Marshal(loginParams)
	if err != nil {
		return err
	}

	err = b.Post("https://iamresponding.com/v3/Pages/memberlogin.aspx/ValidateLoginInfo", "application/json", strings.NewReader(string(post)))
	if err != nil {
		return err
	}
	// Catch authentication issues
	if c.Debug {
		log.Printf("Login(): %d, %q", strings.Index(b.Body(), "The log-in information that you have entered is incorrect."), b.Body())
	}
	if strings.Index(b.Body(), "The log-in information that you have entered is incorrect.") > -1 {
		return ErrIarAuthentication
	}

	v := url.Values{}
	v.Set("username", user)
	v.Set("password", pass)
	v.Set("email", "")
	err = b.PostForm("https://iamresponding.com/v3/agency/def.aspx", v)
	if err != nil {
		return err
	}

	// Parse out tokens
	b.Dom().Find("INPUT#orgCrypted").Each(func(_ int, s *goquery.Selection) {
		c.orgCrypted, _ = s.Attr("value")
	})
	b.Dom().Find("INPUT#memberCrypted").Each(func(_ int, s *goquery.Selection) {
		c.memberCrypted, _ = s.Attr("value")
	})
	b.Dom().Find("INPUT#hdnIsAgencyCrypted").Each(func(_ int, s *goquery.Selection) {
		c.agencyCrypted, _ = s.Attr("value")
	})
	b.Dom().Find("INPUT#hdnIsAdminCrypted").Each(func(_ int, s *goquery.Selection) {
		c.adminCrypted, _ = s.Attr("value")
	})
	b.Dom().Find("INPUT#TokenForApi").Each(func(_ int, s *goquery.Selection) {
		c.tokenForAPI, _ = s.Attr("value")
	})
	b.Dom().Find("INPUT#ClientIarApiUrl").Each(func(_ int, s *goquery.Selection) {
		c.clientIarAPIURL, _ = s.Attr("value")
	})

	// Get API token
	err = b.Open("https://iamresponding.com/v3/agency/incidentsdashboard.aspx")
	if err != nil {
		return err
	}

	b.Dom().Find("INPUT#TokenForApi").Each(func(_ int, s *goquery.Selection) {
		c.tokenForAPI, _ = s.Attr("value")
	})
	b.Dom().Find("INPUT#memberID").Each(func(_ int, s *goquery.Selection) {
		x, _ := s.Attr("value")
		c.member, _ = strconv.ParseInt(x, 10, 64)
	})
	b.Dom().Find("INPUT#subscriberID").Each(func(_ int, s *goquery.Selection) {
		x, _ := s.Attr("value")
		c.subscriber, _ = strconv.ParseInt(x, 10, 64)
	})
	b.Dom().Find("INPUT#hdnAgency").Each(func(_ int, s *goquery.Selection) {
		x, _ := s.Attr("value")
		c.agency, _ = strconv.ParseInt(x, 10, 64)
	})
	b.Dom().Find("INPUT#ClientIarApiUrl").Each(func(_ int, s *goquery.Selection) {
		c.clientIarAPIURL, _ = s.Attr("value")
	})

	if c.memberCrypted == "" || c.adminCrypted == "" {
		return ErrIarAuthentication
	}

	if c.Debug {
		log.Printf("org = %s, member = %s, agency = %s, admin = %s, tokenForApi = %s, clientIarApiUrl = %s", c.orgCrypted, c.memberCrypted, c.agencyCrypted, c.adminCrypted, c.tokenForAPI, c.clientIarAPIURL)
	}

	pattern, err := regexp.Compile(`var Credentials={"Agency":(\d+),"Member":(\d+),"Type":(\d),"Token":"([^\"]+)",`) //"AgencyType":"([^\"]+)","SessionToken":"([A-Za-z=]+)",`)
	if err != nil {
		return err
	}
	groups := pattern.FindStringSubmatch(b.Body())
	if len(groups) < 5 {
		return errors.New("Did not find API token")
	}
	if c.Debug {
		log.Printf("%#v", groups)
	}

	c.agency, _ = strconv.ParseInt(groups[1], 10, 64)
	c.member, _ = strconv.ParseInt(groups[2], 10, 64)
	c.apiToken = groups[4]

	if groups[1] == "" || groups[2] == "" {
		return ErrIarAuthentication
	}

	c.initialized = true

	return nil
}

// GetNowRespondingWithSort()

type getNowRespondingWithSortData struct {
	NowResponding []NowResponding `xml:"nowresponding"`
}

func (c *IamRespondingAPI) GetNowRespondingWithSort() ([]NowResponding, error) {
	nr := make([]NowResponding, 0)

	if !c.initialized {
		return nr, errors.New("Not initialized")
	}

	b := c.browserObject

	v := url.Values{}
	v.Set("org", c.orgCrypted)
	v.Set("member", c.memberCrypted)
	v.Set("sort", "")
	v.Set("flag", "")
	v.Set("userType", "")
	err := b.PostForm("https://iamresponding.com/v3/AgencyServices.asmx/GetNowRespondingWithSort", v)
	if err != nil {
		return nr, err
	}

	var d getNowRespondingWithSortData
	err = xml.Unmarshal([]byte(b.Body()), &d)
	if err != nil {
		return nr, err
	}

	for _, x := range d.NowResponding {
		if strings.TrimSpace(x.MemberFName) != "" {
			nr = append(nr, x)
		}
	}
	//nr = d.NowResponding

	if c.Debug {
		log.Printf("%#v", nr)
	}

	return nr, nil
}

// GetOnScheduleWithSort()

type getOnScheduleWithSortData struct {
	XMLName    xml.Name     `xml:"newdataset"`
	OnSchedule []OnSchedule `xml:"onschedule"`
}

func (c *IamRespondingAPI) GetOnScheduleWithSort() ([]OnSchedule, error) {
	data := make([]OnSchedule, 0)

	if !c.initialized {
		return data, errors.New("Not initialized")
	}

	b := c.browserObject

	v := url.Values{}
	v.Set("org", c.orgCrypted)
	v.Set("member", c.memberCrypted)
	v.Set("sort", "")
	v.Set("flag", "0")
	v.Set("userType", "")
	err := b.PostForm("https://iamresponding.com/v3/AgencyServices.asmx/GetOnScheduleWithSort", v)
	if err != nil {
		return data, err
	}

	var d getOnScheduleWithSortData
	err = xml.Unmarshal([]byte(strings.ReplaceAll(b.Body(), "\n    ", "")), &d)
	if err != nil {
		return data, err
	}

	for _, x := range d.OnSchedule {
		if strings.TrimSpace(x.MemberName) != "" {
			data = append(data, x)
		}
	}

	//data = d.OnSchedule

	if c.Debug {
		log.Printf("%s / %#v", b.Body(), data)
	}

	return data, nil
}

// ListWithParser()

type listWithParserData struct {
	XMLName          xml.Name          `xml:"newdataset"`
	DispatchMessages []DispatchMessage `xml:"dispatchmessages"`
}

func (c *IamRespondingAPI) ListWithParser() ([]DispatchMessage, error) {
	data := make([]DispatchMessage, 0)

	if !c.initialized {
		return data, errors.New("Not initialized")
	}

	b := c.browserObject

	v := url.Values{}
	v.Set("subscriber", c.orgCrypted)
	v.Set("member", c.memberCrypted)
	v.Set("admin", c.adminCrypted)
	v.Set("agency", c.agencyCrypted)
	err := b.PostForm("https://iamresponding.com/v3/DispatchMessages.asmx/ListWithParser", v)
	if err != nil {
		return data, err
	}

	var d listWithParserData
	err = xml.Unmarshal([]byte(strings.ReplaceAll(b.Body(), "\n    ", "")), &d)
	if err != nil {
		return data, err
	}

	for i := range d.DispatchMessages {
		d.DispatchMessages[i].MessageBody = strings.Replace(d.DispatchMessages[i].MessageBody, "\n<br />", "", -1)
		parts := strings.Split(d.DispatchMessages[i].MessageBody, " * ")
		if len(parts) > 1 {
			d.DispatchMessages[i].Address = parts[1]
		}
	}

	data = d.DispatchMessages

	if c.Debug {
		log.Printf("%s / %#v", b.Body(), data)
	}

	return data, nil
}

type getIncidentInfoResponse struct {
	Data []IncidentInfoData `json:"d"`
}

// GetIncidentInfo retrieves atomic incident information based on the unique
// ID of the incient in question
func (c *IamRespondingAPI) GetIncidentInfo(incident int64) (IncidentInfoData, error) {
	if !c.initialized {
		return IncidentInfoData{}, errors.New("Not initialized")
	}

	b := c.browserObject

	params := map[string]interface{}{
		"messageID": incident,
		"token":     c.tokenForAPI,
	}
	post, err := json.Marshal(params)
	if err != nil {
		return IncidentInfoData{}, err
	}

	err = b.Post("https://iamresponding.com/v3/agency/IncidentsDashboard.aspx/GetIncidentInfo", "application/json", strings.NewReader(string(post)))
	if err != nil {
		return IncidentInfoData{}, err
	}

	if c.Debug {
		log.Printf("GetIncidentInfo: Body: [[ %s ]]", strings.ReplaceAll(b.Body(), "&#34;", `"`))
	}

	var d getIncidentInfoResponse
	err = json.Unmarshal([]byte(strings.ReplaceAll(b.Body(), "&#34;", `"`)), &d)
	if err != nil {
		return IncidentInfoData{}, err
	}

	if len(d.Data) == 0 {
		return IncidentInfoData{}, errors.New("No data for incident")
	}
	return d.Data[0], nil
}

// GetLatestIncidents fetches the latest list of IAmResponding emergency
// incidents
func (c *IamRespondingAPI) GetLatestIncidents() ([]IncidentInfoData, error) {
	if !c.initialized {
		return []IncidentInfoData{}, errors.New("Not initialized")
	}

	b := c.browserObject

	params := map[string]interface{}{
		"memberID": c.member,
		"token":    c.tokenForAPI,
	}
	post, err := json.Marshal(params)
	if err != nil {
		return []IncidentInfoData{}, err
	}

	err = b.Post("https://iamresponding.com/v3/agency/IncidentsDashboard.aspx/GetLatestIncidents?buster=%27+new%20Date().getTime();", "application/json", strings.NewReader(string(post)))
	if err != nil {
		return []IncidentInfoData{}, err
	}

	if c.Debug {
		log.Printf("GetLatestIncidents: post: %s, Body: [[ %s ]]", post, strings.ReplaceAll(b.Body(), "&#34;", `"`))
	}

	var d getIncidentInfoResponse
	err = json.Unmarshal([]byte(strings.ReplaceAll(b.Body(), "&#34;", `"`)), &d)
	if err != nil {
		return []IncidentInfoData{}, err
	}

	return d.Data, nil
}

type getRemindersByMemberResponse struct {
	XMLName xml.Name `xml:"newdataset"`
	Data    []Event  `xml:"event"`
}

func (c *IamRespondingAPI) GetRemindersByMember() ([]Event, error) {
	if !c.initialized {
		return []Event{}, errors.New("Not initialized")
	}

	b := c.browserObject

	log.Printf("GetRemindersByMember()")
	v := url.Values{}
	v.Set("subsString", fmt.Sprintf("%d", c.agency))
	v.Set("days", "7")
	err := b.PostForm("https://iamresponding.com/v3/AgencyServices.asmx/GetRemindersByMember", v)
	if err != nil {
		return []Event{}, err
	}

	if c.Debug {
		log.Printf("GetRemindersByMember: subsString: %d: Body: [[ %s ]]", c.agency, b.Body())
	}

	var d getRemindersByMemberResponse
	err = xml.Unmarshal([]byte(b.Body()), &d)
	if err != nil {
		return []Event{}, err
	}

	return d.Data, nil
}
