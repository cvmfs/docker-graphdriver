package lib

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type Image struct {
	Id         int
	User       string
	Scheme     string
	Registry   string
	Repository string
	Tag        string
	Digest     string
	IsThin     bool
}

type Desiderata struct {
	Id          int
	InputImage  int
	OutputImage int
	CvmfsRepo   string
}

func (i Image) WholeName() string {
	root := fmt.Sprintf("%s://%s/%s", i.Scheme, i.Registry, i.Repository)
	if i.Tag != "" {
		root = fmt.Sprintf("%s:%s", root, i.Tag)
	}
	if i.Digest != "" {
		root = fmt.Sprintf("%s@%s", root, i.Digest)
	}
	return root
}

func (i Image) GetManifestUrl() string {
	url := fmt.Sprintf("%s://%s/v2/%s/manifests/", i.Scheme, i.Registry, i.Repository)
	if i.Digest != "" {
		url = fmt.Sprintf("%s@%s", url, i.Digest)
	} else {
		url = fmt.Sprintf("%s%s", url, i.Tag)
	}
	return url
}

func (img Image) PrintImage(machineFriendly, csv_header bool) {
	if machineFriendly {
		if csv_header {
			fmt.Printf("name,scheme,registry,repository,tag,digest,is_thin\n")
		}
		fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s\n",
			img.WholeName(), img.User, img.Scheme,
			img.Registry, img.Repository,
			img.Tag, img.Digest,
			fmt.Sprint(img.IsThin))
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeader([]string{"Key", "Value"})
		table.Append([]string{"Name", img.WholeName()})
		table.Append([]string{"User", img.User})
		table.Append([]string{"Scheme", img.Scheme})
		table.Append([]string{"Registry", img.Registry})
		table.Append([]string{"Repository", img.Repository})
		table.Append([]string{"Tag", img.Tag})
		table.Append([]string{"Digest", img.Digest})
		table.Render()
	}
}

func ParseImage(image string) (img Image, err error) {
	url, err := url.Parse(image)
	if err != nil {
		return Image{}, err
	}
	if url.Host == "" {
		return Image{}, fmt.Errorf("Impossible to identify the registry of the image: %s", image)
	}
	if url.Path == "" {
		return Image{}, fmt.Errorf("Impossible to identify the repository of the image: %s", image)
	}
	colonPathSplitted := strings.Split(url.Path, ":")
	if len(colonPathSplitted) == 0 {
		return Image{}, fmt.Errorf("Impossible to identify the path of the image: %s", image)
	}
	// no split happened, hence we don't have neither a tag nor a digest, but only a path
	if len(colonPathSplitted) == 1 {

		// we remove the first  and the trailing `/`
		repository := strings.TrimLeft(colonPathSplitted[0], "/")
		repository = strings.TrimRight(repository, "/")
		if repository == "" {
			return Image{}, fmt.Errorf("Impossible to find the repository for: %s", image)
		}
		return Image{
			Scheme:     url.Scheme,
			Registry:   url.Host,
			Repository: repository,
		}, nil

	}
	if len(colonPathSplitted) > 3 {
		fmt.Println(colonPathSplitted)
		return Image{}, fmt.Errorf("Impossible to parse the string into an image, too many `:` in : %s", image)
	}
	// the colon `:` is used also as separator in the digest between sha256
	// and the actuall digest, a len(pathSplitted) == 2 could either means
	// a repository and a tag or a repository and an hash, in the case of
	// the hash however the split will be more complex.  Now we split for
	// the at `@` which separate the digest from everything else. If this
	// split produce only one result we have a repository and maybe a tag,
	// if it produce two we have a repository, maybe a tag and definitely a
	// digest, if it produce more than two we have an error.
	atPathSplitted := strings.Split(url.Path, "@")
	if len(atPathSplitted) > 2 {
		return Image{}, fmt.Errorf("To many `@` in the image name: %s", image)
	}
	var repoTag, digest string
	if len(atPathSplitted) == 2 {
		digest = atPathSplitted[1]
		repoTag = atPathSplitted[0]
	}
	if len(atPathSplitted) == 1 {
		repoTag = atPathSplitted[0]
	}
	// finally we break up also the repoTag to find out if we have also a
	// tag or just a repository name
	colonRepoTagSplitted := strings.Split(repoTag, ":")

	// only the repository, without the tag
	if len(colonRepoTagSplitted) == 1 {
		repository := strings.TrimLeft(colonRepoTagSplitted[0], "/")
		repository = strings.TrimRight(repository, "/")
		if repository == "" {
			return Image{}, fmt.Errorf("Impossible to find the repository for: %s", image)
		}
		return Image{
			Scheme:     url.Scheme,
			Registry:   url.Host,
			Repository: repository,
			Digest:     digest,
		}, nil
	}

	// both repository and tag
	if len(colonRepoTagSplitted) == 2 {
		repository := strings.TrimLeft(colonRepoTagSplitted[0], "/")
		repository = strings.TrimRight(repository, "/")
		if repository == "" {
			return Image{}, fmt.Errorf("Impossible to find the repository for: %s", image)
		}
		return Image{
			Scheme:     url.Scheme,
			Registry:   url.Host,
			Repository: repository,
			Tag:        colonRepoTagSplitted[1],
			Digest:     digest,
		}, nil
	}
	return Image{}, fmt.Errorf("Impossible to parse the image: %s", image)
}
