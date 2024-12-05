package iarapi

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
	"github.com/jbuchbinder/shims"
	"gopkg.in/yaml.v2"
)

var (
	// ErrIarAuthentication is an error thrown when IAR does not authenticate
	// a query
	ErrIarAuthentication = errors.New("unable to authenticate to IAR due to incorrect credentials")
)

// IamRespondingAPI represents an API access object
type IamRespondingAPI struct {
	// Debug sets whether debugging output is enabled
	Debug bool

	browserObject *browser.Browser
	initialized   bool
	member        int64
	agency        int64
	tokenForAPI   string
	sessionToken  string
	agencies      []int
}

func (c *IamRespondingAPI) Login(agency, user, pass string) error {
	b := surf.NewBrowser()
	c.browserObject = b

	jar, _ := cookiejar.New(&cookiejar.Options{})

	b.SetUserAgent(agent.Chrome())
	b.SetCookieJar(jar)

	// Required to not have ASP.NET garbage yak all over me
	b.AddRequestHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	b.AddRequestHeader("Cookie", "CookieConsent=yes")
	// Get token
	if c.Debug {
		log.Printf("Attempt to get request verification token")
	}
	err := b.Open("https://auth.iamresponding.com/login/member")
	if err != nil {
		return nil
	}
	rvt := ""
	b.Dom().Find("INPUT[name='__RequestVerificationToken']").Each(func(_ int, s *goquery.Selection) {
		rvt, _ = s.Attr("value")
	})
	if rvt == "" {
		return fmt.Errorf("unable to obtain request verification token")
	}
	if c.Debug {
		log.Printf("request verification token = %s", rvt)
	}

	l := url.Values{}
	l.Add("Input.Agency", agency)
	l.Add("Input.Username", user)
	l.Add("Input.Password", pass)
	l.Add("Input.button", "login")
	l.Add("Input.ReturnUrl", "")
	l.Add("__RequestVerificationToken", rvt)
	if c.Debug {
		log.Printf("Attempt to login")
	}

	err = b.PostForm("https://auth.iamresponding.com/login/member", l)
	if err != nil {
		return err
	}

	if c.Debug {
		log.Printf("DEBUG: /login/member: state = %d, body = %d", b.StatusCode(), len(b.Body()))
		//log.Printf("DEBUG: /login/member: state = %d, body = %s", b.StatusCode(), b.Body())
	}

	// Catch authentication issues
	if strings.Contains(b.Body(), "The log-in information that you have entered is incorrect.") {
		if c.Debug {
			log.Printf("Login(): %d, %q", strings.Index(b.Body(), "The log-in information that you have entered is incorrect."), b.Body())
		}
		return ErrIarAuthentication
	}

	err = b.Open("https://app.iamresponding.com/v3/Pages/DispatcherScreen.aspx")
	if c.Debug {
		log.Printf("DEBUG: /v3/Pages/DispatcherScreen.aspx: state = %d, body = %d", b.StatusCode(), len(b.Body()))
	}
	if err != nil {
		log.Printf("ERR: %s", err.Error())
	}
	if !strings.Contains(b.Body(), "var Credentials={") {
		return fmt.Errorf("no credentials found")
	}

	// ;//]]>

	{
		if c.Debug {
			log.Printf("creds regexp")
		}
		credsRegex, err := regexp.Compile(`var Credentials=(.+);var ValidAgencies=`) //   ;//]]>`)
		if err != nil {
			return err
		}
		creds := credsRegex.FindStringSubmatch(b.Body())[1]

		var o Credentials
		err = json.Unmarshal([]byte(creds), &o)
		if err != nil {
			return err
		}
		c.agency = o.Agency
		c.member = o.Member
		c.tokenForAPI = o.Token
		c.sessionToken = o.SessionToken

		if c.Debug {
			log.Printf("INFO: Credentials = %#v", o)
		}
	}

	{
		if c.Debug {
			log.Printf("agencies regexp")
		}
		agenciesRegex, err := regexp.Compile(`var ValidAgencies=(.+);var AgencyNames=`) //   ;//]]>`)
		if err != nil {
			return err
		}
		agencies := agenciesRegex.FindStringSubmatch(b.Body())[1]
		agencies = strings.ReplaceAll(agencies, `'`, `"`)

		var o AgenciesList
		err = json.Unmarshal([]byte(agencies), &o)
		if err != nil {
			return err
		}

		if c.Debug {
			log.Printf("INFO: Agencies = %#v", o)
		}

		c.agencies = shims.ArrayConvert(o, func(x string) int {
			return shims.SingleValueDiscardError(strconv.Atoi(x))
		})
	}

	{
		if c.Debug {
			log.Printf("DEBUG: agency names regexp")
		}
		agencyNamesRegex, err := regexp.Compile(`;var AgencyNames=(.+);//]]>`) //   ;//]]>`)
		if err != nil {
			return err
		}
		agencyNames := agencyNamesRegex.FindStringSubmatch(b.Body())[1]

		if c.Debug {
			log.Printf("names = %s", agencyNames)
		}

		var o AgencyNames
		err = yaml.Unmarshal([]byte(agencyNames), &o)
		if err != nil {
			return err
		}

		if c.Debug {
			log.Printf("INFO: Agency Names = %#v", o)
		}
	}

	if c.Debug {
		log.Printf("INFO: apiToken = %s, sessionToken = %s, agency = %d, member = %d", c.tokenForAPI, c.sessionToken, c.agency, c.member)
	}

	c.initialized = true

	return nil
}

// GetAgenciesInformation returns dispatch view information
func (c *IamRespondingAPI) GetAgenciesInformation(agencies []int) ([]DispatcherAgencyInformation, error) {
	nr := make([]DispatcherAgencyInformation, 0)

	if !c.initialized {
		return nr, errors.New("not initialized")
	}

	b := c.browserObject

	type getAgenciesInformationRequest struct {
		Credentials struct {
			Agency       int64  `json:"agency"`
			Member       int64  `json:"member"`
			Type         int    `json:"type"`
			Token        string `json:"token"`
			AgencyType   string `json:"AgencyType"`
			SessionToken string `json:"SessionToken"`
		} `json:"credentials"`
		Agencies    string `json:"agencies"`
		OnDutySort1 int    `json:"onDutySort1"`
		OnDutySort2 int    `json:"onDutySort2"`
		RespSort1   int    `json:"respSort1"`
		RespSort2   int    `json:"respSort2"`
	}

	v := getAgenciesInformationRequest{}
	v.Credentials.Agency = c.agency
	v.Credentials.Member = c.member
	v.Credentials.Type = 4
	v.Credentials.Token = c.tokenForAPI
	v.Credentials.AgencyType = "Dispatcher"
	v.Credentials.SessionToken = c.sessionToken
	v.Agencies = strings.Join(
		shims.ArrayConvert(agencies, func(x int) string {
			return fmt.Sprintf("%d", x)
		}), ",",
	)
	v.OnDutySort1 = 1
	v.OnDutySort2 = 1
	v.RespSort1 = 1
	v.RespSort2 = 1

	//log.Printf("%#v", v)

	buf := bytes.NewReader(shims.SingleValueDiscardError(json.Marshal(v)))
	err := b.Post("https://app.iamresponding.com/v3/AgencyServices.asmx/GetAgenciesInformation", "application/json", buf)
	if err != nil {
		return nr, err
	}

	type dwrapper struct {
		D []DispatcherAgencyInformation `json:"d"`
	}

	body := b.Body()
	body = strings.ReplaceAll(body, "&#34;", `"`)
	//log.Printf("%s", body)
	var o dwrapper
	err = json.Unmarshal([]byte(body), &o)
	if err == nil {
		nr = o.D
	}
	return nr, err
}

/*
// GetNowRespondingWithSort()

type getNowRespondingWithSortData struct {
	NowResponding []NowResponding `xml:"nowresponding"`
}

func (c *IamRespondingAPI) GetNowRespondingWithSort() ([]NowResponding, error) {
	nr := make([]NowResponding, 0)

	if !c.initialized {
		return nr, errors.New("not initialized")
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
*/

// GetOnScheduleWithSort()

type getOnScheduleWithSortData struct {
	XMLName    xml.Name     `xml:"newdataset"`
	OnSchedule []OnSchedule `xml:"onschedule"`
}

func (c *IamRespondingAPI) GetOnScheduleWithSort() ([]OnSchedule, error) {
	data := make([]OnSchedule, 0)

	if !c.initialized {
		return data, errors.New("not initialized")
	}

	b := c.browserObject

	v := url.Values{}
	//v.Set("org", c.orgCrypted)
	//v.Set("member", c.memberCrypted)
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
		return data, errors.New("not initialized")
	}

	b := c.browserObject

	v := url.Values{}
	//v.Set("subscriber", c.orgCrypted)
	//v.Set("member", c.memberCrypted)
	//v.Set("admin", c.adminCrypted)
	//v.Set("agency", c.agencyCrypted)
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
		return IncidentInfoData{}, errors.New("not initialized")
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
		return IncidentInfoData{}, errors.New("no data for incident")
	}
	return d.Data[0], nil
}

// GetLatestIncidents fetches the latest list of IAmResponding emergency
// incidents
func (c *IamRespondingAPI) GetLatestIncidents() ([]IncidentInfoData, error) {
	if !c.initialized {
		return []IncidentInfoData{}, errors.New("not initialized")
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
		return []Event{}, errors.New("not initialized")
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
