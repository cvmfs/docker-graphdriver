package cmd

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	rootCmd.AddCommand(migrateDatabaseCmd)
}

var migrations = &migrate.MemoryMigrationSource{
	Migrations: []*migrate.Migration{
		&migrate.Migration{
			Id: "desiderata and converted table",
			Up: []string{
				`CREATE TABLE image(
					id INTEGER PRIMARY KEY,
					scheme STRING NOT NULL,
					registry STRING NOT NULL,
					repository STRING NOT NULL,
					tag STRING,
					digest STRING,
					is_thin INT NOT NULL,

					CONSTRAINT no_empty_string 
						CHECK (
							scheme != '' AND
							registry != '' AND
							repository != '' AND
							tag != '' AND
							digest != ''
						),
					CONSTRAINT at_least_tag_or_digest
						CHECK (COALESCE(tag, digest) NOT NULL),
					UNIQUE(
						registry,
						repository,
						tag
					),
					UNIQUE(
						registry, 
						repository, 
						digest
					)
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
							cvmfs_repo
						),
					FOREIGN KEY (input_image)
						REFERENCES image(id),
					FOREIGN KEY (output_image)
						REFERENCES image(id)

				);`,
				`CREATE TABLE converted(
					desiderata INTEGER,
					FOREIGN KEY (desiderata)
						REFERENCES desiderata(id)
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

var migrateDatabaseCmd = &cobra.Command{
	Use:   "migrate-database",
	Short: "migrate the database to the newest version supported by this version of the software",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := sql.Open("sqlite3", lib.Database)
		if err != nil {
			lib.LogE(err).Fatal("Impossible to open the database.")
		}
		n, err := migrate.Exec(db, "sqlite3", migrations, migrate.Up)
		if err != nil {
			lib.LogE(err).Fatal("Impossible to migrate the database")
		}
		log.WithFields(log.Fields{"n": n}).Info("Made migrations")
	},
}
