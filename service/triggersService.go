package service

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"wecode.sorint.it/opensource/papagaio-api/api/agola"
	"wecode.sorint.it/opensource/papagaio-api/api/git"
	"wecode.sorint.it/opensource/papagaio-api/dto"
	"wecode.sorint.it/opensource/papagaio-api/repository"
	"wecode.sorint.it/opensource/papagaio-api/trigger"
	"wecode.sorint.it/opensource/papagaio-api/utils"

	triggerDto "wecode.sorint.it/opensource/papagaio-api/trigger/dto"
)

type TriggersService struct {
	Db          repository.Database
	Tr          utils.ConfigUtils
	CommonMutex *utils.CommonMutex
	AgolaApi    agola.AgolaApiInterface
	GitGateway  *git.GitGateway

	RtDtoOrganizationSynk  *triggerDto.TriggerRunTimeDto
	RtDtoDiscoveryRunFails *triggerDto.TriggerRunTimeDto
	RtDtoUserSynk          *triggerDto.TriggerRunTimeDto
}

const ALL = "all"
const ORGANIZATION_SYNK_TRIGGER = "organizationsynktrigger"
const RUNS_FAILED_DISCOVERY_TRIGGER = "runsFailedDiscoveryTrigger"
const USERS_SYNK_TRIGGER = "usersSynkTrigger"

// @Summary Return time triggers
// @Description Get trigger timers
// @Tags Triggers
// @Produce  json
// @Success 200 {object} dto.ConfigTriggersDto "ok"
// @Router /gettriggersconfig [get]
// @Security ApiKeyToken
func (service *TriggersService) GetTriggersConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	dto := dto.ConfigTriggersDto{}
	dto.OrganizationsTriggerTime = service.Tr.GetOrganizationsTriggerTime()
	dto.RunFailedTriggerTime = service.Tr.GetRunFailedTriggerTime()
	dto.UsersTriggerTime = service.Tr.GetUsersTriggerTime()

	JSONokResponse(w, dto)
}

// @Summary Save time triggers
// @Description Save trigger timers
// @Tags Triggers
// @Produce  json
// @Param configTriggersDto body dto.ConfigTriggersDto true "Config triggers"
// @Success 200 "ok"
// @Router /savetriggersconfig [post]
// @Security ApiKeyToken
func (service *TriggersService) SaveTriggersConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var req dto.ConfigTriggersDto
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("encode error:", err)
		InternalServerError(w)
		return
	}

	if req.OrganizationsTriggerTime != 0 {
		err := service.Db.SaveOrganizationsTriggerTime(int(req.OrganizationsTriggerTime))
		if err != nil {
			log.Println("SaveOrganizationsTriggerTime error:", err)
		}
	}
	if req.RunFailedTriggerTime != 0 {
		err := service.Db.SaveRunFailedTriggerTime(int(req.RunFailedTriggerTime))
		if err != nil {
			log.Println("SaveRunFailedTriggerTime error:", err)
		}
	}
	if req.UsersTriggerTime != 0 {
		err := service.Db.SaveUsersTriggerTime(int(req.UsersTriggerTime))
		if err != nil {
			log.Println("SaveUsersTriggerTime error:", err)
		}
	}
}

// @Summary restart triggers
// @Description Restart timers
// @Tags Triggers
// @Produce  json
// @Param all query bool false "?all"
// @Param organizationsynktrigger query bool false "?organizationsynktrigger"
// @Param runsFailedDiscoveryTrigger query bool false "?runsFailedDiscoveryTrigger"
// @Param usersSynkTrigger query bool false "?usersSynkTrigger"
// @Success 200 "ok"
// @Router /restarttriggers [post]
// @Security ApiKeyToken
func (service *TriggersService) RestartTriggers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	all, err := getBoolParameter(r, ALL)
	if err != nil {
		UnprocessableEntityResponse(w, err.Error())
		return
	}
	if all {
		service.restartOrganizationSynkTrigger()
		service.restartRunsFailedDiscoveryTrigger()
		service.restartUsersSynkTrigger()

		return
	}

	organizationSynkTrigger, err := getBoolParameter(r, ORGANIZATION_SYNK_TRIGGER)
	if err != nil {
		UnprocessableEntityResponse(w, err.Error())
		return
	}
	if organizationSynkTrigger {
		err := service.restartOrganizationSynkTrigger()
		if err != nil {
			log.Println("error:", err)
			InternalServerError(w)
			return
		}
	}

	runsFailedDiscoveryTrigger, err := getBoolParameter(r, RUNS_FAILED_DISCOVERY_TRIGGER)
	if err != nil {
		UnprocessableEntityResponse(w, err.Error())
		return
	}
	if runsFailedDiscoveryTrigger {
		err = service.restartRunsFailedDiscoveryTrigger()
		if err != nil {
			log.Println("error:", err)
			InternalServerError(w)
			return
		}
	}

	usersSynkTrigger, err := getBoolParameter(r, USERS_SYNK_TRIGGER)
	if err != nil {
		UnprocessableEntityResponse(w, err.Error())
		return
	}
	if usersSynkTrigger {
		err = service.restartUsersSynkTrigger()
		if err != nil {
			log.Println("error:", err)
			InternalServerError(w)
			return
		}
	}
}

func (service *TriggersService) restartOrganizationSynkTrigger() error {
	if service.RtDtoOrganizationSynk == nil {
		return errors.New("OrganizationsTrigger nil")
	}
	if service.RtDtoOrganizationSynk.IsRunning {
		return errors.New("OrganizationsTrigger can't restart at the moment")
	}

	if !service.RtDtoOrganizationSynk.IsStopping && len(service.RtDtoOrganizationSynk.Chan) < cap(service.RtDtoOrganizationSynk.Chan) {
		service.RtDtoOrganizationSynk.Chan <- triggerDto.Restart
	}

	return nil
}

func (service *TriggersService) restartRunsFailedDiscoveryTrigger() error {
	if service.RtDtoDiscoveryRunFails == nil {
		return errors.New("RunsFailedDiscoveryTrigger nil")
	}
	if service.RtDtoDiscoveryRunFails.IsRunning {
		return errors.New("RunsFailedDiscoveryTrigger can't restart at the moment")
	}

	if !service.RtDtoDiscoveryRunFails.IsStopping && len(service.RtDtoDiscoveryRunFails.Chan) < cap(service.RtDtoDiscoveryRunFails.Chan) {
		service.RtDtoDiscoveryRunFails.Chan <- triggerDto.Restart
	}

	return nil
}

func (service *TriggersService) restartUsersSynkTrigger() error {
	if service.RtDtoUserSynk == nil {
		return errors.New("UsersSynkTrigger nil")
	}
	if service.RtDtoUserSynk.IsRunning {
		return errors.New("UsersSynkTrigger can't restart at the moment")
	}

	if !service.RtDtoUserSynk.IsStopping && len(service.RtDtoUserSynk.Chan) < cap(service.RtDtoUserSynk.Chan) {
		service.RtDtoUserSynk.Chan <- triggerDto.Restart
	}

	return nil
}

// @Summary get triggers status
// @Description Get triggers status
// @Tags Triggers
// @Produce  json
// @Success 200 "ok"
// @Router /triggersstatus [get]
// @Security ApiKeyToken
func (service *TriggersService) GetTriggersStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	retVal := dto.TriggersStatusDto{
		OrganizationStatus:      dto.TriggerDto{},
		DiscoveryRunFailsStatus: dto.TriggerDto{},
		UserSynkStatus:          dto.TriggerDto{},
	}

	if service.RtDtoOrganizationSynk != nil {
		retVal.OrganizationStatus.IsStarted = true
		retVal.OrganizationStatus.IsRunning = utils.NewBool(service.RtDtoOrganizationSynk.IsRunning)
		retVal.OrganizationStatus.LastRun = &service.RtDtoOrganizationSynk.LastRun
		retVal.OrganizationStatus.IsStopping = &service.RtDtoOrganizationSynk.IsStopping
		if !*retVal.OrganizationStatus.IsRunning {
			retVal.OrganizationStatus.TimeLeft = utils.NewUint(uint(time.Until(service.RtDtoOrganizationSynk.LastRun.Add(time.Duration(time.Minute.Nanoseconds() * int64(service.RtDtoOrganizationSynk.TriggerTime))))))
		}
	}

	if service.RtDtoDiscoveryRunFails != nil {
		retVal.DiscoveryRunFailsStatus.IsStarted = true
		retVal.DiscoveryRunFailsStatus.IsRunning = utils.NewBool(service.RtDtoDiscoveryRunFails.IsRunning)
		retVal.DiscoveryRunFailsStatus.LastRun = &service.RtDtoDiscoveryRunFails.LastRun
		retVal.DiscoveryRunFailsStatus.IsStopping = &service.RtDtoDiscoveryRunFails.IsStopping
		if !*retVal.DiscoveryRunFailsStatus.IsRunning {
			retVal.DiscoveryRunFailsStatus.TimeLeft = utils.NewUint(uint(time.Until(service.RtDtoDiscoveryRunFails.LastRun.Add(time.Duration(time.Minute.Nanoseconds() * int64(service.RtDtoDiscoveryRunFails.TriggerTime))))))
		}
	}

	if service.RtDtoUserSynk != nil {
		retVal.UserSynkStatus.IsStarted = true
		retVal.UserSynkStatus.IsRunning = utils.NewBool(service.RtDtoUserSynk.IsRunning)
		retVal.UserSynkStatus.LastRun = &service.RtDtoUserSynk.LastRun
		retVal.UserSynkStatus.IsStopping = &service.RtDtoUserSynk.IsStopping
		if !*retVal.UserSynkStatus.IsRunning {
			retVal.UserSynkStatus.TimeLeft = utils.NewUint(uint(time.Until(service.RtDtoUserSynk.LastRun.Add(time.Duration(time.Minute.Nanoseconds() * int64(service.RtDtoUserSynk.TriggerTime))))))
		}
	}

	JSONokResponse(w, retVal)
}

// @Summary stop triggers
// @Description Stop timers
// @Tags Triggers
// @Produce  json
// @Param all query bool false "?all"
// @Param organizationsynktrigger query bool false "?organizationsynktrigger"
// @Param runsFailedDiscoveryTrigger query bool false "?runsFailedDiscoveryTrigger"
// @Param usersSynkTrigger query bool false "?usersSynkTrigger"
// @Success 200 "ok"
// @Router /stoptriggers [post]
// @Security ApiKeyToken
func (service *TriggersService) StopTriggers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	all, err := getBoolParameter(r, ALL)
	if err != nil {
		UnprocessableEntityResponse(w, err.Error())
		return
	}
	if all {
		service.stopAll()

		return
	}

	organizationSynkTrigger, err := getBoolParameter(r, ORGANIZATION_SYNK_TRIGGER)
	if err != nil {
		UnprocessableEntityResponse(w, err.Error())
		return
	}
	if organizationSynkTrigger {
		service.stopOrganizationSynkTrigger()
	}

	runsFailedDiscoveryTrigger, err := getBoolParameter(r, RUNS_FAILED_DISCOVERY_TRIGGER)
	if err != nil {
		UnprocessableEntityResponse(w, err.Error())
		return
	}
	if runsFailedDiscoveryTrigger {
		service.stopRunsFailedDiscoveryTrigger()
	}

	usersSynkTrigger, err := getBoolParameter(r, USERS_SYNK_TRIGGER)
	if err != nil {
		UnprocessableEntityResponse(w, err.Error())
		return
	}
	if usersSynkTrigger {
		service.stopUsersSynkTrigger()
	}
}

func (service *TriggersService) stopAll() {
	if service.RtDtoOrganizationSynk != nil {
		service.RtDtoOrganizationSynk.IsStopping = true

		if len(service.RtDtoOrganizationSynk.Chan) < cap(service.RtDtoOrganizationSynk.Chan) {
			service.RtDtoOrganizationSynk.Chan <- triggerDto.Stop
		}
	}

	if service.RtDtoDiscoveryRunFails != nil {
		service.RtDtoDiscoveryRunFails.IsStopping = true

		if len(service.RtDtoDiscoveryRunFails.Chan) < cap(service.RtDtoDiscoveryRunFails.Chan) {
			service.RtDtoDiscoveryRunFails.Chan <- triggerDto.Stop
		}
	}

	if service.RtDtoUserSynk != nil {
		service.RtDtoUserSynk.IsStopping = true

		if len(service.RtDtoUserSynk.Chan) < cap(service.RtDtoUserSynk.Chan) {
			service.RtDtoUserSynk.Chan <- triggerDto.Stop
		}
	}
}

func (service *TriggersService) stopOrganizationSynkTrigger() {
	if service.RtDtoOrganizationSynk != nil {
		service.RtDtoOrganizationSynk.IsStopping = true

		if len(service.RtDtoOrganizationSynk.Chan) < cap(service.RtDtoOrganizationSynk.Chan) {
			service.RtDtoOrganizationSynk.Chan <- triggerDto.Stop
		}
	}
}

func (service *TriggersService) stopRunsFailedDiscoveryTrigger() {
	if service.RtDtoDiscoveryRunFails != nil {
		service.RtDtoDiscoveryRunFails.IsStopping = true

		if len(service.RtDtoDiscoveryRunFails.Chan) < cap(service.RtDtoDiscoveryRunFails.Chan) {
			service.RtDtoDiscoveryRunFails.Chan <- triggerDto.Stop
		}
	}
}

func (service *TriggersService) stopUsersSynkTrigger() {
	if service.RtDtoUserSynk != nil {
		service.RtDtoUserSynk.IsStopping = true

		if len(service.RtDtoUserSynk.Chan) < cap(service.RtDtoUserSynk.Chan) {
			service.RtDtoUserSynk.Chan <- triggerDto.Stop
		}
	}
}

// @Summary start triggers
// @Description Start timers
// @Tags Triggers
// @Produce  json
// @Param all query bool false "?all"
// @Param organizationsynktrigger query bool false "?organizationsynktrigger"
// @Param runsFailedDiscoveryTrigger query bool false "?runsFailedDiscoveryTrigger"
// @Param usersSynkTrigger query bool false "?usersSynkTrigger"
// @Success 200 "ok"
// @Router /starttriggers [post]
// @Security ApiKeyToken
func (service *TriggersService) StartTriggers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	all, err := getBoolParameter(r, ALL)
	if err != nil {
		UnprocessableEntityResponse(w, err.Error())
		return
	}
	if all {
		service.startOrganizationSynkTrigger()
		service.startRunsFailedDiscoveryTrigger()
		service.startUsersSynkTrigger()

		return
	}

	organizationSynkTrigger, err := getBoolParameter(r, ORGANIZATION_SYNK_TRIGGER)
	if err != nil {
		UnprocessableEntityResponse(w, err.Error())
		return
	}
	if organizationSynkTrigger {
		err := service.startOrganizationSynkTrigger()
		if err != nil {
			log.Println("error:", err)
			InternalServerError(w)
			return
		}
	}

	runsFailedDiscoveryTrigger, err := getBoolParameter(r, RUNS_FAILED_DISCOVERY_TRIGGER)
	if err != nil {
		UnprocessableEntityResponse(w, err.Error())
		return
	}
	if runsFailedDiscoveryTrigger {
		err = service.startRunsFailedDiscoveryTrigger()
		if err != nil {
			log.Println("error:", err)
			InternalServerError(w)
			return
		}
	}

	usersSynkTrigger, err := getBoolParameter(r, USERS_SYNK_TRIGGER)
	if err != nil {
		UnprocessableEntityResponse(w, err.Error())
		return
	}
	if usersSynkTrigger {
		err = service.startUsersSynkTrigger()
		if err != nil {
			log.Println("error:", err)
			InternalServerError(w)
			return
		}
	}
}

func (service *TriggersService) startOrganizationSynkTrigger() error {
	if service.RtDtoOrganizationSynk != nil {
		return errors.New("OrganizationsTrigger just started")
	}

	service.RtDtoOrganizationSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}

	trigger.StartOrganizationSync(service.Db, service.Tr, service.CommonMutex, service.AgolaApi, service.GitGateway, &service.RtDtoOrganizationSynk)

	return nil
}

func (service *TriggersService) startRunsFailedDiscoveryTrigger() error {
	if service.RtDtoDiscoveryRunFails != nil {
		return errors.New("RunsFailedDiscoveryTrigger just started")
	}

	service.RtDtoDiscoveryRunFails = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}

	trigger.StartRunFailsDiscovery(service.Db, service.Tr, service.CommonMutex, service.AgolaApi, service.GitGateway, &service.RtDtoDiscoveryRunFails)

	return nil
}

func (service *TriggersService) startUsersSynkTrigger() error {
	if service.RtDtoUserSynk != nil {
		return errors.New("UsersSynkTrigger just started")
	}

	service.RtDtoUserSynk = &triggerDto.TriggerRunTimeDto{
		Chan: make(chan triggerDto.TriggerMessage, 1),
	}

	trigger.StartSynkUsers(service.Db, service.Tr, service.CommonMutex, service.AgolaApi, service.GitGateway, &service.RtDtoUserSynk)

	return nil
}
