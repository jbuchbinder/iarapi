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

type OnSchedule struct {
	MemberName      string `xml:"membername"`
	MemberCat       string `xml:"membercat"`
	InStationOrHome string `xml:"instationorhome"`
	MemberStation   string `xml:"memberstation"`
	UntilAt         string `xml:"untilat"`
}
