package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"gotest.tools/assert"
	"wecode.sorint.it/opensource/papagaio-api/config"
	"wecode.sorint.it/opensource/papagaio-api/dto"
	"wecode.sorint.it/opensource/papagaio-api/test"
	"wecode.sorint.it/opensource/papagaio-api/test/mock/mock_repository"
	triggerDto "wecode.sorint.it/opensource/papagaio-api/trigger/dto"
	"wecode.sorint.it/opensource/papagaio-api/utils"
)

var serviceTrigger TriggersService

func setupTriggerMock(t *testing.T) {
	config.Config.TriggersConfig.StartOrganizationsTrigger = true
	config.Config.TriggersConfig.StartRunFailedTrigger = true
	config.Config.TriggersConfig.StartUsersTrigger = true

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	db = mock_repository.NewMockDatabase(ctl)
	tr := utils.ConfigUtils{Db: db}

	serviceTrigger = TriggersService{
		Db: db,
		Tr: tr,
		RtDtoOrganizationSynk: &triggerDto.TriggerRunTimeDto{
			Chan: make(chan triggerDto.TriggerMessage, 1),
		},
		RtDtoDiscoveryRunFails: &triggerDto.TriggerRunTimeDto{
			Chan: make(chan triggerDto.TriggerMessage, 1),
		},
		RtDtoUserSynk: &triggerDto.TriggerRunTimeDto{
			Chan: make(chan triggerDto.TriggerMessage, 1),
		},
	}
}

func TestGetTriggetsConfigOK(t *testing.T) {
	setupTriggerMock(t)

	db.EXPECT().GetOrganizationsTriggerTime().Return(1)
	db.EXPECT().GetRunFailedTriggerTime().Return(2)
	db.EXPECT().GetUsersTriggerTime().Return(3)

	router := test.SetupBaseRouter(nil)
	router.HandleFunc("/gettriggersconfig", serviceTrigger.GetTriggersConfig)
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	resp, err := client.Get(ts.URL + "/gettriggersconfig")

	var responseDto dto.ConfigTriggersDto
	test.ParseBody(resp, &responseDto)

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")
	assert.Equal(t, responseDto.OrganizationsTriggerTime, uint(1))
	assert.Equal(t, responseDto.RunFailedTriggerTime, uint(2))
	assert.Equal(t, responseDto.UsersTriggerTime, uint(3))
}

func TestSaveTriggetsConfigOK(t *testing.T) {
	setupTriggerMock(t)

	reqDto := dto.ConfigTriggersDto{}
	reqDto.OrganizationsTriggerTime = 4
	reqDto.RunFailedTriggerTime = 5
	reqDto.UsersTriggerTime = 6

	db.EXPECT().SaveOrganizationsTriggerTime(int(reqDto.OrganizationsTriggerTime))
	db.EXPECT().SaveRunFailedTriggerTime(int(reqDto.RunFailedTriggerTime))
	db.EXPECT().SaveUsersTriggerTime(int(reqDto.UsersTriggerTime))

	router := test.SetupBaseRouter(nil)
	router.HandleFunc("/savetriggersconfig", serviceTrigger.SaveTriggersConfig)
	ts := httptest.NewServer(router)
	defer ts.Close()

	data, _ := json.Marshal(reqDto)
	requestBody := strings.NewReader(string(data))

	client := ts.Client()
	resp, err := client.Post(ts.URL+"/savetriggersconfig", "application/json", requestBody)

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")
}

func TestRestartTriggersOK(t *testing.T) {
	setupTriggerMock(t)

	router := test.SetupBaseRouter(nil)
	router.HandleFunc("/restarttriggers", serviceTrigger.RestartTriggers)
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	resp, err := client.Get(ts.URL + "/restarttriggers?restartorganizationsynktrigger&restartRunsFailedDiscoveryTrigger&restartUsersSynkTrigger")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")
}

func TestRestartTriggersAllOK(t *testing.T) {
	setupTriggerMock(t)

	router := test.SetupBaseRouter(nil)
	router.HandleFunc("/restarttriggers", serviceTrigger.RestartTriggers)
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	resp, err := client.Get(ts.URL + "/restarttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")
}

func TestRestartTriggersWhenParameterIsInvalid(t *testing.T) {
	setupTriggerMock(t)

	router := test.SetupBaseRouter(nil)
	router.HandleFunc("/restarttriggers", serviceTrigger.RestartTriggers)
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	resp, err := client.Get(ts.URL + "/restarttriggers?all=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")

	resp, err = client.Get(ts.URL + "/restarttriggers?organizationsynktrigger=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")

	resp, err = client.Get(ts.URL + "/restarttriggers?runsFailedDiscoveryTrigger=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")

	resp, err = client.Get(ts.URL + "/restarttriggers?usersSynkTrigger=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")
}

func TestRestartTriggerWhenTriggersAreNil(t *testing.T) {

	//1

	serviceTrigger = TriggersService{}
	router := test.SetupBaseRouter(nil)
	router.HandleFunc("/restarttriggers", serviceTrigger.RestartTriggers)
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	resp, err := client.Get(ts.URL + "/restarttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

	//2

	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

	//3
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

	//4
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoUserSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?organizationsynktrigger=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")

	//5

	resp, err = client.Get(ts.URL + "/restarttriggers?organizationsynktrigger=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")

	//6

	serviceTrigger = TriggersService{}
	resp, err = client.Get(ts.URL + "/restarttriggers?organizationsynktrigger")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusInternalServerError, "http StatusCode is not OK")

	// 7
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoUserSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?runsFailedDiscoveryTrigger=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")

	// 8
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoUserSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}

	serviceTrigger.RtDtoDiscoveryRunFails = nil

	resp, err = client.Get(ts.URL + "/restarttriggers?runsFailedDiscoveryTrigger")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusInternalServerError, "http StatusCode is not OK")

	// 9
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoUserSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?usersSynkTrigger=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")

	// 10

	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoUserSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}

	serviceTrigger.RtDtoUserSynk = nil

	resp, err = client.Get(ts.URL + "/restarttriggers?usersSynkTrigger")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusInternalServerError, "http StatusCode is not OK")
}

func TestRestartTriggersWithAllWhenIsRunning(t *testing.T) {
	// When query parameter is ?all

	serviceTrigger = TriggersService{}
	router := test.SetupBaseRouter(nil)
	router.HandleFunc("/restarttriggers", serviceTrigger.RestartTriggers)
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()

	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		IsRunning: true,
	}

	resp, err := client.Get(ts.URL + "/restarttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "OrganizationsTrigger can't restart at the moment")

	//
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		IsRunning: true,
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "RunsFailedDiscoveryTrigger can't restart at the moment")

	//
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoUserSynk = &triggerDto.TriggerRunTimeDto{
		IsRunning: true,
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "UsersSynkTrigger can't restart at the moment")

	//
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}
	serviceTrigger.RtDtoUserSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http status is not ok")

}
func TestGetTriggerStatus(t *testing.T) {
	setupTriggerMock(t)

	today := time.Now()

	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan:        make(chan triggerDto.TriggerMessage, 1),
		LastRun:     today,
		IsRunning:   false,
		TriggerTime: 1,
	}

	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan:        make(chan triggerDto.TriggerMessage, 1),
		LastRun:     today,
		IsRunning:   false,
		TriggerTime: 1,
	}

	serviceTrigger.RtDtoUserSynk = &triggerDto.TriggerRunTimeDto{
		Chan:        make(chan triggerDto.TriggerMessage, 1),
		LastRun:     today,
		IsRunning:   false,
		TriggerTime: 1,
	}

	router := test.SetupBaseRouter(nil)
	router.HandleFunc("/triggersstatus", serviceTrigger.GetTriggersStatus)
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()

	resp, err := client.Get(ts.URL + "/triggersstatus")

	responseDto := dto.TriggersStatusDto{
		OrganizationStatus:      dto.TriggerDto{},
		DiscoveryRunFailsStatus: dto.TriggerDto{},
		UserSynkStatus:          dto.TriggerDto{},
	}
	test.ParseBody(resp, &responseDto)

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

	//RtDtoOrganizationSynk

	assert.Equal(t, responseDto.OrganizationStatus.IsStarted, true)
	assert.Equal(t, *responseDto.OrganizationStatus.IsRunning, false)
	data1 := *responseDto.OrganizationStatus.LastRun
	str := data1.String()
	runes := []rune(str)
	data1c := string(runes[0:19])

	str = today.String()
	runes = []rune(str)
	data2c := string(runes[0:19])
	assert.Equal(t, data1c, data2c)

	timeLeft := time.Duration(int64(*responseDto.OrganizationStatus.TimeLeft))
	tt := timeLeft.Round(time.Duration(serviceTrigger.RtDtoOrganizationSynk.TriggerTime) * time.Minute)

	assert.Equal(t, tt, time.Minute)

	//RtDtoDiscoveryRunFails

	assert.Equal(t, responseDto.DiscoveryRunFailsStatus.IsStarted, true)
	assert.Equal(t, *responseDto.DiscoveryRunFailsStatus.IsRunning, false)
	data1 = *responseDto.DiscoveryRunFailsStatus.LastRun
	str = data1.String()
	runes = []rune(str)
	data1c = string(runes[0:19])

	str = today.String()
	runes = []rune(str)
	data2c = string(runes[0:19])
	assert.Equal(t, data1c, data2c)

	timeLeft = time.Duration(int64(*responseDto.DiscoveryRunFailsStatus.TimeLeft))
	tt = timeLeft.Round(time.Duration(serviceTrigger.RtDtoDiscoveryRunFails.TriggerTime) * time.Minute)

	assert.Equal(t, tt, time.Minute)

	//RtDtoDiscoveryRunFails

	assert.Equal(t, responseDto.DiscoveryRunFailsStatus.IsStarted, true)
	assert.Equal(t, *responseDto.DiscoveryRunFailsStatus.IsRunning, false)
	data1 = *responseDto.DiscoveryRunFailsStatus.LastRun
	str = data1.String()
	runes = []rune(str)
	data1c = string(runes[0:19])

	str = today.String()
	runes = []rune(str)
	data2c = string(runes[0:19])
	assert.Equal(t, data1c, data2c)

	timeLeft = time.Duration(int64(*responseDto.DiscoveryRunFailsStatus.TimeLeft))
	tt = timeLeft.Round(time.Duration(serviceTrigger.RtDtoDiscoveryRunFails.TriggerTime) * time.Minute)

	assert.Equal(t, tt, time.Minute)

	//UserSynkStatus

	assert.Equal(t, responseDto.UserSynkStatus.IsStarted, true)
	assert.Equal(t, *responseDto.UserSynkStatus.IsRunning, false)
	data1 = *responseDto.UserSynkStatus.LastRun
	str = data1.String()
	runes = []rune(str)
	data1c = string(runes[0:19])

	str = today.String()
	runes = []rune(str)
	data2c = string(runes[0:19])
	assert.Equal(t, data1c, data2c)

	timeLeft = time.Duration(int64(*responseDto.UserSynkStatus.TimeLeft))
	tt = timeLeft.Round(time.Duration(serviceTrigger.RtDtoUserSynk.TriggerTime) * time.Minute)

	assert.Equal(t, tt, time.Minute)
}

func TestStopTriggersWhenAllAreOk(t *testing.T) {
	serviceTrigger = TriggersService{}
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}
	serviceTrigger.RtDtoUserSynk = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}

	router := test.SetupBaseRouter(nil)
	router.HandleFunc("/stoptriggers", serviceTrigger.StopTriggers)
	ts := httptest.NewServer(router)
	defer ts.Close()

	data, _ := json.Marshal(nil)
	requestBody := strings.NewReader(string(data))

	client := ts.Client()
	resp, err := client.Post(ts.URL+"/stoptriggers?all", "application/json", requestBody)

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

}
func TestStopTriggersWhenEachTriggerIsOk(t *testing.T) {
	serviceTrigger = TriggersService{}
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}
	serviceTrigger.RtDtoUserSynk = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}

	router := test.SetupBaseRouter(nil)
	router.HandleFunc("/stoptriggers", serviceTrigger.StopTriggers)
	ts := httptest.NewServer(router)
	defer ts.Close()

	data, _ := json.Marshal(nil)
	requestBody := strings.NewReader(string(data))

	client := ts.Client()
	resp, err := client.Post(ts.URL+"/stoptriggers?organizationsynktrigger", "application/json", requestBody)

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

	//

	resp, err = client.Post(ts.URL+"/stoptriggers?runsFailedDiscoveryTrigger", "application/json", requestBody)

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

	//

	resp, err = client.Post(ts.URL+"/stoptriggers?usersSynkTrigger", "application/json", requestBody)

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")
}

func TestStopTriggersWhenAllAreInvalid(t *testing.T) {
	router := test.SetupBaseRouter(nil)
	serviceTrigger = TriggersService{}
	router.HandleFunc("/restarttriggers", serviceTrigger.StopTriggers)
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	resp, err := client.Get(ts.URL + "/restarttriggers?all=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")

	//no orgtrig

	resp, err = client.Get(ts.URL + "/restarttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

	//with orgtrig and no runtrig

	serviceTrigger = TriggersService{}
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

	//with orgtrig, with runtrig and no usertrig
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

	//with orgtrig, with runtrig and with usertrig
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}
	serviceTrigger.RtDtoUserSynk = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

}

func TestStopTriggersWhenEachTriggerAreInvalid(t *testing.T) {
	router := test.SetupBaseRouter(nil)
	serviceTrigger = TriggersService{}
	router.HandleFunc("/restarttriggers", serviceTrigger.StopTriggers)
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	resp, err := client.Get(ts.URL + "/restarttriggers?organizationsynktrigger")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

	//
	//with orgtrig and no runtrig

	serviceTrigger = TriggersService{}
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?runsFailedDiscoveryTrigger")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

	//with orgtrig, with runtrig and no usertrig

	serviceTrigger = TriggersService{}
	serviceTrigger.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}
	serviceTrigger.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan:      make(chan triggerDto.TriggerMessage, 1),
		IsRunning: false,
	}

	resp, err = client.Get(ts.URL + "/restarttriggers?usersSynkTrigger")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")

}
func TestStartTriggersWhenAllOk(t *testing.T) {
	setupTriggerMock(t)

	router := test.SetupBaseRouter(nil)
	router.HandleFunc("/starttriggers", serviceTrigger.StartTriggers)
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	resp, err := client.Get(ts.URL + "/starttriggers?all")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusOK, "http StatusCode is not OK")
}
func TestStartTriggersWhenParameterIsInvalid(t *testing.T) {
	setupTriggerMock(t)

	router := test.SetupBaseRouter(nil)
	router.HandleFunc("/starttriggers", serviceTrigger.StartTriggers)
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	resp, err := client.Get(ts.URL + "/starttriggers?all=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")

	resp, err = client.Get(ts.URL + "/starttriggers?organizationsynktrigger=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")

	resp, err = client.Get(ts.URL + "/starttriggers?runsFailedDiscoveryTrigger=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")

	resp, err = client.Get(ts.URL + "/starttriggers?usersSynkTrigger=invalid")

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity, "http StatusCode is not OK")
}
