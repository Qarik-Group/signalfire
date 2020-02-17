package core

import "sync"

type Cache struct {
	data      []CacheEnvironment
	lock      sync.RWMutex
	listeners []chan bool
}

type CacheEnvironment struct {
	Name        string
	UUID        string
	Deployments CacheDeployments
}

type CacheDeployment struct {
	Name     string
	Releases CacheReleases
}

type CacheDeployments []CacheDeployment

func (deployments CacheDeployments) Copy() CacheDeployments {
	ret := make(CacheDeployments, 0, len(deployments))
	for i := range deployments {
		ret = append(ret, CacheDeployment{
			Name:     deployments[i].Name,
			Releases: deployments[i].Releases.Copy(),
		})
	}
	return ret
}

type CacheRelease struct {
	Name    string
	Version string
}

type CacheReleases []CacheRelease

func (releases CacheReleases) Copy() CacheReleases {
	ret := make(CacheReleases, 0, len(releases))
	for _, release := range releases {
		ret = append(ret, release)
	}
	return ret
}

func NewCache() *Cache {
	return &Cache{}
}

//AddListener registers a channel that will get the
func (c *Cache) AddListener(ch chan bool) {
	c.lock.Lock()
	c.listeners = append(c.listeners, ch)
	c.lock.Unlock()
}

func (c *Cache) notifyListeners() {
	for i := range c.listeners {
		c.listeners[i] <- true
	}
}

func (c *Cache) UpdateEnvironment(e CacheEnvironment) {
	c.lock.Lock()
	idx := c.findEnvironmentIdx(cacheEnvironmentQuery{UUID: e.UUID})
	if idx < 0 {
		c.data = append(c.data, e)
	} else {
		c.data[idx] = e
	}
	c.lock.Unlock()
	c.notifyListeners()
}

type cacheEnvironmentQuery struct {
	Name string
	UUID string
}

//Returns negative if not found
func (c *Cache) findEnvironmentIdx(q cacheEnvironmentQuery) int {
	ret := -1
	if q.Name == "" && q.UUID == "" {
		return ret
	}

	for i, e := range c.data {
		if (q.Name == "" || e.Name == q.Name) &&
			(q.UUID == "" || e.UUID == q.UUID) {
			ret = i
			break
		}
	}

	return ret
}

func (c *Cache) GetEnvironments() []CacheEnvironment {
	ret := make([]CacheEnvironment, 0, len(c.data))
	c.lock.RLock()
	//Deep copy each environment
	for _, env := range c.data {
		toAdd := CacheEnvironment{
			Name:        env.Name,
			UUID:        env.UUID,
			Deployments: env.Deployments.Copy(),
		}

		ret = append(ret, toAdd)
	}
	c.lock.RUnlock()
	return ret
}
