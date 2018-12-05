package iarapi

import (
	"testing"
)

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

	{
		data, err := m.GetIncidentInfo(Dm[0].ID, m.apiToken)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("GetIncidentInfo: %#v", data)
	}
}
