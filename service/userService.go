package service

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"wecode.sorint.it/opensource/papagaio-api/api/agola"
	"wecode.sorint.it/opensource/papagaio-api/controller"
	"wecode.sorint.it/opensource/papagaio-api/dto"
	"wecode.sorint.it/opensource/papagaio-api/model"
	"wecode.sorint.it/opensource/papagaio-api/repository"
)

type UserService struct {
	Db       repository.Database
	AgolaApi agola.AgolaApiInterface
}

func (service *UserService) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var requestDto dto.ChangeUserRoleRequestDto
	data, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(data, &requestDto)
	if err != nil {
		log.Println("unmarshal error:", err)
		InternalServerError(w)
		return
	}

	err = requestDto.IsValid()
	if err != nil {
		log.Println(err)
		InternalServerError(w)
		return
	}

	user, _ := service.Db.GetUserByUserId(*requestDto.UserID)
	if user == nil {
		log.Println("user", requestDto.UserID, "not found")
		InternalServerError(w)
		return
	}

	user.IsAdmin = requestDto.UserRole == dto.Administrator

	err = service.Db.SaveUser(user)
	if err != nil {
		log.Println("error in SaveUser:", err)
		InternalServerError(w)
	}
}

func (service *UserService) GetAllAgolaRunningRuns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	isAdmin := r.Context().Value(controller.AdminUserParameter).(bool)
	var user *model.User

	if !isAdmin {
		userId, _ := r.Context().Value(controller.UserIdParameter).(uint64)
		user, _ = service.Db.GetUserByUserId(userId)
		if user == nil {
			log.Println("User", userId, "not found")
			InternalServerError(w)
			return
		}
	}

	resp := make([]*agola.RunsDto, 0)

	//TODO valutare se vogliamo prendere tutte le org di agola o solo quelle presenti in papagaio
	//TODO valutare se vogliamo prendere anche i progetti dell'utente loggato per visualizzare le run
	//TODO utente admin deve avere la possibilità di vedere anche le run di tutti gli utenti?
	projectgrouprefs := make([]string, 0)
	var orgs []*agola.OrganizationDto
	var err error

	if isAdmin {
		orgs, err = service.AgolaApi.GetOrganizations()
		if err != nil {
			log.Println("GetUserOrganizations error:", err)
			InternalServerError(w)
			return
		}
		for _, org := range orgs {
			projectgrouprefs = append(projectgrouprefs, url.QueryEscape("org/"+org.Name))
		}

		users, err := service.AgolaApi.GetUsers()
		if err != nil {
			log.Println("GetUsers error:", err)
			InternalServerError(w)
			return
		}
		for _, user := range users {
			userProjectgroupref := url.QueryEscape("user/" + user.Username)
			projectgrouprefs = append(projectgrouprefs, userProjectgroupref)

			//directruns
			runs, err := service.AgolaApi.GetUserRuns(nil, true, user.Username, false, "running", nil, 0, false)
			if err != nil {
				log.Println("GetUserRuns error:", err)
				InternalServerError(w)
				return
			}

			for _, run := range runs {
				resp = append(resp, run)
			}
		}
	} else if user.AgolaUserRef != nil {
		userOrgs, err := service.AgolaApi.GetUserOrganizations(user, isAdmin)
		if err != nil {
			log.Println("GetUserOrganizations error:", err)
			InternalServerError(w)
			return
		}
		for _, userOrg := range userOrgs {
			projectgrouprefs = append(projectgrouprefs, url.QueryEscape("org/"+userOrg.Organization.Name))
		}

		agolaUser, err := service.AgolaApi.GetUser(*user.AgolaUserRef)
		if err != nil {
			log.Println("GetUser error:", err)
			InternalServerError(w)
			return
		}

		projectgrouprefs = append(projectgrouprefs, url.QueryEscape("user/"+agolaUser.Username))

		//directruns
		runs, err := service.AgolaApi.GetUserRuns(user, false, *user.AgolaUserRef, false, "running", nil, 0, false)
		if err != nil {
			log.Println("GetUserRuns error:", err)
			InternalServerError(w)
			return
		}

		for _, run := range runs {
			resp = append(resp, run)
		}
	}

	for _, projectgroupref := range projectgrouprefs {
		projects, err := service.getAllProjectgrouprefProjects(projectgroupref)
		if err != nil {
			log.Println("getAllProjectgrouprefProjects error:", err)
			InternalServerError(w)
			return
		}

		for _, project := range projects {
			runs, err := service.AgolaApi.GetRuns(project.ID, false, "running", nil, 0, false)
			if err != nil {
				log.Println("GetRuns error:", err)
				InternalServerError(w)
				return
			}

			for _, run := range runs {
				resp = append(resp, run)
			}
		}
	}

	JSONokResponse(w, resp)
}

func (service *UserService) getAllProjectgrouprefProjects(projectgroupref string) ([]*agola.ProjectDto, error) {
	resp := make([]*agola.ProjectDto, 0)

	projects, err := service.AgolaApi.GetProjectgroupProjects(projectgroupref)
	if err != nil {
		log.Println("GetProjectgroupProjects error:", err)
		return nil, err
	}
	for _, project := range projects {
		resp = append(resp, project)
	}

	subgroups, err := service.getAllProjectgroupSubgroups(projectgroupref)
	if err != nil {
		log.Println("getAllProjectgroupSubgroups error:", err)
		return nil, err
	}
	for _, subgroup := range subgroups {
		projects, err = service.AgolaApi.GetProjectgroupProjects(subgroup.ID)
		if err != nil {
			log.Println("GetProjectgroupProjects error:", err)
			return nil, err
		}
		for _, project := range projects {
			resp = append(resp, project)
		}
	}

	return resp, nil
}

func (service *UserService) getAllProjectgroupSubgroups(projectgroupref string) ([]*agola.ProjectGroupDto, error) {
	resp := make([]*agola.ProjectGroupDto, 0)

	subgroups, err := service.AgolaApi.GetProjectgroupSubgroups(projectgroupref)
	if err != nil {
		log.Println("GetProjectgroupSubgroups error:", err)
		return nil, err
	}
	for _, subgroup := range subgroups {
		resp = append(resp, subgroup)

		subSubgroups, err := service.getAllProjectgroupSubgroups(subgroup.ID)
		if err != nil {
			log.Println("getAllProjectgroupSubgroups error:", err)
			return nil, err
		}

		for _, subSubgroup := range subSubgroups {
			resp = append(resp, subSubgroup)
		}
	}

	return resp, nil
}
