package membersManager

import (
	"log"
	"strings"

	"wecode.sorint.it/opensource/papagaio-api/api/agola"
	agolaApi "wecode.sorint.it/opensource/papagaio-api/api/agola"
	"wecode.sorint.it/opensource/papagaio-api/api/git"
	"wecode.sorint.it/opensource/papagaio-api/api/git/gitlab"
	"wecode.sorint.it/opensource/papagaio-api/model"
)

//Return the users map by the agola remoteSource. Key is the git username and value agola userref
func getGitlabUsersMapByRemotesource(agolaApi agola.AgolaApiInterface, agolaRemoteSource string, gitUsers *[]gitlab.GitlabUser) *map[string]string {
	usersMap := make(map[string]string)

	remotesource, _ := agolaApi.GetRemoteSource(agolaRemoteSource)
	if remotesource == nil {
		return nil
	}

	for _, u := range *gitUsers {
		user, _ := agolaApi.GetUsersFilterbyRemoteUser(remotesource.ID, int64(u.ID))
		if len(user) == 1 {
			usersMap[u.Username] = user[0].Username
		}
	}

	return &usersMap
}

//Sincronizzo i membri della organization tra github e agola
func SyncMembersForGitlab(organization *model.Organization, gitSource *model.GitSource, agolaApi agolaApi.AgolaApiInterface, gitGateway *git.GitGateway, user *model.User) {
	gitlabUsers, _ := gitGateway.GitlabApi.GetOrganizationMembers(gitSource, user, organization.GitPath)
	agolaMembers, _ := agolaApi.GetOrganizationMembers(organization)

	agolaUsersMap := getGitlabUsersMapByRemotesource(agolaApi, gitSource.AgolaRemoteSource, gitlabUsers)

	for _, gitMember := range *gitlabUsers {
		agolaUserRef, usersExists := (*agolaUsersMap)[gitMember.Username]
		if !usersExists {
			continue
		}

		var role string
		if gitMember.HasOwnerPermission() {
			role = "owner"
		} else {
			role = "member"
		}
		err := agolaApi.AddOrUpdateOrganizationMember(organization, agolaUserRef, role)
		if err != nil {
			log.Println("AddOrUpdateOrganizationMember error:", err)
		}
	}

	//Verifico i membri eliminati su git
	for _, agolaMember := range agolaMembers.Members {
		if findGitlabMemberByAgolaUserRef(gitlabUsers, agolaUsersMap, agolaMember.User.Username) == nil {
			err := agolaApi.RemoveOrganizationMember(organization, agolaMember.User.Username)
			if err != nil {
				log.Println("RemoveOrganizationMember error:", err)
			}
		}
	}
}

func findGitlabMemberByAgolaUserRef(gitMembers *[]gitlab.GitlabUser, agolaUsersMap *map[string]string, agolaUserRef string) *gitlab.GitlabUser {
	for _, gitMember := range *gitMembers {
		if strings.Compare(agolaUserRef, (*agolaUsersMap)[gitMember.Username]) == 0 {
			return &gitMember
		}
	}

	return nil
}
