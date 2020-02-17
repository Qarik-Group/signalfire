package core

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type CollationDeploymentGroup struct {
	Name        string
	Deployments []CollationDeployment
	Releases    []CollationRelease
}

func newCollationDeploymentGroup(name string) *CollationDeploymentGroup {
	return &CollationDeploymentGroup{Name: name}
}

func (c *CollationDeploymentGroup) addDeployment(in CollationDeploymentInput) {
	dep := CollationDeployment{
		ID:           in.calcID(),
		Name:         in.DeploymentName,
		DirectorUUID: in.DirectorUUID,
	}
	c.Deployments = append(c.Deployments, dep)
	for _, release := range in.Releases {
		c.addRelease(dep.ID, release)
	}
}

func (c *CollationDeploymentGroup) removeDeploymentByID(id string) bool {
	for i := range c.Deployments {
		if c.Deployments[i].ID == id {
			lastIdx := len(c.Deployments) - 1
			c.Deployments[i], c.Deployments[lastIdx] = c.Deployments[lastIdx], c.Deployments[i]
			c.Deployments = c.Deployments[:lastIdx]
			break
		}
	}

	if len(c.Deployments) == 0 {
		c.Releases = nil
		return true
	}

	newReleaseList := []CollationRelease{}
	for i := range c.Releases {
		noRemainingVersions := c.Releases[i].removeDeploymentByID(id)
		if !noRemainingVersions {
			newReleaseList = append(newReleaseList, c.Releases[i])
		}
	}
	c.Releases = newReleaseList
	return false
}

func (c *CollationDeploymentGroup) findReleaseIdxByName(name string) int {
	ret := -1
	for i := range c.Releases {
		if c.Releases[i].Name == name {
			ret = i
			break
		}
	}

	return ret
}

func (c *CollationDeploymentGroup) addRelease(deploymentID string, release CacheRelease) {
	relIdx := c.findReleaseIdxByName(release.Name)
	if relIdx < 0 {
		c.Releases = append(c.Releases, CollationRelease{
			Name: release.Name,
		})
		relIdx = len(c.Releases) - 1
	}
	c.Releases[relIdx].addDeploymentVersion(deploymentID, release.Version)
}

type CollationDeployment struct {
	ID           string
	Name         string
	DirectorUUID string
}

type CollationRelease struct {
	Name     string
	Versions []CollationReleaseVersion
}

func (r *CollationRelease) findReleaseVersionIdx(version string) int {
	for i := range r.Versions {
		if r.Versions[i].Version == version {
			return i
		}
	}
	return -1
}

func (r *CollationRelease) addDeploymentVersion(id, version string) {
	vIdx := r.findReleaseVersionIdx(version)
	if vIdx < 0 {
		r.Versions = append(r.Versions, CollationReleaseVersion{Version: version})
		vIdx = len(r.Versions) - 1
	}

	r.Versions[vIdx].addDeployment(id)
	r.sortVersions()
}

//removeDeploymentByID returns true if the release is empty of versions
// after the deletion
func (r *CollationRelease) removeDeploymentByID(id string) bool {
	for i := range r.Versions {
		if r.Versions[i].findDeploymentIdx(id) < 0 {
			continue
		}

		noRemainingDeploymentsForVersion := r.Versions[i].removeDeployment(id)
		if noRemainingDeploymentsForVersion {
			lastIdx := len(r.Versions) - 1
			r.Versions[i], r.Versions[lastIdx] = r.Versions[lastIdx], r.Versions[i]
			r.Versions = r.Versions[:lastIdx]
		}
		break
	}

	r.sortVersions()
	return len(r.Versions) == 0
}

func (r *CollationRelease) sortVersions() {
	sort.Slice(r.Versions, func(i, j int) bool {
		return r.Versions[i].LessThan(r.Versions[j])
	})
}

type CollationReleaseVersion struct {
	Version string
	//Deployments is a list of deployment IDs using this
	// release version
	Deployments []string
}

func (v *CollationReleaseVersion) addDeployment(id string) {
	v.Deployments = append(v.Deployments, id)
}

//removeDeployment returns true if the version is empty of deployments after
// the deletion
func (v *CollationReleaseVersion) removeDeployment(id string) bool {
	dIdx := v.findDeploymentIdx(id)
	if dIdx >= 0 {
		lastIdx := len(v.Deployments) - 1
		v.Deployments[dIdx], v.Deployments[lastIdx] = v.Deployments[lastIdx], v.Deployments[dIdx]
		v.Deployments = v.Deployments[:lastIdx]
	}
	return len(v.Deployments) == 0
}

func (v *CollationReleaseVersion) findDeploymentIdx(id string) int {
	for i := range v.Deployments {
		if v.Deployments[i] == id {
			return i
		}
	}
	return -1
}

func (v1 CollationReleaseVersion) LessThan(v2 CollationReleaseVersion) bool {
	n1, rc1 := v1.parseVersionAndRC()
	n2, rc2 := v2.parseVersionAndRC()
	vDiff := v1.versionNumDiff(n1, n2)
	if vDiff == 0 {
		//A non-rc is greater than an rc
		if len(rc1) == 0 {
			return false
		}
		if len(rc2) == 0 {
			return true
		}
		vDiff = v1.versionNumDiff(rc1, rc2)
	}
	return vDiff < 0
}

func (v CollationReleaseVersion) parseVersionAndRC() (version []int64, rc []int64) {
	matches := regexp.MustCompile("rc[^0-9]*(.*)$").FindStringSubmatch(v.Version)
	if len(matches) > 0 {
		v.Version = strings.Replace(v.Version, matches[0], "", -1)
		rc = []int64{0}
		if len(matches) > 1 {
			rc = v.parseVersionString(matches[1])
		}
	}
	version = v.parseVersionString(v.Version)
	return
}

func (CollationReleaseVersion) parseVersionString(version string) []int64 {
	numberedComponents := regexp.MustCompile("[^0-9]+").Split(version, -1)
	if len(numberedComponents) == 0 {
		return []int64{0}
	}
	ret := make([]int64, len(numberedComponents))
	for i, numStr := range numberedComponents {
		ret[i], _ = strconv.ParseInt(numStr, 10, 64)
	}

	return ret
}

// versionNumDiff returns
// < 0 if v1 < v2
// 0   if v1 == v2
// > 0 if v1 > v2
func (CollationReleaseVersion) versionNumDiff(v1, v2 []int64) int64 {
	maxLen := len(v1)
	if len(v2) > maxLen {
		maxLen = len(v2)
	}

	for i := 0; i < maxLen; i++ {
		//Default the point to 0, in the case that both versions don't have the
		// same number of points
		var v1Val, v2Val int64 = 0, 0
		if len(v1) > i {
			v1Val = v1[i]
		}
		if len(v2) > i {
			v2Val = v2[i]
		}
		if v1Val != v2Val {
			return v1Val - v2Val
		}
	}

	return 0
}
