package iamresponding

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
)

type IamRespondingAPI struct {
	browserObject *browser.Browser
	initialized   bool
	orgCrypted    string
	member        int64
	memberCrypted string
	adminCrypted  string
	agency        int64
	agencyCrypted string
	apiToken      string
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

	log.Printf("org = %s, member = %s, agency = %s, admin = %s", c.orgCrypted, c.memberCrypted, c.agencyCrypted, c.adminCrypted)

	// Get API token
	err = b.Open("https://iamresponding.com/v3/agency/incidentsdashboard.aspx")
	if err != nil {
		return err
	}

	pattern, err := regexp.Compile(`var Credentials={"Agency":(\d+),"Member":(\d+),"Type":(\d),"Token":"([^\"]+)",`) //"AgencyType":"([^\"]+)","SessionToken":"([A-Za-z=]+)",`)
	if err != nil {
		return err
	}
	groups := pattern.FindStringSubmatch(b.Body())
	if len(groups) < 5 {
		return errors.New("Did not find API token")
	}
	log.Printf("%#v", groups)
	c.agency, _ = strconv.ParseInt(groups[1], 10, 64)
	c.member, _ = strconv.ParseInt(groups[2], 10, 64)
	c.apiToken = groups[4]

	c.initialized = true

	return nil
}

// GetNowRespondingWithSort()

type GetNowRespondingWithSortData struct {
	NowResponding []NowResponding `xml:"nowresponding"`
}

type NowResponding struct {
	MemberFName   string `xml:"memberfname"`
	MemberCat     string `xml:"membercat"`
	MemberStation string `xml:"memberstation"`
	CallingTime   string `xml:"callingtime"`
	ETA           string `xml:"eta"`
	CallerNo      string `xml:"callerno"`
	CalledNo      string `xml:"calledno"`
	UserInput     string `xml:"userinput"`
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

	var d GetNowRespondingWithSortData
	err = xml.Unmarshal([]byte(b.Body()), &d)
	if err != nil {
		return nr, err
	}

	nr = d.NowResponding

	log.Printf("%#v", nr)

	return nr, nil
}

// GetOnScheduleWithSort()

type GetOnScheduleWithSortData struct {
	XMLName    xml.Name     `xml:"newdataset"`
	OnSchedule []OnSchedule `xml:"onschedule"`
}

type OnSchedule struct {
	MemberName      string `xml:"membername"`
	MemberCat       string `xml:"membercat"`
	InStationOrHome string `xml:"instationorhome"`
	MemberStation   string `xml:"memberstation"`
	UntilAt         string `xml:"untilat"`
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

	var d GetOnScheduleWithSortData
	err = xml.Unmarshal([]byte(b.Body()), &d)
	if err != nil {
		return data, err
	}

	data = d.OnSchedule

	//log.Printf("%s / %#v", b.Body(), data)

	return data, nil
}

// ListWithParser()

type ListWithParserData struct {
	XMLName          xml.Name          `xml:"newdataset"`
	DispatchMessages []DispatchMessage `xml:"dispatchmessages"`
}

type DispatchMessage struct {
	ID                      int64  `xml:"id"`
	MessageBody             string `xml:"messagebody"`
	DestinationEmailAddress string `xml:"destinationemailaddress"`
	MessageSubject          string `xml:"messagesubject"`
	VerifiedStatus          string `xml:"verifiedstatus"`
	ArrivedOnString         string `xml:"arrivedonstring"`
	Address                 string `xml:"-"`
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

	var d ListWithParserData
	err = xml.Unmarshal([]byte(b.Body()), &d)
	if err != nil {
		return data, err
	}

	for i, _ := range d.DispatchMessages {
		d.DispatchMessages[i].MessageBody = strings.Replace(d.DispatchMessages[i].MessageBody, "\n<br />", "", -1)
		parts := strings.Split(d.DispatchMessages[i].MessageBody, " * ")
		d.DispatchMessages[i].Address = parts[1]
	}

	data = d.DispatchMessages

	log.Printf("%s / %#v", b.Body(), data)

	return data, nil
}

// GetIncidentInfo()

type GetIncidentInfoResponse struct {
	Data []IncidentInfoData `json:"d"`
}

type IncidentInfoData struct {
	IncidentInfoData       string    `json:"__type"`
	Id                     int64     `json:"Id"`
	ArrivedOn              time.Time `json:"ArrivedOn"`
	SubscriberId           int64     `json:"SubscriberId"`
	Subject                string    `json:"Subject"`
	Address                string    `json:"Address"`
	OverrideBounds         bool      `json:"OverrideBounds"`
	VerifiedAddressStatus  int       `json:"VerifiedAddressStatus"`
	VerifiedAddressId      int       `json:"VerifiedAddressId"`
	VerifiedAddressAddedBy string    `json:"VerifiedAddressAddedBy"`
	UpdatedOn              string    `json:"UpdatedOn"`
	UpdatedOnToShow        string    `json:"UpdatedOnToShow"`
	LongDirection          string    `json:"longDirection"`
}

func (c *IamRespondingAPI) GetIncidentInfo(incident int64, token string) (IncidentInfoData, error) {
	if !c.initialized {
		return IncidentInfoData{}, errors.New("Not initialized")
	}

	b := c.browserObject

	params := map[string]interface{}{
		"messageId": incident,
		"token":     token,
	}
	post, err := json.Marshal(params)
	if err != nil {
		return IncidentInfoData{}, err
	}

	err = b.Post("https://iamresponding.com/v3/IncidentsDashboard.asmx/GetIncidentInfo", "application/json", strings.NewReader(string(post)))
	if err != nil {
		return IncidentInfoData{}, err
	}

	log.Printf("GetIncidentInfo: Body: [[ %s ]]", b.Body())

	var d GetIncidentInfoResponse
	err = json.Unmarshal([]byte(b.Body()), &d)
	if err != nil {
		return IncidentInfoData{}, err
	}

	return d.Data[0], nil
}
