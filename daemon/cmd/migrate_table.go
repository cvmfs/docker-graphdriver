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
			Id: "wish and converted table",
			Up: []string{
				`CREATE TABLE credential(
					user STRING NOT NULL,
					registry STRING NOT NULL,
					password STRING NOT NULL,
					
					PRIMARY KEY(
						user,
						registry
					),
					CONSTRAINT no_empty_string
						CHECK (
							user != ''
							AND registry != ''
							AND password != ''
						)
				);`,
				`CREATE TABLE image(
					id INTEGER PRIMARY KEY,
					user STRING,
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
					CONSTRAINT thin_images_must_have_an_user
						CHECK (
							is_thin = 0
							OR (
								is_thin != 0 
								AND user NOT NULL
								AND user != ''
							)
						),
					UNIQUE(
						user,
						registry,
						repository,
						tag
					),
					UNIQUE(
						user,
						registry, 
						repository, 
						digest
					),
					FOREIGN KEY (user, registry)
						REFERENCES credential(user, registry)
				);`,
				`CREATE TABLE wish(
					id INTEGER PRIMARY KEY,
					input_image INTEGER NOT NULL,
					output_image INTEGER NOT NULL,
					cvmfs_repo STRING NOT NULL,

					CONSTRAINT unique_wish 
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
					wish INTEGER,
					input_reference STRING NOT NULL,
					CONSTRAINT unique_wish_input
						UNIQUE(
							wish,
							input_reference
					),
					FOREIGN KEY (wish)
						REFERENCES wish(id)
				);`,
			},
			Down: []string{
				`DROP TABLE credential;`,
				`DROP TABLE image;`,
				`DROP TABLE wish;`,
				`DROP TABLE converted;`,
			},
		},
		&migrate.Migration{
			Id: "Add view for image_name",
			Up: []string{
				`
				-- if there is not tag, we print an empty string, 
				-- then we will remove the ":" with the trim function, 
				-- similarly for the digest
				CREATE VIEW image_name(
					image_id,
					name,
					manifest_url
				) AS 
			SELECT 
	id, 
	
	rtrim(
		printf("%s@%s", 
			rtrim(
				printf("%s://%s/%s:%s", scheme, registry, repository, tag),
			':'), 
			digest
			), 
		'@'
	),

	printf("%s://%s/v2/%s/manifests/%s", scheme, registry, repository, 
		COALESCE(tag, printf("@%s", digest)))			

					FROM image
				;`,
			},
			Down: []string{`DROP VIEW image_name`},
		},
	},
}

var migrateDatabaseCmd = &cobra.Command{
	Use:     "migrate-database",
	Aliases: []string{"init", "migrate", "init-db", "migrate-db", "init-database"},
	Short:   "migrate the database to the newest version supported by this version of the software",
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
