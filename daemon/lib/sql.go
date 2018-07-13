package lib

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var Database = "docker2cvmfs_archive.sqlite"

var (
	BB           *sql.DB
	GetImageStmt *sql.Stmt
)

var getImageQuery = `
SELECT * FROM image WHERE
	registry = :registry AND
	repository = :repository AND
	(
		(tag = :tag AND digest = :digest) OR
		(tag IS NULL AND digest = :digest) OR
		(tag = :tag AND digest IS NULL)
	);
`

func IsImageInDatabase(image Image) bool {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getImageStmt, err := db.Prepare(getImageQuery)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	rows, err := getImageStmt.Query(
		sql.Named("registry", image.Registry),
		sql.Named("repository", image.Repository),
		sql.Named("tag", image.Tag),
		sql.Named("digest", image.Digest),
	)
	defer rows.Close()
	return rows.Next()
}

var getAllImages = `SELECT scheme, registry, repository, tag, digest, is_thin FROM image`

func GetAllImagesInDatabase() ([]Image, error) {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getAllImagesStmt, err := db.Prepare(getAllImages)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	rows, err := getAllImagesStmt.Query()
	defer rows.Close()
	if err != nil {
		return []Image{}, err
	}
	imgs := []Image{}
	for rows.Next() {
		var scheme, registry, repository, tag, digest string
		var is_thin bool

		var n_scheme, n_registry, n_repository, n_tag, n_digest sql.NullString
		var n_is_thin sql.NullBool
		err = rows.Scan(&n_scheme, &n_registry, &n_repository, &n_tag, &n_digest, &n_is_thin)
		if err != nil {
			return []Image{}, err
		}

		if n_scheme.Valid {
			scheme = n_scheme.String
		} else {
			scheme = ""
		}
		if n_registry.Valid {
			registry = n_registry.String
		} else {
			registry = ""
		}
		if n_repository.Valid {
			repository = n_repository.String
		} else {
			repository = ""
		}
		if n_tag.Valid {
			tag = n_tag.String
		} else {
			tag = ""
		}
		if n_digest.Valid {
			digest = n_digest.String
		} else {
			digest = ""
		}
		if n_is_thin.Valid {
			is_thin = n_is_thin.Bool
		} else {
			is_thin = false
		}

		imgs = append(imgs, Image{
			Scheme:     scheme,
			Registry:   registry,
			Repository: repository,
			Tag:        tag,
			Digest:     digest,
			IsThin:     is_thin,
		})
	}
	if err != nil {
		return []Image{}, err
	}
	return imgs, nil
}

var addImage = `
INSERT INTO image(scheme, registry, repository, tag, digest, is_thin) 
	VALUES(:scheme, :registry, :repository, :tag, :digest, 0)
`

func AddImage(img Image) error {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	addImagesStmt, err := db.Prepare(addImage)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	args := []interface{}{
		sql.Named("scheme", img.Scheme),
		sql.Named("registry", img.Registry),
		sql.Named("repository", img.Repository),
	}
	if img.Tag != "" {
		args = append(args, sql.Named("tag", img.Tag))
	} else {
		args = append(args, sql.Named("tag", nil))
	}
	if img.Digest != "" {
		args = append(args, sql.Named("digest", img.Digest))
	} else {
		args = append(args, sql.Named("digest", nil))
	}
	_, err = addImagesStmt.Exec(args...)
	return err
}
