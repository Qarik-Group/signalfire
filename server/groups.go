package server

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/starkandwayne/signalfire/core"
)

type APIGroups struct {
	collator *core.Collator
}

func NewAPIGroups(collator *core.Collator) *APIGroups {
	return &APIGroups{collator: collator}
}

type APIGroupsResponse struct {
	Groups []APIGroupsGroup `json:"groups"`
}

type APIGroupsGroup struct {
	Name        string                `json:"name"`
	Deployments []APIGroupsDeployment `json:"deployments"`
	Releases    []APIGroupsRelease    `json:"releases"`
}

type APIGroupsDeployment struct {
	Name         string `json:"name"`
	ID           string `json:"id"`
	DirectorUUID string `json:"director_id"`
}

type APIGroupsRelease struct {
	Name     string             `json:"name"`
	Versions []APIGroupsVersion `json:"versions"`
}

type APIGroupsVersion struct {
	Version     string   `json:"version"`
	Deployments []string `json:"deployments"`
}

func (a *APIGroups) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	groups := a.collator.GetDeploymentGroups()
	responseObj := APIGroupsResponse{
		Groups: []APIGroupsGroup{},
	}
	for _, group := range groups {
		responseObj.Groups = append(
			responseObj.Groups,
			APIGroupsGroup{
				Name:        group.Name,
				Deployments: a.encodeDeployments(group.Deployments),
				Releases:    a.encodeReleases(group.Releases),
			},
		)
	}
	sort.Slice(responseObj.Groups,
		func(i, j int) bool {
			return responseObj.Groups[i].Name < responseObj.Groups[j].Name
		},
	)

	code := http.StatusOK
	out, err := json.Marshal(&responseObj)
	if err != nil {
		code = http.StatusInternalServerError
		out = []byte(InternalServerErrorMessagePayload)
	}

	writeResponseBytes(w, code, out)
}

func (a *APIGroups) encodeDeployments(deployments []core.CollationDeployment) []APIGroupsDeployment {
	ret := make([]APIGroupsDeployment, 0, len(deployments))
	for _, dep := range deployments {
		ret = append(ret, APIGroupsDeployment{
			ID:           dep.ID,
			Name:         dep.Name,
			DirectorUUID: dep.DirectorUUID,
		})
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].ID < ret[j].ID })
	return ret
}

func (a *APIGroups) encodeReleases(releases []core.CollationRelease) []APIGroupsRelease {
	ret := make([]APIGroupsRelease, 0, len(releases))
	for _, release := range releases {
		ret = append(ret, APIGroupsRelease{
			Name:     release.Name,
			Versions: a.encodeVersions(release.Versions),
		})
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Name < ret[j].Name })
	return ret
}

func (a *APIGroups) encodeVersions(versions []core.CollationReleaseVersion) []APIGroupsVersion {
	ret := make([]APIGroupsVersion, 0, len(versions))
	for _, version := range versions {
		ret = append(ret, APIGroupsVersion{
			Version:     version.Version,
			Deployments: version.Deployments,
		})
	}
	return ret
}
