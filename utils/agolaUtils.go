package utils

import (
	"strings"

	"wecode.sorint.it/opensource/papagaio-api/api/agola"
	"wecode.sorint.it/opensource/papagaio-api/api/git/dto"
	"wecode.sorint.it/opensource/papagaio-api/config"
	"wecode.sorint.it/opensource/papagaio-api/model"
)

func GetOrganizationUrl(organization *model.Organization) string {
	return config.Config.Agola.AgolaAddr + "/org/" + organization.AgolaOrganizationRef
}

func GetProjectUrl(organization *model.Organization, project *model.Project) *string {
	url := config.Config.Agola.AgolaAddr + "/org/" + organization.AgolaOrganizationRef + "/projects/" + project.AgolaProjectRef + ".proj"
	return &url
}

func ConvertToAgolaProjectRef(projectName string) string {
	agolaProjectName := strings.ReplaceAll(projectName, ".", "")
	agolaProjectName = strings.ReplaceAll(agolaProjectName, "_", "")

	return agolaProjectName
}

//Return the users map by the agola remoteSource. Key is the git username and value agola userref
func GetUsersMapByRemotesource(agolaApi agola.AgolaApiInterface, agolaRemoteSource string, gitUsers map[int64]dto.UserTeamResponseDto) *map[string]string {
	usersMap := make(map[string]string)

	remotesource, _ := agolaApi.GetRemoteSource(agolaRemoteSource)
	if remotesource == nil {
		return nil
	}

	for _, u := range gitUsers {
		user, _ := agolaApi.GetUsersFilterbyRemoteUser(remotesource.ID, u.ID)
		if len(user) == 1 {
			usersMap[u.Username] = user[0].Username
		}
	}

	return &usersMap
}

func GetAgolaUserRefByGitUserID(agolaApi agola.AgolaApiInterface, agolaRemoteSource string, gitUserID int64) *string {
	remotesource, _ := agolaApi.GetRemoteSource(agolaRemoteSource)
	if remotesource == nil {
		return nil
	}

	users, _ := agolaApi.GetUsersFilterbyRemoteUser(remotesource.ID, gitUserID)
	if len(users) == 1 {
		return &users[0].Username
	}

	return nil
}
