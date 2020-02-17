package core

import (
	"regexp"
)

//CollationRule sorts deployments pulled from BOSH directors into deployment
// categories.
type CollationRule interface {
	//DeploymentName should return empty-string if the rule cannot apply to this
	// input
	DeploymentGroup(CollationDeploymentInput) string
}

//DeploymentRegexCaptureRule returns the contents of the first
//capturing group as the deployment name
type DeploymentRegexCaptureRule struct {
	Match *regexp.Regexp
}

func (r DeploymentRegexCaptureRule) DeploymentGroup(in CollationDeploymentInput) string {
	matches := r.Match.FindStringSubmatch(in.DeploymentName)
	if len(matches) < 2 {
		return ""
	}

	return matches[1]
}
