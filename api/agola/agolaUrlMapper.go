package agola

import (
	"fmt"
	"net/url"

	"wecode.sorint.it/opensource/papagaio-api/config"
)

const organizationPath string = "%s/api/v1alpha/orgs/%s"
const orgPath string = "%s/api/v1alpha/orgs"
const createMemberPath string = "%s/api/v1alpha/orgs/%s/members/%s"
const createProjectPath string = "%s/api/v1alpha/projects"
const projectPath string = "%s/api/v1alpha/projects/%s"
const organizationMembersPath string = "%s/api/v1alpha/orgs/%s/members"
const runsListPath string = "%s/api/v1alpha/projects/%s/runs?%s"
const runPath string = "%s/api/v1alpha/projects/%s/runs/%d"
const taskPath string = "%s/api/v1alpha/projects/%s/runs/%d/tasks/%s"
const logsPath string = "%s/api/v1alpha/projects/%s/runs/%d/tasks/%s/logs?%s"
const remoteSourcePath string = "%s/api/v1alpha/remotesources/%s"
const remoteSourcesPath string = "%s/api/v1alpha/remotesources"
const usersPath string = "%s/api/v1alpha/users?start=%s&limit=%d"
const userPath string = "%s/api/v1alpha/users/%s"
const usersfilterbyremoteuserPath string = "%s/api/v1alpha/users?start=%s&limit=%d&byremoteuser&remotesourceid=%s&remoteuserid=%d"
const organizationsPath = "%s/api/v1alpha/orgs"
const deleteRemotesourcePath = "%s/api/v1alpha/remotesources/%s"
const userOrganizationsPath = "%s/api/v1alpha/user/orgs"
const projectgroupProjectsPath = "%s/api/v1alpha/projectgroups/%s/projects"
const userRunsPath = "%s/api/v1alpha/users/%s/runs?%s"
const subgroupsPath = "%s/api/v1alpha/projectgroups/%s/subgroups"

const createTokenPath = "%s/api/v1alpha/users/%s/tokens"

func getOrganizationsUrl() string {
	return fmt.Sprintf(organizationsPath, config.Config.Agola.AgolaAddr)
}

func getOrganizationUrl(agolaOrganizationRef string) string {
	return fmt.Sprintf(organizationPath, config.Config.Agola.AgolaAddr, agolaOrganizationRef)
}

func getOrgUrl() string {
	return fmt.Sprintf(orgPath, config.Config.Agola.AgolaAddr)
}

func getAddOrgMemberUrl(agolaOrganizationRef string, agolaUserRef string) string {
	return fmt.Sprintf(createMemberPath, config.Config.Agola.AgolaAddr, agolaOrganizationRef, agolaUserRef)
}

func getCreateProjectUrl() string {
	return fmt.Sprintf(createProjectPath, config.Config.Agola.AgolaAddr)
}

func getProjectUrl(organizationName string, projectName string) string {
	projectref := url.QueryEscape("org/" + organizationName + "/" + projectName)
	return fmt.Sprintf(projectPath, config.Config.Agola.AgolaAddr, projectref)
}

func getOrganizationMembersUrl(organizationName string) string {
	return fmt.Sprintf(organizationMembersPath, config.Config.Agola.AgolaAddr, organizationName)
}

func getRunsListUrl(projectRef string, lastRun bool, phase string, startRunNumber *uint64, limit uint, asc bool) string {
	query := ""
	if lastRun {
		query += "&lastrun"
	}
	if len(phase) > 0 {
		query += "&phase=" + phase
	}
	if startRunNumber != nil {
		query += "&start=" + fmt.Sprint(*startRunNumber)
	}
	if limit > 0 {
		query += "&limit=" + fmt.Sprint(limit)
	}
	if asc {
		query += "&asc"
	}

	return fmt.Sprintf(runsListPath, config.Config.Agola.AgolaAddr, projectRef, query)
}

func getRunUrl(projectRef string, runNumber uint64) string {
	return fmt.Sprintf(runPath, config.Config.Agola.AgolaAddr, projectRef, runNumber)
}

func getTaskUrl(projectRef string, runNumber uint64, taskID string) string {
	return fmt.Sprintf(taskPath, config.Config.Agola.AgolaAddr, projectRef, runNumber, taskID)
}

func getLogsUrl(projectRef string, runNumber uint64, taskID string, step int) string {
	stepParam := "setup"
	if step != -1 {
		stepParam = "step=" + fmt.Sprint(step)
	}

	return fmt.Sprintf(logsPath, config.Config.Agola.AgolaAddr, projectRef, runNumber, taskID, stepParam)
}

func getRemoteSourceUrl(agolaRemoteSource string) string {
	return fmt.Sprintf(remoteSourcePath, config.Config.Agola.AgolaAddr, agolaRemoteSource)
}

func getUsersUrl(start string, limit uint) string {
	return fmt.Sprintf(usersPath, config.Config.Agola.AgolaAddr, start, limit)
}

func getUserUrl(userRef string) string {
	return fmt.Sprintf(userPath, config.Config.Agola.AgolaAddr, userRef)
}

func getUsersFilterbyRemoteUserUrl(start string, limit uint, remoteSourceID string, remoteUserID int64) string {
	return fmt.Sprintf(usersfilterbyremoteuserPath, config.Config.Agola.AgolaAddr, start, limit, remoteSourceID, remoteUserID)
}

func getCreateTokenUrl(agolaUserName string) string {
	return fmt.Sprintf(createTokenPath, config.Config.Agola.AgolaAddr, agolaUserName)
}

func getRemoteSourcesUrl() string {
	return fmt.Sprintf(remoteSourcesPath, config.Config.Agola.AgolaAddr)
}

func getDeleteRemotesourceUrl(agolaRemoteSource string) string {
	return fmt.Sprintf(deleteRemotesourcePath, config.Config.Agola.AgolaAddr, agolaRemoteSource)
}

func getUserOrganizationsUrl() string {
	return fmt.Sprintf(userOrganizationsPath, config.Config.Agola.AgolaAddr)
}

func getProjectgroupProjectsUrl(projectgroupref string) string {
	return fmt.Sprintf(projectgroupProjectsPath, config.Config.Agola.AgolaAddr, projectgroupref)
}

func getUserRunsUrl(userRef string, lastRun bool, phase string, startRunNumber *uint64, limit uint, asc bool) string {
	query := ""
	if lastRun {
		query += "&lastrun"
	}
	if len(phase) > 0 {
		query += "&phase=" + phase
	}
	if startRunNumber != nil {
		query += "&start=" + fmt.Sprint(*startRunNumber)
	}
	if limit > 0 {
		query += "&limit=" + fmt.Sprint(limit)
	}
	if asc {
		query += "&asc"
	}

	return fmt.Sprintf(userRunsPath, config.Config.Agola.AgolaAddr, userRef, query)
}

func getSubgroupsUrl(projectgroupref string) string {
	return fmt.Sprintf(subgroupsPath, config.Config.Agola.AgolaAddr, projectgroupref)
}
