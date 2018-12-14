package iarapi

import (
	"testing"
)

func Test_API_Failure_Catch(t *testing.T) {
	m := IamRespondingAPI{}
	err := m.Login(TEST_AGENCY, "someuser", "somepassword")
	t.Error(err)
	if err == nil {
		t.Fail()
	}
}

func Test_API(t *testing.T) {
	m := IamRespondingAPI{}
	err := m.Login(TEST_AGENCY, TEST_USER, TEST_PASS)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	{
		nr, err := m.GetNowRespondingWithSort()
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("NowResponding: %#v", nr)
	}

	{
		sched, err := m.GetOnScheduleWithSort()
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("OnSchedule: %#v", sched)
	}

	var Dm []DispatchMessage
	{
		Dm, err = m.ListWithParser()
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("ListWithParser: %#v", Dm)
	}

	currentCase := Dm[len(Dm)-1]
	t.Logf("CURRENT ADDRESS : %s", currentCase.Address)

	var E []Event
	{
		E, err = m.GetRemindersByMember()
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("GetRemindersByMember: %#v", E)
	}

	{
		i, err := m.GetLatestIncidents()
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("GetLatestIncidents: %#v", i)
	}

	t.Logf("%#v", m)

	/*
		{
			data, err := m.GetIncidentInfo(Dm[len(Dm)-1].ID, m.apiToken)
			if err != nil {
				t.Error(err)
				t.Fail()
			}
			t.Logf("GetIncidentInfo: %#v", data)
		}
	*/
}
