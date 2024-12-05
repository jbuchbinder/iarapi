package iarapi

import (
	"encoding/xml"
	"time"
)

// CustomTime is an object meant to allow for marshalling and unmarshalling of
// IAmResponding formatted dates
type CustomTime struct {
	time.Time
}

func (c *CustomTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	const shortForm = "2006-01-02T15:04:05.999999999" // Mon Jan 2 15:04:05 -0700 MST 2006
	var v string
	d.DecodeElement(&v, &start)
	parse, err := time.Parse(shortForm, v)
	if err != nil {
		return err
	}
	*c = CustomTime{parse}
	return nil
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

// Event is an IAmResponding calendar event
type Event struct {
	XMLName            xml.Name   `xml:"event"`
	EventID            int64      `xml:"eventid"`
	Subject            string     `xml:"subject"`
	EventStart         CustomTime `xml:"eventstart"`
	EventEnd           CustomTime `xml:"eventend"`
	EventsRecurrenceID int64      `xml:"eventsrecurrenceid"`
	SubscriberID       int64      `xml:"subscriberid"`
	//<RecurrenceRule></RecurrenceRule>
}

// IncidentInfoData represents a single IAmResponding emergency
// incident
type IncidentInfoData struct {
	IncidentInfoData       string `json:"__type"`
	ID                     int64  `json:"Id"`
	ArrivedOn              string `json:"ArrivedOn"`
	SubscriberID           int64  `json:"SubscriberId"`
	Body                   string `json:"Body,omitempty"`
	Subject                string `json:"Subject,omitempty"`
	MessageSubject         string `json:"MessageSubject,omitempty"`
	Address                string `json:"Address"`
	OverrideBounds         bool   `json:"OverrideBounds"`
	VerifiedAddressStatus  int    `json:"VerifiedAddressStatus"`
	VerifiedAddressID      int    `json:"VerifiedAddressId"`
	VerifiedAddressAddedBy string `json:"VerifiedAddressAddedBy"`
	UpdatedOn              string `json:"UpdatedOn"`
	UpdatedOnToShow        string `json:"UpdatedOnToShow"`
	LongDirection          string `json:"longDirection"`
}

/*
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
*/

type OnSchedule struct {
	MemberName      string `xml:"membername"`
	MemberCat       string `xml:"membercat"`
	InStationOrHome string `xml:"instationorhome"`
	MemberStation   string `xml:"memberstation"`
	UntilAt         string `xml:"untilat"`
}

type Credentials struct {
	Agency       int64  `json:"Agency"`
	Member       int64  `json:"Member"`
	Type         int    `json:"Type"`
	Token        string `json:"Token"`
	AgencyType   string `json:"AgencyType"`
	SessionToken string `json:"SessionToken"`
}

type AgenciesList []string

type AgencyNames map[string]AgencyInfo

type AgencyInfo struct {
	ID          int64  `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Index       int    `json:"index" yaml:"index"`
	AutoDisplay string `json:"autoDisplay" yaml:"autoDisplay"`
}

type DispatcherAgencyInformation struct {
	AgencyId                int64             `json:"AgencyId"`
	OnDuties                []PersonnelStatus `json:"OnDuties"`
	NowRespondings          []NowResponding   `json:"NowRespondings"`
	OutOfServiceApparatuses []map[string]any  `json:"OutOfServiceApparatuses"`
	InServiceApparatuses    []map[string]any  `json:"InServiceApparatuses"`
}

type PersonnelStatus struct {
	Name          string `json:"Name"`
	ID            int64  `json:"Id"`
	Position      string `json:"Position"`
	OnDutyFor     string `json:"OnDutyFor"`
	StationForm   string `json:"StationForm"`
	UntilAt       string `json:"UntilAt"`
	UntilAtString string `json:"UntilAtString"`
	SubscriberId  int64  `json:"SubscriberId"`
}

type NowResponding struct {
	CallingTime       string `json:"CallingTime"`
	ID                int64  `json:"Id"`
	Name              string `json:"Name"`
	Position          string `json:"Position"`
	RespondingTo      string `json:"RespondingTo"`
	ETA               string `json:"ETA"`
	CallerNumber      string `json:"CallerNumber"`
	CalledNumber      string `json:"CalledNumber"`
	UserInput         string `json:"UserInput"`
	CallingTimeString string `json:"CallingTimeString"`
	ETAString         string `json:"ETAString"`
}
