package bcc

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type Artifact struct {
	// A map of regions to baiducloud image ids, which created by the build process
	BaiduCloudImages map[string]string

	// BuilderId is the unique ID for the builder that created this alicloud image
	BuilderIdValue string

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}

	Client *bcc.Client
}

func (a *Artifact) BuilderId() string {
	return a.BuilderIdValue
}

func (a *Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	parts := make([]string, 0, len(a.BaiduCloudImages))
	for region, bccImageId := range a.BaiduCloudImages {
		parts = append(parts, fmt.Sprintf("%s:%s", region, bccImageId))
	}

	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func (a *Artifact) String() string {
	parts := make([]string, 0, len(a.BaiduCloudImages))
	for region, bccImageId := range a.BaiduCloudImages {
		parts = append(parts, fmt.Sprintf("%s: %s", region, bccImageId))
	}
	sort.Strings(parts)

	return fmt.Sprintf("Baiducloud images(%s) were created.\n\n", strings.Join(parts, "\n"))
}

func (a *Artifact) State(name string) interface{} {
	if _, ok := a.StateData[name]; ok {
		return a.StateData[name]
	}

	switch name {
	case "atlas.artifact.metadata":
		return a.stateAtlasMetadata()
	default:
		return nil
	}
}

func (a *Artifact) Destroy() error {
	errs := make([]error, 0)

	// if builder mean to destroy the image, usually no need to copy image to remote or
	// share the image to other users. So, in this place, we assume that the client can
	// get image detail and delete the image.
	for region, imageId := range a.BaiduCloudImages {
		log.Printf("Deleting baiducloud image(%s) from region(%s)", imageId, region)

		_, err := a.Client.GetImageDetail(imageId)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		shareList, err := a.Client.GetImageSharedUser(imageId)
		for _, shareUser := range shareList.Users {
			if err := a.Client.UnShareImage(imageId, &shareUser); err != nil {
				errs = append(errs, err)
			}
		}

		if err := a.Client.DeleteImage(imageId); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return &packersdk.MultiError{Errors: errs}
	}
	return nil
}

func (a *Artifact) stateAtlasMetadata() interface{} {
	metadata := make(map[string]string)
	for region, imageId := range a.BaiduCloudImages {
		k := fmt.Sprintf("region.%s", region)
		metadata[k] = imageId
	}

	return metadata
}
