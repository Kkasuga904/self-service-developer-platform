package contract

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var dnsLabel = regexp.MustCompile(`^[a-z0-9](?:[-a-z0-9]*[a-z0-9])?$`)

var approvedOwners = map[string]string{
	"payments-team": "team-payments",
}

var resourceSizes = map[string]struct{}{
	"small": {}, "medium": {}, "large": {},
}

func Validate(service Service) []error {
	var errors []error
	if service.APIVersion != APIVersion {
		errors = append(errors, fmt.Errorf("apiVersion must be %q", APIVersion))
	}
	if service.Kind != Kind {
		errors = append(errors, fmt.Errorf("kind must be %q", Kind))
	}
	if !validDNSLabel(service.Metadata.Name) {
		errors = append(errors, fmt.Errorf("metadata.name must be a valid DNS label of at most 63 characters"))
	}
	if _, ok := approvedOwners[service.Spec.Owner]; !ok {
		errors = append(errors, fmt.Errorf("spec.owner %q is not approved; allowed: %s", service.Spec.Owner, strings.Join(ApprovedOwners(), ", ")))
	}
	if service.Spec.Image == "" {
		errors = append(errors, fmt.Errorf("spec.image is required"))
	} else if !hasPinnedImageReference(service.Spec.Image) {
		errors = append(errors, fmt.Errorf("spec.image must use a non-latest tag or digest"))
	}
	if service.Spec.Port < 1 || service.Spec.Port > 65535 {
		errors = append(errors, fmt.Errorf("spec.port must be between 1 and 65535"))
	}
	if _, ok := resourceSizes[service.Spec.Resources.Size]; !ok {
		errors = append(errors, fmt.Errorf("spec.resources.size must be one of small, medium, large"))
	}
	if service.Spec.Availability.Replicas < 2 || service.Spec.Availability.Replicas > 10 {
		errors = append(errors, fmt.Errorf("spec.availability.replicas must be between 2 and 10"))
	}
	if service.Spec.Autoscaling != nil && service.Spec.Autoscaling.Enabled {
		if service.Spec.Autoscaling.MinReplicas < 2 {
			errors = append(errors, fmt.Errorf("spec.autoscaling.minReplicas must be at least 2"))
		}
		if service.Spec.Autoscaling.MaxReplicas < service.Spec.Autoscaling.MinReplicas || service.Spec.Autoscaling.MaxReplicas > 20 {
			errors = append(errors, fmt.Errorf("spec.autoscaling.maxReplicas must be between minReplicas and 20"))
		}
		if service.Spec.Autoscaling.TargetCPU < 1 || service.Spec.Autoscaling.TargetCPU > 100 {
			errors = append(errors, fmt.Errorf("spec.autoscaling.targetCPUUtilization must be between 1 and 100"))
		}
	}
	return errors
}

func ApprovedOwners() []string {
	owners := make([]string, 0, len(approvedOwners))
	for owner := range approvedOwners {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	return owners
}

func NamespaceForOwner(owner string) (string, bool) {
	namespace, ok := approvedOwners[owner]
	return namespace, ok
}

func validDNSLabel(value string) bool {
	return len(value) > 0 && len(value) <= 63 && dnsLabel.MatchString(value)
}

func hasPinnedImageReference(image string) bool {
	if strings.Contains(image, "@sha256:") {
		return true
	}
	lastSlash := strings.LastIndex(image, "/")
	colon := strings.LastIndex(image, ":")
	if colon <= lastSlash {
		return false
	}
	tag := image[colon+1:]
	return tag != "" && tag != "latest"
}
