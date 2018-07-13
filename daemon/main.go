package main

import (
	"database/sql"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rubenv/sql-migrate"

	"github.com/cvmfs/docker-graphdriver/daemon/cmd"
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
	cmd.EntryPoint()
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
