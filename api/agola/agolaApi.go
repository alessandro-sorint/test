package agola

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"wecode.sorint.it/opensource/papagaio-api/api"
	"wecode.sorint.it/opensource/papagaio-api/config"
	"wecode.sorint.it/opensource/papagaio-api/model"
	"wecode.sorint.it/opensource/papagaio-api/repository"
	"wecode.sorint.it/opensource/papagaio-api/types"
)

type AgolaApiInterface interface {
	CheckOrganizationExists(organization *model.Organization) (bool, string, error)
	CheckProjectExists(organization *model.Organization, projectName string) (bool, string)
	CreateOrganization(organization *model.Organization, visibility types.VisibilityType) (string, error)
	DeleteOrganization(organization *model.Organization, user *model.User) error
	CreateProject(projectName string, agolaProjectRef string, organization *model.Organization, remoteSourceName string, user *model.User) (string, error)
	DeleteProject(organization *model.Organization, agolaProjectRef string, user *model.User) error
	AddOrUpdateOrganizationMember(organization *model.Organization, agolaUserRef string, role string) error
	RemoveOrganizationMember(organization *model.Organization, agolaUserRef string) error
	GetOrganizationMembers(organization *model.Organization) (*OrganizationMembersResponseDto, error)
	ArchiveProject(organization *model.Organization, agolaProjectRef string) error
	UnarchiveProject(organization *model.Organization, agolaProjectRef string) error
	GetRuns(projectRef string, lastRun bool, phase string, startRunNumber *uint64, limit uint, asc bool) ([]*RunsDto, error)
	GetRun(projectRef string, runNumber uint64) (*RunDto, error)
	GetTask(projectRef string, runNumber uint64, taskID string) (*TaskDto, error)
	GetLogs(projectRef string, runNumber uint64, taskID string, step int) (string, error)
	GetRemoteSource(agolaRemoteSource string) (*RemoteSourceDto, error)
	GetUsers() ([]*UserDto, error)
	GetUser(userRef string) (*UserDto, error)
	GetUsersFilterbyRemoteUser(remoteSourceID string, remoteUserID int64) ([]*UserDto, error)
	GetOrganizations() ([]*OrganizationDto, error)
	GetUserOrganizations(user *model.User, isAdminUser bool) ([]*UserOrgDto, error)
	GetProjectgroupProjects(projectgroupref string) ([]*ProjectDto, error)
	GetUserRuns(user *model.User, isAdminUser bool, userRef string, lastRun bool, phase string, startRunNumber *uint64, limit uint, asc bool) ([]*RunsDto, error)
	GetProjectgroupSubgroups(projectgroupref string) ([]*ProjectGroupDto, error)

	CreateUserToken(user *model.User) error
	GetRemoteSources() (*[]RemoteSourceDto, error)
	CreateRemoteSource(remoteSourceName string, gitType string, apiUrl string, oauth2ClientId string, oauth2ClientSecret string) error
	DeleteRemotesource(remoteSourceName string) error
}

type AgolaApi struct {
	Db repository.Database
}

const baseTokenName string = "papagaioToken"

func (agolaApi *AgolaApi) GetOrganizations() ([]*OrganizationDto, error) {
	client := agolaApi.getClient(nil, true)
	URLApi := getOrganizationsUrl()

	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var jsonResponse []*OrganizationDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return jsonResponse, err
}

func (agolaApi *AgolaApi) CheckOrganizationExists(organization *model.Organization) (bool, string, error) {
	client := agolaApi.getClient(nil, true)
	URLApi := getOrganizationUrl(organization.AgolaOrganizationRef)

	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}

	defer resp.Body.Close()

	var organizationID string
	organizationExists := api.IsResponseOK(resp.StatusCode)
	if organizationExists {
		body, _ := ioutil.ReadAll(resp.Body)
		var jsonResponse AgolaCreateORGDto
		err := json.Unmarshal(body, &jsonResponse)
		if err != nil {
			return false, "", nil
		}

		organizationID = jsonResponse.ID
	}

	return organizationExists, organizationID, nil
}

func (agolaApi *AgolaApi) CheckProjectExists(organization *model.Organization, agolaProjectRef string) (bool, string) {
	log.Println("CheckProjectExists start")

	client := agolaApi.getClient(nil, true)
	URLApi := getProjectUrl(organization.AgolaOrganizationRef, agolaProjectRef)
	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return false, ""
	}
	defer resp.Body.Close()

	var projectID string
	projectExists := api.IsResponseOK(resp.StatusCode)
	if projectExists {
		body, _ := ioutil.ReadAll(resp.Body)
		var jsonResponse CreateProjectResponseDto
		err = json.Unmarshal(body, &jsonResponse)
		if err != nil {
			return false, ""
		}

		projectID = jsonResponse.ID
	}

	return projectExists, projectID
}

func (agolaApi *AgolaApi) CreateOrganization(organization *model.Organization, visibility types.VisibilityType) (string, error) {
	client := agolaApi.getClient(nil, true)
	URLApi := getOrgUrl()
	reqBody := strings.NewReader(`{"name": "` + organization.AgolaOrganizationRef + `", "visibility": "` + string(visibility) + `"}`)
	req, _ := http.NewRequest("POST", URLApi, reqBody)
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return "", errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var jsonResponse AgolaCreateORGDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return "", err
	}

	return jsonResponse.ID, err
}

func (agolaApi *AgolaApi) DeleteOrganization(organization *model.Organization, user *model.User) error {
	client := agolaApi.getClient(user, false)
	URLApi := getOrganizationUrl(organization.AgolaOrganizationRef)
	req, _ := http.NewRequest("DELETE", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(respMessage))
	}

	return nil
}

func (agolaApi *AgolaApi) CreateProject(projectName string, agolaProjectRef string, organization *model.Organization, remoteSourceName string, user *model.User) (string, error) {
	log.Println("CreateProject start")

	if exists, projectID := agolaApi.CheckProjectExists(organization, agolaProjectRef); exists {
		log.Println("project already exists with ID:", projectID)
		return projectID, nil
	}
	client := agolaApi.getClient(user, false)
	URLApi := getCreateProjectUrl()

	projectRequest := &CreateProjectRequestDto{
		Name:             agolaProjectRef,
		ParentRef:        "org/" + organization.AgolaOrganizationRef,
		Visibility:       organization.Visibility,
		RemoteSourceName: remoteSourceName,
		RepoPath:         organization.GitPath + "/" + projectName,
	}

	data, _ := json.Marshal(projectRequest)
	reqBody := strings.NewReader(string(data))

	req, _ := http.NewRequest("POST", URLApi, reqBody)
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return "", errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var jsonResponse CreateProjectResponseDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return "", err
	}

	return jsonResponse.ID, err
}

func (agolaApi *AgolaApi) DeleteProject(organization *model.Organization, agolaProjectRef string, user *model.User) error {
	log.Println("DeleteProject start")

	client := agolaApi.getClient(user, false)
	URLApi := getProjectUrl(organization.AgolaOrganizationRef, agolaProjectRef)
	req, _ := http.NewRequest("DELETE", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(respMessage))
	}

	log.Println("DeleteProject end")

	return err
}

func (agolaApi *AgolaApi) AddOrUpdateOrganizationMember(organization *model.Organization, agolaUserRef string, role string) error {
	log.Println("AddOrUpdateOrganizationMember start")

	log.Println("AddOrUpdateOrganizationMember", agolaUserRef, "for", organization.GitName, "with role:", role)

	var err error
	client := agolaApi.getClient(nil, true)
	URLApi := getAddOrgMemberUrl(organization.AgolaOrganizationRef, agolaUserRef)
	reqBody := strings.NewReader(`{"role": "` + role + `"}`)
	req, _ := http.NewRequest("PUT", URLApi, reqBody)
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(respMessage))
	}

	log.Println("AddOrUpdateOrganizationMember end")

	return err
}

func (agolaApi *AgolaApi) RemoveOrganizationMember(organization *model.Organization, agolaUserRef string) error {
	log.Println("RemoveOrganizationMember", organization.GitName, "with agolaUserRef", agolaUserRef)

	var err error
	client := agolaApi.getClient(nil, true)
	URLApi := getAddOrgMemberUrl(organization.AgolaOrganizationRef, agolaUserRef)

	reqBody := strings.NewReader(`{}`)
	req, _ := http.NewRequest("DELETE", URLApi, reqBody)
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println("RemoveOrganizationMember StatusCode:", resp.StatusCode)

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		log.Println("RemoveOrganizationMember respMessage:", string(respMessage))
		return errors.New(string(respMessage))
	}

	return errors.New("response status: " + resp.Status)
}

func (agolaApi *AgolaApi) GetOrganizationMembers(organization *model.Organization) (*OrganizationMembersResponseDto, error) {
	log.Println("GetOrganizationMembers start")

	client := agolaApi.getClient(nil, true)
	URLApi := getOrganizationMembersUrl(organization.AgolaOrganizationRef)
	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var jsonResponse OrganizationMembersResponseDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return &jsonResponse, err
}

//TODO after Agola Issue
func (agolaApi *AgolaApi) ArchiveProject(organization *model.Organization, projectName string) error {
	log.Println("ArchiveProject:", organization.AgolaOrganizationRef, projectName)

	return nil
}

//TODO after Agola Issue
func (agolaApi *AgolaApi) UnarchiveProject(organization *model.Organization, projectName string) error {
	log.Println("UnarchiveProject:", organization.AgolaOrganizationRef, projectName)

	return nil
}

func (agolaApi *AgolaApi) GetRuns(projectRef string, lastRun bool, phase string, startRunNumber *uint64, limit uint, asc bool) ([]*RunsDto, error) {
	log.Println("GetRuns start:", projectRef)

	client := agolaApi.getClient(nil, true)
	URLApi := getRunsListUrl(projectRef, lastRun, phase, startRunNumber, limit, asc)

	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var jsonResponse []*RunsDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return jsonResponse, err
}

func (agolaApi *AgolaApi) GetRun(projectRef string, runNumber uint64) (*RunDto, error) {
	log.Println("GetRuns start")

	client := agolaApi.getClient(nil, true)
	URLApi := getRunUrl(projectRef, runNumber)
	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var jsonResponse RunDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return &jsonResponse, err
}

func (agolaApi *AgolaApi) GetTask(projectRef string, runNumber uint64, taskID string) (*TaskDto, error) {
	log.Println("GetRuns start")

	client := agolaApi.getClient(nil, true)
	URLApi := getTaskUrl(projectRef, runNumber, taskID)
	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var jsonResponse TaskDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return &jsonResponse, err
}

func (agolaApi *AgolaApi) GetLogs(projectRef string, runNumber uint64, taskID string, step int) (string, error) {
	log.Println("GetRuns start")

	client := agolaApi.getClient(nil, true)
	URLApi := getLogsUrl(projectRef, runNumber, taskID, step)
	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return "", errors.New(string(respMessage))
	}

	logs, _ := ioutil.ReadAll(resp.Body)

	return string(logs), err
}

func (agolaApi *AgolaApi) GetRemoteSource(agolaRemoteSource string) (*RemoteSourceDto, error) {
	log.Println("GetRemoteSource start")

	client := agolaApi.getClient(nil, true)
	URLApi := getRemoteSourceUrl(agolaRemoteSource)

	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var jsonResponse RemoteSourceDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return &jsonResponse, nil
}

func (agolaApi *AgolaApi) getUsers(start string, limit uint) ([]*UserDto, error) {
	log.Println("GetRemoteSource start")

	client := agolaApi.getClient(nil, true)
	URLApi := getUsersUrl(start, limit)

	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var jsonResponse []*UserDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return jsonResponse, nil
}

func (agolaApi *AgolaApi) GetUser(userRef string) (*UserDto, error) {
	client := agolaApi.getClient(nil, true)
	URLApi := getUserUrl(userRef)

	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var jsonResponse UserDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return &jsonResponse, nil
}

func (agolaApi *AgolaApi) GetUsersFilterbyRemoteUser(remoteSourceID string, remoteUserID int64) ([]*UserDto, error) {
	log.Println("GetRemoteSource start")

	client := agolaApi.getClient(nil, true)
	URLApi := getUsersFilterbyRemoteUserUrl("", 0, remoteSourceID, remoteUserID)

	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var jsonResponse []*UserDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return jsonResponse, nil
}

const usersLimit = 20

func (agolaApi *AgolaApi) GetUsers() ([]*UserDto, error) {
	retVal := make([]*UserDto, 0)

	start := ""
	for {
		users, err := agolaApi.getUsers(start, usersLimit)
		if err != nil {
			return nil, err
		}

		retVal = append(retVal, users...)
		if len(users) < usersLimit {
			break
		}

		start = users[usersLimit-1].Username
	}

	return retVal, nil
}

func (agolaApi *AgolaApi) CreateUserToken(user *model.User) error {
	if user == nil || user.AgolaUserRef == nil {
		log.Println("CreateUserToken error user nil")
		return errors.New("user nil error")
	}

	log.Println("RefreshAgolaUserToken user", *user.AgolaUserRef)

	if user.UserID == nil {
		log.Println("UserID is nil")
		return errors.New("UsersID is nil")
	}
	if user.AgolaUserRef == nil {
		log.Println("AgolaUserRef is nil")
		return errors.New("AgolaUserRef is nil")
	}

	tokenName := baseTokenName + "-" + fmt.Sprint(time.Now().Unix())
	user.AgolaTokenName = &tokenName

	client := &http.Client{}
	URLApi := getCreateTokenUrl(*user.AgolaUserRef)

	tokenRequest := &TokenRequestDto{
		TokenName: *user.AgolaTokenName,
	}
	data, _ := json.Marshal(tokenRequest)
	reqBody := strings.NewReader(string(data))

	req, _ := http.NewRequest("POST", URLApi, reqBody)
	req.Header.Set("Authorization", "token "+config.Config.Agola.AdminToken)
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var jsonResponse TokenResponseDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return err
	}

	user.AgolaToken = &jsonResponse.Token
	err = agolaApi.Db.SaveUser(user)

	return err
}

func (agolaApi *AgolaApi) GetRemoteSources() (*[]RemoteSourceDto, error) {
	log.Println("GetRemoteSources start")

	client := agolaApi.getClient(nil, true)
	URLApi := getRemoteSourcesUrl()

	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var jsonResponse []RemoteSourceDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return &jsonResponse, nil
}

func (agolaApi *AgolaApi) CreateRemoteSource(remoteSourceName string, gitType string, apiUrl string, oauth2ClientId string, oauth2ClientSecret string) error {
	log.Println("CreateRemoteSource start")

	client := agolaApi.getClient(nil, true)
	URLApi := getRemoteSourcesUrl()

	projectRequest := &CreateRemoteSourceRequestDto{
		Name:                remoteSourceName,
		APIURL:              apiUrl,
		Type:                gitType,
		AuthType:            "oauth2",
		SkipSSHHostKeyCheck: true,
		SkipVerify:          false,
		Oauth2ClientID:      oauth2ClientId,
		Oauth2ClientSecret:  oauth2ClientSecret,
	}
	data, _ := json.Marshal(projectRequest)
	reqBody := strings.NewReader(string(data))

	req, _ := http.NewRequest("POST", URLApi, reqBody)
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(respMessage))
	}

	return nil
}

func (agolaApi *AgolaApi) DeleteRemotesource(remoteSourceName string) error {
	log.Println("DeleteRemotesource ", remoteSourceName)

	client := agolaApi.getClient(nil, true)
	URLApi := getDeleteRemotesourceUrl(remoteSourceName)

	req, _ := http.NewRequest("DELETE", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(respMessage))
	}

	return nil
}

func (agolaApi *AgolaApi) GetUserOrganizations(user *model.User, isAdminUser bool) ([]*UserOrgDto, error) {
	client := agolaApi.getClient(user, isAdminUser)
	URLApi := getUserOrganizationsUrl()

	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var jsonResponse []*UserOrgDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return jsonResponse, nil
}

func (agolaApi *AgolaApi) GetProjectgroupProjects(projectgroupref string) ([]*ProjectDto, error) {
	client := agolaApi.getClient(nil, true)
	URLApi := getProjectgroupProjectsUrl(projectgroupref)

	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var jsonResponse []*ProjectDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return jsonResponse, nil
}

func (agolaApi *AgolaApi) GetUserRuns(user *model.User, isAdminUser bool, userRef string, lastRun bool, phase string, startRunNumber *uint64, limit uint, asc bool) ([]*RunsDto, error) {
	client := agolaApi.getClient(user, isAdminUser)
	URLApi := getUserRunsUrl(userRef, lastRun, phase, startRunNumber, limit, asc)

	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var jsonResponse []*RunsDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return jsonResponse, nil
}

func (agolaApi *AgolaApi) GetProjectgroupSubgroups(projectgroupref string) ([]*ProjectGroupDto, error) {
	client := agolaApi.getClient(nil, true)
	URLApi := getSubgroupsUrl(projectgroupref)

	req, _ := http.NewRequest("GET", URLApi, nil)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !api.IsResponseOK(resp.StatusCode) {
		respMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(respMessage))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var jsonResponse []*ProjectGroupDto
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return jsonResponse, nil
}

///////////////

func (agolaApi *AgolaApi) getClient(user *model.User, isAdminUser bool) *httpClient {
	client := &httpClient{c: &http.Client{}, user: user, agolaApi: agolaApi, isAdminUser: isAdminUser}

	return client
}

type httpClient struct {
	c           *http.Client
	user        *model.User
	agolaApi    *AgolaApi
	isAdminUser bool
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	if c.isAdminUser {
		req.Header.Set("Authorization", "token "+config.Config.Agola.AdminToken)
		return c.c.Do(req)
	}

	var response *http.Response
	var err error

	if c.user.AgolaToken != nil {
		req.Header.Set("Authorization", "token "+*c.user.AgolaToken)
		response, err = c.c.Do(req)
		if err != nil {
			return nil, err
		}
	}

	if response == nil || response.StatusCode == 401 {
		err = c.agolaApi.CreateUserToken(c.user)
		if err != nil {
			log.Println("error in agola CreateUserToken:", err)
			return nil, err
		}

		req.Header.Set("Authorization", "token "+*c.user.AgolaToken)
		response, err = c.c.Do(req)
	}

	return response, err
}
