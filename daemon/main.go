package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rubenv/sql-migrate"
)

var migrations = &migrate.MemoryMigrationSource{
	Migrations: []*migrate.Migration{
		&migrate.Migration{
			Id: "desiderata and converted table",
			Up: []string{
				`CREATE TABLE image(
					id INTEGER PRIMARY KEY,
					protocol STRING NOT NULL,
					registry STRING NOT NULL,
					repository STRING NOT NULL,
					tag_input STRING,
					digest STRING,
					is_thin INT NOT NULL,
				);`,
				`CREATE TABLE desiderata(
					id INTEGER PRIMARY KEY,
					input_image INTEGER,
					output_image INTEGER,
					cvmfs_repo STRING NOT NULL,

					CONSTRAINT unique_desiderata 
						UNIQUE(
							input_image,
							output_image,
							cvmfs_repo,
						),
					FOREIGN KEY (input_image)
						REFERENCE image(id),
					FOREIGN KEY (output_image)
						REFERENCE image(id),

				);`,
				`CREATE TABLE converted(
					desiderata INTEGER,
					FOREIGN KEY (desiderata)
						REFERENCE desiderata(id)
				);`,
			},
			Down: []string{
				`DROP TABLE image`,
				`DROP TABLE desiderata;`,
				`DROP TABLE converted`,
			},
		},
	},
}

var InsertDesiderataQuery = `INSERT INTO desiderata(
				repository_input, 
				tag_input, 
				cvmfs_repo, 
				repository_output, 
				tag_output,
				tag_is_digest) VALUES(?, ?, ?, ?, ?, ?);`

var (
	InsertDesiderata *sql.Stmt
)

/*

Three simple commands:
	1. add desiderata
	2. refresh
	3. check

add desiderata will add a new row to the desiderata table and will try to convert it.
refresh will check all the converted and be sure that they are up to date, maybe we changed a tag
check will simply make sure that every desiderata is been converted

We will talk with the docker daemon and use it for most of the work

*/

var (
	username_input, password_input,
	input_repository,

	username_output, password_output,
	output_repository,
	cvmfs_repository string

	machineFriendly bool
)

func main() {
	rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Show the several commands available.",
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

var addCmd = &cobra.Command{
	Use:   "add-desiderata",
	Short: "Add a new image to convert into thin images",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			inputSpecifier, outputSpecifier string
			specifier_is_digest             bool
		)
		flags := cmd.Flags()
		inputReference := flags.Lookup("input-repository").Value.String()
		outputReference := flags.Lookup("output-repository").Value.String()
		cvmfsRepo := flags.Lookup("repository").Value.String()

		inputRegistry, inputTag, inputDigest, err := SplitReference(inputReference)
		if err != nil {
			logE(err).Fatal("Make sure that you specify either a tag repository:tag or a digest repository@digest for the input image")
		}
		if inputTag != "" {
			inputSpecifier = inputTag
			specifier_is_digest = false
		} else {
			inputSpecifier = inputDigest
			specifier_is_digest = true
		}

		outputRegistry, outputTag, _, err := SplitReference(outputReference)
		if err != nil || outputTag == "" {
			logE(err).Fatal("Make sure that you specify a tag repository:tag for the output image")
		}
		outputSpecifier = outputTag

		_, err = InsertDesiderata.Exec(inputRegistry, inputSpecifier, cvmfsRepo, outputRegistry, outputSpecifier, specifier_is_digest)
		if err != nil {
			logE(err).Fatal("Error in creating your desiderata, maybe already exists?")
		}

	},
}

var checkImageSyntaxCmd = &cobra.Command{
	Use:   "check-image-syntax",
	Short: "Check that the provide image has a valid syntax, the same checks are applied before any command in the converter.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		img, err := ParseImage(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if machineFriendly {
			fmt.Printf("scheme,registry,repository,tag,digest\n")
			fmt.Printf("%s,%s,%s,%s,%s\n", img.Scheme, img.Registry, img.Repository, img.Tag, img.Digest)
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetHeader([]string{"Key", "Value"})
			table.Append([]string{"Scheme", img.Scheme})
			table.Append([]string{"Registry", img.Registry})
			table.Append([]string{"Repository", img.Repository})
			table.Append([]string{"Tag", img.Tag})
			table.Append([]string{"Digest", img.Digest})
			table.Render()
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&username_input, "username-input",
		"", "username to log into the input register")
	rootCmd.PersistentFlags().StringVar(&password_input, "password-input",
		"", "password to log into the input register")
	rootCmd.PersistentFlags().StringVar(&username_output, "username-output",
		"", "username to log into the output register")
	rootCmd.PersistentFlags().StringVar(&password_output, "password-output",
		"", "password to log into the output register")

	addCmd.Flags().StringVarP(&input_repository,
		"input-repository", "i", "", "The repository to convert")
	addCmd.Flags().StringVarP(&output_repository,
		"output-repository", "o", "", "The repository of the converted image")
	addCmd.Flags().StringVarP(&cvmfs_repository,
		"repository", "r", "", "The cvmfs repository where to store the images")
	addCmd.MarkFlagRequired("input-repository")
	addCmd.MarkFlagRequired("output-repository")
	addCmd.MarkFlagRequired("repository")

	checkImageSyntaxCmd.Flags().BoolVarP(&machineFriendly, "machine-friendly", "z", false, "produce machine friendly output, one line of csv")

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(checkImageSyntaxCmd)

	createDb()
}

func logE(err error) *log.Entry {
	return log.WithFields(log.Fields{"error": err})
}

func createDb() *sql.DB {
	db, err := sql.Open("sqlite3", "docker2cvmfs_archive.sqlite")
	if err != nil {
		logE(err).Fatal("Impossible to open the database.")
	}
	n, err := migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		logE(err).Fatal("Impossible to migrate the database")
	}
	log.WithFields(log.Fields{"n": n}).Info("Made migrations")
	InsertDesiderata, err = db.Prepare(InsertDesiderataQuery)
	return db
}

func run() {

}

func SplitReference(repository string) (register, tag, digest string, err error) {
	tryTag := strings.Split(repository, ":")
	if len(tryTag) == 2 {
		return tryTag[0], tryTag[1], "", nil
	}
	tryDigest := strings.Split(repository, "@")
	if len(tryDigest) == 2 {
		return tryTag[0], "", tryTag[1], nil
	}
	return "", "", "", fmt.Errorf("Impossible to split the repository into repository and tag")
}

type Image struct {
	Scheme     string
	Registry   string
	Repository string
	Tag        string
	Digest     string
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
