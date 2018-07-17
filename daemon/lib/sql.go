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
SELECT id, user, scheme, registry, repository, tag, digest, is_thin 
FROM image 
WHERE
	registry = :registry 
	AND repository = :repository 
	AND (
		(user = :user)
		OR ("" = :user AND user IS NULL)
	)
	AND (
		(tag = :tag AND digest = :digest) 
		OR (tag IS NULL AND digest = :digest)
		OR (tag = :tag AND digest IS NULL)
		OR ("" = :tag AND digest = :digest)
		OR (tag = :tag AND "" = :digest)
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
		sql.Named("user", image.User),
		sql.Named("tag", image.Tag),
		sql.Named("digest", image.Digest),
	)
	defer rows.Close()
	return rows.Next()
}

func GetImage(queryImage Image) (Image, error) {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getImageStmt, err := db.Prepare(getImageQuery)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	var id int
	var scheme, registry, repository string
	var n_user, n_tag, n_digest sql.NullString
	var is_thin bool
	var user, tag, digest string
	err = getImageStmt.QueryRow(
		sql.Named("registry", queryImage.Registry),
		sql.Named("repository", queryImage.Repository),
		sql.Named("user", queryImage.User),
		sql.Named("tag", queryImage.Tag),
		sql.Named("digest", queryImage.Digest),
	).Scan(&id, &n_user, &scheme, &registry, &repository, &n_tag, &n_digest, &is_thin)
	if err != nil {
		return Image{}, err
	}
	if n_user.Valid {
		user = n_user.String
	} else {
		user = ""
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
	return Image{
		User:       user,
		Scheme:     scheme,
		Registry:   registry,
		Repository: repository,
		Tag:        tag,
		Digest:     digest,
		IsThin:     is_thin}, err

}

func GetImageId(imageQuery Image) (id int, err error) {
	image, err := GetImage(imageQuery)
	if err != nil {
		return
	}
	id = image.Id
	return
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
			LogE(err).Info("Error in getting the images")
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
	VALUES(:scheme, :registry, :repository, :tag, :digest, 0);
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

var getDesiderata = `
SELECT id, input_image, output_image, cvmfs_repo 
FROM desiderata
WHERE (
	input_image = :input
	AND output_image = :output
	AND cvmfs_repo = :repo
);`

func GetDesiderata(input_image, output_image int, cvmfs_repo string) (Desiderata, error) {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getDesiderataStmt, err := db.Prepare(getDesiderata)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	var id, input, output int
	var repo string
	err = getDesiderataStmt.QueryRow(
		sql.Named("input", input_image),
		sql.Named("output", output_image),
		sql.Named("repo", cvmfs_repo),
	).Scan(&id, &input, &output, &repo)
	if err != nil {
		return Desiderata{}, err
	}
	return Desiderata{
		Id:          id,
		InputImage:  input,
		OutputImage: output,
		CvmfsRepo:   repo,
	}, err
}

var getRefreshToken = `
SELECT refresh_token FROM credential WHERE
user = :user AND registry = :registry
`

func GetRefreshToken(user, registry string) (string, error) {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getRefreshTokenStmt, err := db.Prepare(getRefreshToken)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	var n_token sql.NullString
	err = getRefreshTokenStmt.QueryRow(
		sql.Named("user", user),
		sql.Named("registry", registry),
	).Scan(&n_token)
	if err != nil {
		return "", err
	}
	if n_token.Valid {
		return n_token.String, nil
	} else {
		return "", nil
	}
}

var addUser = `
INSERT INTO credential(user, registry, refresh_token)
VALUES(:user, :registry, :refresh_token);
`

func AddUser(user, registry, token string) error {
	var n_token sql.NullString
	if token == "" {
		n_token = sql.NullString{}
	} else {
		n_token = sql.NullString{
			String: token,
			Valid:  true}
	}
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	addUserStmt, err := db.Prepare(addUser)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	_, err = addUserStmt.Exec(
		sql.Named("user", user),
		sql.Named("registry", registry),
		sql.Named("refresh_token", n_token))
	return err
}
