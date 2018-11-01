package lib

import (
	"fmt"
)

type Wish struct {
	Id          int
	InputImage  int
	OutputImage int
	CvmfsRepo   string
}

/*
func (w Wish) Equal(other Wish) bool {
	return (w.InputImage == other.InputImage) && (w.OutputImage == other.OutputImage) && (w.CvmfsRepo == other.CvmfsRepo)
}
*/

type WishFriendly struct {
	Id         int
	InputId    int
	InputName  string
	OutputId   int
	OutputName string
	CvmfsRepo  string
	Converted  bool
}

/*
func (w WishFriendly) Equal(other WishFriendly) bool {
	return ((w.InputId == other.InputId) &&
		(w.OutputId == other.OutputId) &&
		(w.CvmfsRepo == other.CvmfsRepo)) ||

		((w.InputName == other.InputName) &&
			(w.OutputName == other.OutputName) &&
			(w.CvmfsRepo == other.CvmfsRepo))
}
*/

type WishAlreadyInDBError struct{}

func (e *WishAlreadyInDBError) Error() string {
	return "Wish is already in the database"
}

func CreateWish(inputImage, outputImage, cvmfsRepo, userInput, userOutput string) (wish WishFriendly, err error) {

	inputImg, err := ParseImage(inputImage)
	if err != nil {
		err = fmt.Errorf("%s | %s", err.Error(), "Error in parsing the input image")
		return
	}
	inputImg.User = userInput

	wish.InputName = inputImg.WholeName()

	outputImg, err := ParseImage(outputImage)
	if err != nil {
		err = fmt.Errorf("%s | %s", err.Error(), "Error in parsing the output image")
		return
	}
	outputImg.User = userOutput
	outputImg.IsThin = true
	wish.OutputName = outputImg.WholeName()

	/*
		// check if the images are already in the database
		inputImgDb, errIn := GetImage(inputImg)
		outputImgDb, errOut := GetImage(outputImg)

		// some error, that is not because the image is not in the db
		if errIn != nil && errIn != sql.ErrNoRows {
			err = fmt.Errorf("%s | %s", errIn.Error(), "Error in querying the database for the input image")
			return
		}
		if errOut != nil && errOut != sql.ErrNoRows {
			err = fmt.Errorf("%s | %s", errIn.Error(), "Error in querying the database for the output image")
			return
		}

		// the input image is not in the db, we download again the manifest
		if errIn != nil && errIn == sql.ErrNoRows {
			// trying to get the input manifest, if we are not able to get the input image manifest there is something wrong, hence we avoid to add the wish to the database itself
			_, err = inputImg.GetManifest()
			if err != nil {
				err = fmt.Errorf("%s | %s", err.Error(), "Impossible to get the input manifest")
				return
			}
		}

		// we need to identify the real input and output, if they are already in the db we can use them directly, otherwise we first add them
	*/
	/*
		var (
			inputImgDbId, outputImgDbId int
		)

			if inputImgDb.Id != 0 {
				inputImgDbId = inputImgDb.Id
			} else {
				err = AddImage(inputImg)
				if err != nil {
					err = fmt.Errorf("%s | %s", err.Error(), "Impossible to add the input image to the database")
					return
				}
				inputImgDbId, err = GetImageId(inputImg)
				if err != nil {
					err = fmt.Errorf("%s | %s", err.Error(), "Impossible to get the ID of the input image just added")
					return
				}
			}
			if outputImgDb.Id != 0 {
				outputImgDbId = outputImgDb.Id
			} else {
				err = AddImage(outputImg)
				if err != nil {
					err = fmt.Errorf("%s | %s", err.Error(), "Impossible to add the out image to the database")
					return
				}
				outputImgDbId, err = GetImageId(outputImg)
				if err != nil {
					err = fmt.Errorf("%s | %s", err.Error(), "Impossible to get the ID of the output image just added")
					return
				}

			}
			// at this point we add the real wish to the database
	*/
	wish.Id = 0
	//wish.InputId = inputImgDbId
	//wish.OutputId = outputImgDbId
	wish.InputId = 0
	wish.OutputId = 0
	wish.CvmfsRepo = cvmfsRepo
	return
}
