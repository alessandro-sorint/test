package membersManager

import (
	"log"
	"strings"

	"wecode.sorint.it/opensource/papagaio-api/api/agola"
	"wecode.sorint.it/opensource/papagaio-api/api/git"
	"wecode.sorint.it/opensource/papagaio-api/api/git/dto"
	"wecode.sorint.it/opensource/papagaio-api/model"
	"wecode.sorint.it/opensource/papagaio-api/utils"
)

//Sincronizzo i membri della organization tra gitea e agola
func SyncMembersForGitea(organization *model.Organization, gitSource *model.GitSource, agolaApi agola.AgolaApiInterface, gitGateway *git.GitGateway, user *model.User) {
	log.Println("SyncMembersForGitea start")

	gitTeams, err := gitGateway.GiteaApi.GetOrganizationTeams(gitSource, user, organization.GitPath)
	if err != nil {
		log.Println("error in GetOrganizationTeams:", err)
		return
	}
	gitTeamOwners := make(map[int64]dto.UserTeamResponseDto)
	gitTeamMembers := make(map[int64]dto.UserTeamResponseDto)

	for _, team := range *gitTeams {
		log.Println("team", team.Name, "owner permission:", team.HasOwnerPermission(), team.Permission)
		teamMembers, _ := gitGateway.GiteaApi.GetTeamMembers(gitSource, user, team.ID)

		var teamToCheck *map[int64]dto.UserTeamResponseDto
		if team.HasOwnerPermission() {
			teamToCheck = &gitTeamOwners
		} else {
			teamToCheck = &gitTeamMembers
		}

		for _, member := range *teamMembers {
			(*teamToCheck)[member.ID] = member
		}
	}

	// clear gitTeamMembers
	for userId := range gitTeamOwners {
		_, ok := gitTeamMembers[userId]
		if ok {
			delete(gitTeamMembers, userId)
		}
	}

	gitUsers := map[int64]dto.UserTeamResponseDto{}
	for k, v := range gitTeamMembers {
		gitUsers[k] = v
	}
	for k, v := range gitTeamOwners {
		gitUsers[k] = v
	}

	agolaOrganizationMembers, _ := agolaApi.GetOrganizationMembers(organization)
	agolaOrganizationMembersMap := toMapMembers(&agolaOrganizationMembers.Members)

	agolaUsersMap := utils.GetUsersMapByRemotesource(agolaApi, gitSource.AgolaRemoteSource, gitUsers)

	for _, gitMember := range gitTeamMembers {
		agolaUserRef, usersExists := (*agolaUsersMap)[gitMember.Username]
		if !usersExists {
			continue
		}

		if agolaMember, ok := (*agolaOrganizationMembersMap)[agolaUserRef]; !ok || agolaMember.Role == agola.Owner {
			err := agolaApi.AddOrUpdateOrganizationMember(organization, agolaUserRef, string(agola.Member))
			if err != nil {
				log.Println("AddOrUpdateOrganizationMember error:", err)
			}
		}
	}

	for _, gitMember := range gitTeamOwners {
		agolaUserRef, usersExists := (*agolaUsersMap)[gitMember.Username]
		if !usersExists {
			continue
		}

		if agolaMember, ok := (*agolaOrganizationMembersMap)[agolaUserRef]; !ok || agolaMember.Role == agola.Member {
			err := agolaApi.AddOrUpdateOrganizationMember(organization, agolaUserRef, string(agola.Owner))
			if err != nil {
				log.Println("AddOrUpdateOrganizationMember error:", err)
			}
		}
	}

	//Verifico i membri eliminati su git

	for _, agolaMember := range agolaOrganizationMembers.Members {
		if findGiteaMemberByAgolaUserRef(gitTeamOwners, agolaUsersMap, agolaMember.User.Username) == nil && findGiteaMemberByAgolaUserRef(gitTeamMembers, agolaUsersMap, agolaMember.User.Username) == nil {
			err := agolaApi.RemoveOrganizationMember(organization, agolaMember.User.Username)
			if err != nil {
				log.Println("RemoveOrganizationMember error:", err)
			}
		}
	}

	log.Println("SyncMembersForGitea end")
}

func findGiteaMemberByAgolaUserRef(gitMembers map[int64]dto.UserTeamResponseDto, agolaUsersMap *map[string]string, agolaUserRef string) *dto.UserTeamResponseDto {
	for _, gitMember := range gitMembers {
		if strings.Compare(agolaUserRef, (*agolaUsersMap)[gitMember.Username]) == 0 {
			return &gitMember
		}

	}

	return nil
}

func toMapMembers(members *[]agola.MemberDto) *map[string]agola.MemberDto {
	membersMap := make(map[string]agola.MemberDto)
	for _, member := range *members {
		membersMap[member.User.Username] = member
	}
	return &membersMap
}
