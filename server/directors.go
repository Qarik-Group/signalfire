package server

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/starkandwayne/signalfire/core"
)

type APIDirectors struct {
	cache *core.Cache
}

func NewAPIDirectors(cache *core.Cache) *APIDirectors {
	return &APIDirectors{cache: cache}
}

type APIDirectorsResponse struct {
	Directors []APIDirectorsDirector `json:"directors"`
}

type APIDirectorsDirector struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

func (a *APIDirectors) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	envs := a.cache.GetEnvironments()
	responseObj := APIDirectorsResponse{
		Directors: []APIDirectorsDirector{},
	}
	for _, env := range envs {
		responseObj.Directors = append(
			responseObj.Directors,
			APIDirectorsDirector{
				Name: env.Name,
				UUID: env.UUID,
			},
		)
	}
	sort.Slice(responseObj.Directors,
		func(i, j int) bool {
			return responseObj.Directors[i].UUID < responseObj.Directors[j].UUID
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
