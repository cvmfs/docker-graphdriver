package lib

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os/user"
	"path"

	da "github.com/cvmfs/docker-graphdriver/daemon/docker-api"

	_ "github.com/mattn/go-sqlite3"
)

// var DefaultDatabaseLocation = "/var/lib/docker2cvmfs/docker2cvmfs_archive.sqlite"
var DefaultDatabaseLocation string

func init() {
	usr, err := user.Current()
	if err != nil {
		DefaultDatabaseLocation = path.Join("~", ".docker2cvmfs", "docker2cvmfs_archive.sqlite")
	}
	DefaultDatabaseLocation = path.Join(usr.HomeDir, ".docker2cvmfs", "docker2cvmfs_archive.sqlite")
}

var DatabaseLocation string

var DatabasePostFix = "?_foreign_keys=true"

func DatabaseFile() string {
	return DatabaseLocation
}

func Database() string {
	return DatabaseFile() + DatabasePostFix
}

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

func GetImage(queryImage Image) (Image, error) {
	db, err := sql.Open("sqlite3", Database())
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	return GetImageWithDb(db, queryImage)
}

func GetImageWithDb(db *sql.DB, queryImage Image) (Image, error) {
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
	db, err := sql.Open("sqlite3", Database())
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
	db, err := sql.Open("sqlite3", Database())
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
	db, err := sql.Open("sqlite3", Database())
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	return AddImageWithConnection(db, img)
}

func AddImageWithConnection(db *sql.DB, img Image) error {
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

var addWish = `
INSERT INTO wish(input_image, output_image, cvmfs_repo)
VALUES(:input, :output, :repo);
`

func AddWish(inputId, outputId int, repo string) (err error) {
	db, err := sql.Open("sqlite3", Database())
	if err != nil {
		return
	}
	addWishStmt, err := db.Prepare(addWish)
	if err != nil {
		return
	}
	_, err = addWishStmt.Exec(
		sql.Named("input", inputId),
		sql.Named("output", outputId),
		sql.Named("repo", repo))
	return
}

var getWish = `
SELECT id, input_image, output_image, cvmfs_repo 
FROM wish
WHERE (
	input_image = :input
	AND output_image = :output
	AND cvmfs_repo = :repo
);`

func GetWish(input_image, output_image int, cvmfs_repo string) (Wish, error) {
	db, err := sql.Open("sqlite3", Database())
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getWishStmt, err := db.Prepare(getWish)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	var id, input, output int
	var repo string
	err = getWishStmt.QueryRow(
		sql.Named("input", input_image),
		sql.Named("output", output_image),
		sql.Named("repo", cvmfs_repo),
	).Scan(&id, &input, &output, &repo)
	if err != nil {
		return Wish{}, err
	}
	return Wish{
		Id:          id,
		InputImage:  input,
		OutputImage: output,
		CvmfsRepo:   repo,
	}, err
}

var getAllRawWishes = `
SELECT id, input_image, output_image, cvmfs_repo FROM wish;
`

func GetAllWish() ([]Wish, error) {
	db, err := sql.Open("sqlite3", Database())
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	stmt, err := db.Prepare(getAllRawWishes)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	wishes := []Wish{}
	rows, err := stmt.Query()
	if err != nil {
		return wishes, err
	}
	for rows.Next() {
		var id, input, output int
		var repo string
		err = rows.Scan(&id, &input, &output, &repo)
		if err != nil {
			return wishes, err
		}
		wish := Wish{
			Id:          id,
			InputImage:  input,
			OutputImage: output,
			CvmfsRepo:   repo,
		}
		wishes = append(wishes, wish)
	}
	return wishes, nil

}

var getPassword = `
SELECT password FROM credential WHERE
user = :user AND registry = :registry
`

func GetPassword(user, registry string) (string, error) {
	db, err := sql.Open("sqlite3", Database())
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

var getAllUsers = `SELECT user, registry FROM credential;`

func GetAllUsers() ([]struct{ Username, Registry string }, error) {
	db, err := sql.Open("sqlite3", Database())
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
	db, err := sql.Open("sqlite3", Database())
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

var getAllWishes = `
SELECT 
	d.id, d.input_image, 
	input.name, d.output_image, 
	output.name, d.cvmfs_repo, 
	CASE WHEN
		(SELECT input_reference FROM converted WHERE d.id = wish) IS NULL
		THEN 0
		ELSE 1
	END

FROM wish AS d
	JOIN image_name as input
	JOIN image_name as output

WHERE 
	d.input_image = input.image_id
	AND d.output_image = output.image_id;
`

func GetAllWishes() ([]WishFriendly, error) {
	db, err := sql.Open("sqlite3", Database())
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getAllWishesStmt, err := db.Prepare(getAllWishes)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	rows, err := getAllWishesStmt.Query()
	wishes := []WishFriendly{}
	defer rows.Close()
	if err != nil {
		return wishes, err
	}
	for rows.Next() {
		var id, input_id, output_id int
		var input_name, output_name, cvmfsRepo string
		var converted bool
		err = rows.Scan(&id, &input_id, &input_name, &output_id, &output_name, &cvmfsRepo, &converted)
		if err != nil {
			return wishes, err
		}
		wish := WishFriendly{
			Id:         id,
			InputId:    input_id,
			InputName:  input_name,
			OutputId:   output_id,
			OutputName: output_name,
			CvmfsRepo:  cvmfsRepo,
			Converted:  converted,
		}
		wishes = append(wishes, wish)
	}
	return wishes, nil
}

var getWishF = `
SELECT d.id, d.input_image, input.name, d.output_image, output.name, d.cvmfs_repo
	FROM wish AS d
	JOIN image_name as input
	JOIN image_name as output
	WHERE 
		d.input_image = input.image_id
		AND d.output_image = output.image_id
		AND input.image_id = :input_id
		AND output.image_id = :output_id
		AND d.cvmfs_repo = :repo;
`

func GetWishF(inputId, outputId int, repo string) (wish WishFriendly, err error) {
	db, err := sql.Open("sqlite3", Database())
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getWishFStmt, err := db.Prepare(getWishF)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	var id, input_id, output_id int
	var input_name, output_name, cvmfsRepo string
	err = getWishFStmt.QueryRow(
		sql.Named("input_id", inputId),
		sql.Named("output_id", outputId),
		sql.Named("repo", repo),
	).Scan(&id, &input_id, &input_name, &output_id, &output_name, &cvmfsRepo)
	if err != nil {
		return
	}
	wish.Id = id
	wish.InputId = input_id
	wish.InputName = input_name
	wish.OutputId = output_id
	wish.OutputName = output_name
	wish.CvmfsRepo = cvmfsRepo
	return
}

var addConverted = `INSERT INTO converted VALUES(:wish, :input_reference, json(:manifest));`

func AddConverted(wishId int, manifest da.Manifest) error {
	db, err := sql.Open("sqlite3", Database())
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	AddConvertedStmt, err := db.Prepare(addConverted)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	inputReference := manifest.Config.Digest
	manifestJson, err := json.Marshal(manifest)
	if err != nil {
		LogE(err).Fatal("Impossible to encode the manifest into JSON")
	}
	_, err = AddConvertedStmt.Exec(
		sql.Named("wish", wishId),
		sql.Named("input_reference", inputReference),
		sql.Named("manifest", manifestJson),
	)
	return err
}

var deleteWish = `DELETE FROM wish WHERE id = ?;`

func DeleteWish(wishId int) (int, error) {
	db, err := sql.Open("sqlite3", Database())
	if err != nil {
		return 0, err
	}
	deleteWishStmt, err := db.Prepare(deleteWish)
	if err != nil {
		return 0, err
	}
	res, err := deleteWishStmt.Exec(wishId)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	return int(n), err
}

var deleteAllConverted = `DELETE FROM converted`

func DeleteAllConverted() (int, error) {
	db, err := sql.Open("sqlite3", Database())
	if err != nil {
		return 0, err
	}
	deleteAllConvertedStmt, err := db.Prepare(deleteAllConverted)
	if err != nil {
		return 0, err
	}
	res, err := deleteAllConvertedStmt.Exec()
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	return int(n), err
}

var getAllNeededImages = `SELECT printf("%s://%s/%s@%s", i.scheme, i.registry, i.repository, converted.input_reference) FROM converted, wish, image as i WHERE
converted.wish = wish.id AND
wish.input_image = i.id;
`

var getAllNeededLayers = fmt.Sprintf(`
SELECT DISTINCT( 
	'/cvmfs/' || wish.cvmfs_repo || '/'|| '%s' || '/' || substr(json_extract(layers.value, '$.Digest'), 8)
	) 
FROM 
converted, wish, json_each(converted.manifest, '$.Layers') as layers WHERE
converted.wish = wish.id;
`, subDirInsideRepo)

func GetAllNeededLayers() ([]string, error) {
	db, err := sql.Open("sqlite3", Database())
	if err != nil {
		LogE(err).Fatal("Impossible to open the database.")
	}
	getAllNeededLayersStmt, err := db.Prepare(getAllNeededLayers)
	if err != nil {
		LogE(err).Fatal("Impossible to create the statement.")
	}
	rows, err := getAllNeededLayersStmt.Query()
	var layers []string
	defer rows.Close()
	if err != nil {
		return layers, err
	}
	for rows.Next() {
		var layer string
		err = rows.Scan(&layer)
		if err != nil {
			return layers, err
		}
		layers = append(layers, layer)
	}
	return layers, nil
}
