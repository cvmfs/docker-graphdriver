package lib

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var Database = "/var/lib/docker2cvmfs/docker2cvmfs_archive.sqlite?_foreign_keys=true"

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
		Id:         id,
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

var getImageById = `
SELECT id, user, scheme, registry, repository, tag, digest, is_thin 
FROM image 
WHERE id = :id;
`

func GetImageById(inputId int) (Image, error) {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getImageByIdStmt, err := db.Prepare(getImageById)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	var id int
	var scheme, registry, repository string
	var n_user, n_tag, n_digest sql.NullString
	var is_thin bool
	var user, tag, digest string
	err = getImageByIdStmt.QueryRow(
		sql.Named("id", inputId),
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
		Id:         id,
		User:       user,
		Scheme:     scheme,
		Registry:   registry,
		Repository: repository,
		Tag:        tag,
		Digest:     digest,
		IsThin:     is_thin}, err
}

var getAllImages = `SELECT rowid, user, scheme, registry, repository, tag, digest, is_thin FROM image`

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
		var id int
		var user, scheme, registry, repository, tag, digest string
		var is_thin bool

		var n_id sql.NullInt64
		var n_user, n_scheme, n_registry, n_repository, n_tag, n_digest sql.NullString
		var n_is_thin sql.NullBool
		err = rows.Scan(&n_id, &n_user, &n_scheme,
			&n_registry, &n_repository,
			&n_tag, &n_digest, &n_is_thin)
		if err != nil {
			LogE(err).Info("Error in getting the images")
			return []Image{}, err
		}

		if n_id.Valid {
			id = int(n_id.Int64)
		} else {
			id = 0
		}
		if n_user.Valid {
			user = n_user.String
		} else {
			user = ""
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
			Id:         id,
			User:       user,
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
INSERT INTO image(scheme, user, registry, repository, tag, digest, is_thin) 
VALUES(:scheme, :user, :registry, :repository, :tag, :digest, :is_thin);
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
	var n_user sql.NullString
	if img.User != "" {
		n_user.Valid = true
		n_user.String = img.User
	}
	var is_thin int
	if img.IsThin {
		is_thin = 1
	} else {
		is_thin = 0
	}
	args := []interface{}{
		sql.Named("user", n_user),
		sql.Named("scheme", img.Scheme),
		sql.Named("registry", img.Registry),
		sql.Named("repository", img.Repository),
		sql.Named("is_thin", is_thin),
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

var addDesiderata = `
INSERT INTO desiderata(input_image, output_image, cvmfs_repo)
VALUES(:input, :output, :repo);
`

func AddDesiderata(inputId, outputId int, repo string) (err error) {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		return
	}
	addDesiderataStmt, err := db.Prepare(addDesiderata)
	if err != nil {
		return
	}
	_, err = addDesiderataStmt.Exec(
		sql.Named("input", inputId),
		sql.Named("output", outputId),
		sql.Named("repo", repo))
	return
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

var getPassword = `
SELECT password FROM credential WHERE
user = :user AND registry = :registry
`

func GetPassword(user, registry string) (string, error) {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getRefreshTokenStmt, err := db.Prepare(getPassword)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	var password string
	err = getRefreshTokenStmt.QueryRow(
		sql.Named("user", user),
		sql.Named("registry", registry),
	).Scan(&password)
	if err != nil {
		return "", err
	}
	return password, nil
}

var addUser = `
INSERT INTO credential(user, registry, password)
VALUES(:user, :registry, :password);
`

func AddUser(user, password, registry string) error {
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
		sql.Named("password", password),
		sql.Named("registry", registry))
	return err
}

var getAllUsers = `SELECT user, registry FROM credential;`

func GetAllUsers() ([]struct{ Username, Registry string }, error) {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getAllUserStmt, err := db.Prepare(getAllUsers)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	rows, err := getAllUserStmt.Query()
	users := []struct{ Username, Registry string }{}
	defer rows.Close()
	if err != nil {
		return users, err
	}
	for rows.Next() {
		var user, registry string
		err = rows.Scan(&user, &registry)
		if err != nil {
			return users, err
		}
		users = append(users, struct{ Username, Registry string }{Username: user, Registry: registry})
	}
	return users, nil
}

var getUserPassword = `
SELECT password FROM credential 
WHERE user = :user 
AND registry = :registry;`

func GetUserPassword(user, registry string) (password string, err error) {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getUserPasswordStmt, err := db.Prepare(getUserPassword)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	err = getUserPasswordStmt.QueryRow(
		sql.Named("user", user),
		sql.Named("registry", registry),
	).Scan(&password)
	return
}

var getAllDesiderata = `
SELECT d.id, d.input_image, input.name, d.output_image, output.name, d.cvmfs_repo
	FROM desiderata AS d
	JOIN image_name as input
	JOIN image_name as output
	WHERE 
		d.input_image = input.image_id
		AND d.output_image = output.image_id;
`

func GetAllDesiderata() ([]DesiderataFriendly, error) {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getAllDesiderataStmt, err := db.Prepare(getAllDesiderata)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	rows, err := getAllDesiderataStmt.Query()
	desiderata := []DesiderataFriendly{}
	defer rows.Close()
	if err != nil {
		return desiderata, err
	}
	for rows.Next() {
		var id, input_id, output_id int
		var input_name, output_name, cvmfsRepo string
		err = rows.Scan(&id, &input_id, &input_name, &output_id, &output_name, &cvmfsRepo)
		if err != nil {
			return desiderata, err
		}
		desi := DesiderataFriendly{
			Id:         id,
			InputId:    input_id,
			InputName:  input_name,
			OutputId:   output_id,
			OutputName: output_name,
			CvmfsRepo:  cvmfsRepo,
		}
		desiderata = append(desiderata, desi)
	}
	return desiderata, nil
}

var getDesiderataF = `
SELECT d.id, d.input_image, input.name, d.output_image, output.name, d.cvmfs_repo
	FROM desiderata AS d
	JOIN image_name as input
	JOIN image_name as output
	WHERE 
		d.input_image = input.image_id
		AND d.output_image = output.image_id
		AND input.image_id = :input_id
		AND output.image_id = :output_id
		AND d.cvmfs_repo = :repo;
`

func GetDesiderataF(inputId, outputId int, repo string) (desi DesiderataFriendly, err error) {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getDesiderataFStmt, err := db.Prepare(getDesiderataF)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	var id, input_id, output_id int
	var input_name, output_name, cvmfsRepo string
	err = getDesiderataFStmt.QueryRow(
		sql.Named("input_id", inputId),
		sql.Named("output_id", outputId),
		sql.Named("repo", repo),
	).Scan(&id, &input_id, &input_name, &output_id, &output_name, &cvmfsRepo)
	if err != nil {
		return
	}
	desi.Id = id
	desi.InputId = input_id
	desi.InputName = input_name
	desi.OutputId = output_id
	desi.OutputName = output_name
	desi.CvmfsRepo = cvmfsRepo
	return
}

var addConverted = `INSERT INTO converted VALUES(:desiderata, :input_reference);`

func AddConverted(desiderataId int, inputReferece string) error {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	AddConvertedStmt, err := db.Prepare(addConverted)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	_, err = AddConvertedStmt.Exec(
		sql.Named("desiderata", desiderataId),
		sql.Named("input_reference", inputReferece),
	)
	return err
}

var alreadyConverted = `
SELECT desiderata, input_reference FROM converted WHERE
desiderata = :desiderata_id
AND input_reference = :input_reference
`

func AlreadyConverted(desiderataId int, input_reference string) bool {
	db, err := sql.Open("sqlite3", Database)
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	alreadyConvertedStmt, err := db.Prepare(alreadyConverted)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	var id int
	var inputReference string
	err = alreadyConvertedStmt.QueryRow(
		sql.Named("desiderata_id", desiderataId),
		sql.Named("input_reference", input_reference),
	).Scan(&id, &inputReference)
	return err == nil
}
