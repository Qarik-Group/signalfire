package core

import (
	"fmt"
	"sync"
)

type Collator struct {
	groups      []CollationDeploymentGroup
	idsToGroups map[string]string
	rules       []CollationRule
	lock        sync.RWMutex
}

func NewCollator() *Collator {
	return &Collator{idsToGroups: make(map[string]string)}
}

type CollationDeploymentInput struct {
	DirectorUUID   string
	DirectorName   string
	DeploymentName string
	Releases       CacheReleases
}

func (c *CollationDeploymentInput) calcID() string {
	return fmt.Sprintf("%s/%s", c.DirectorUUID, c.DeploymentName)
}

func (c *Collator) WatchAsync(cache *Cache) {
	listenChan := make(chan bool)
	cache.AddListener(listenChan)
	go func() {
		for range listenChan {
			c.collate(cache.GetEnvironments())
		}
	}()
}

func (c *Collator) AddRule(rule CollationRule) {
	c.lock.Lock()
	c.rules = append(c.rules, rule)
	c.lock.Unlock()
}

func (c *Collator) GetDeploymentGroups() []CollationDeploymentGroup {
	c.lock.RLock()
	ret := make([]CollationDeploymentGroup, 0, len(c.groups))
	for _, deploymentGroup := range c.groups {
		ret = append(ret, deploymentGroup)
	}
	c.lock.RUnlock()
	return ret
}

func (c *Collator) collate(envs []CacheEnvironment) {
	c.lock.Lock()
	c.groups = []CollationDeploymentGroup{}
	c.idsToGroups = map[string]string{}
	for _, env := range envs {
		deployments := c.flattenEnvDeployments(env)
		for _, deployment := range deployments {
			c.addDeployment(deployment)
		}
	}
	c.lock.Unlock()
}

func (*Collator) flattenEnvDeployments(env CacheEnvironment) []CollationDeploymentInput {
	ret := []CollationDeploymentInput{}
	for _, deployment := range env.Deployments {
		toAppend := CollationDeploymentInput{
			DirectorName:   env.Name,
			DirectorUUID:   env.UUID,
			DeploymentName: deployment.Name,
			Releases:       deployment.Releases.Copy(),
		}
		ret = append(ret, toAppend)
	}

	return ret
}

func (c *Collator) calcDeploymentGroupName(deployment CollationDeploymentInput) (string, bool) {
	var group string
	//Sort it into a deployment group
	for _, rule := range c.rules {
		group = rule.DeploymentGroup(deployment)
		if group != "" {
			break
		}
	}

	return group, group != ""
}

func (c *Collator) addDeployment(deployment CollationDeploymentInput) {
	group, gotGroup := c.calcDeploymentGroupName(deployment)
	if !gotGroup {
		//TODO: Do something with deployments which could not be sorted
		fmt.Printf("Dropping deployment `%s' because no group match\n", deployment.DeploymentName)
		return
	}

	//Remove a possibly stale entry, then add this one
	deploymentID := deployment.calcID()
	//The removal may not be needed now, as we're just wiping the whole cache
	// every time we do a collate?
	c.removeDeployment(deploymentID)
	c.idsToGroups[deploymentID] = group
	groupIdx := c.findGroupIdxByName(group)
	if groupIdx < 0 {
		//If group doesn't exist... make it!
		c.groups = append(c.groups, *newCollationDeploymentGroup(group))
		groupIdx = len(c.groups) - 1
	}
	c.groups[groupIdx].addDeployment(deployment)
	fmt.Printf("Inserted deployment with name `%s' into group `%s'\n",
		deployment.DeploymentName,
		group)
}

func (c *Collator) removeDeployment(deploymentID string) {
	groupName, found := c.idsToGroups[deploymentID]
	if !found {
		return
	}

	groupIdx := c.findGroupIdxByName(groupName)
	if groupIdx < 0 {
		panic("Group id mapping found but no group present. This is a bug.")
	}

	groupIsEmpty := c.groups[groupIdx].removeDeploymentByID(deploymentID)
	if groupIsEmpty {
		lastIdx := len(c.groups) - 1
		c.groups[groupIdx], c.groups[lastIdx] = c.groups[lastIdx], c.groups[groupIdx]
		c.groups = c.groups[:lastIdx]
	}
	delete(c.idsToGroups, deploymentID)
}

func (c *Collator) findGroupIdxByName(name string) int {
	ret := -1
	for i := range c.groups {
		if c.groups[i].Name == name {
			ret = i
			break
		}
	}

	return ret
}
