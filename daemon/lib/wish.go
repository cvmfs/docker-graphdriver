package lib

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

type Wish struct {
	Id          int
	InputImage  int
	OutputImage int
	CvmfsRepo   string
}

func (w Wish) Equal(other Wish) bool {
	return (w.InputImage == other.InputImage) && (w.OutputImage == other.OutputImage) && (w.CvmfsRepo == other.CvmfsRepo)
}

type WishFriendly struct {
	Id         int
	InputId    int
	InputName  string
	OutputId   int
	OutputName string
	CvmfsRepo  string
	Converted  bool
}

type WishAlreadyInDBError struct{}

func (e *WishAlreadyInDBError) Error() string {
	return "Wish is already in the database"
}

func PrintMultipleWishes(wish []WishFriendly, machineFriendly, printHeader bool) {
	if machineFriendly {
		if printHeader {
			fmt.Println("id,input_image_id,input_image_name,cvmfs_repo,output_image_id,output_image_name,converted")
		}
		for _, d := range wish {
			fmt.Printf("%d,%d,%s,%s,%d,%s,%t\n", d.Id, d.InputId, d.InputName, d.CvmfsRepo, d.OutputId, d.OutputName, d.Converted)
		}
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		if printHeader {

			table.SetHeader([]string{"Id", "Input Id",
				"Input Image Name", "CVMFS Repo",
				"Output Id", "Output Image Name", "Converted"})

		}
		for _, d := range wish {
			table.Append([]string{strconv.Itoa(d.Id), strconv.Itoa(d.InputId),
				d.InputName, d.CvmfsRepo,
				strconv.Itoa(d.OutputId), d.OutputName, strconv.FormatBool(d.Converted)})
		}
		table.Render()
	}
}

func CreateWish(inputImage, outputImage, cvmfsRepo, userInput, userOutput string) (wish Wish, err error) {

	inputImg, err := ParseImage(inputImage)
	if err != nil {
		err = fmt.Errorf("%s | %s", err.Error(), "Error in parsing the input image")
		return
	}
	inputImg.User = userInput
	outputImg, err := ParseImage(outputImage)
	if err != nil {
		err = fmt.Errorf("%s | %s", err.Error(), "Error in parsing the output image")
		return
	}
	outputImg.User = userOutput
	outputImg.IsThin = true

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

	// both images are already in our database if we enter the if
	// check if also the wish itself is already in the database
	if errIn == nil && errOut == nil {
		inputId := inputImgDb.Id
		outputId := outputImgDb.Id
		wishInDb, errWishDb := GetWish(inputId, outputId, cvmfsRepo)
		if errWishDb == nil {
			err = &WishAlreadyInDBError{}
			return wishInDb, err
		}
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
	wish.Id = 0
	wish.InputImage = inputImgDbId
	wish.OutputImage = outputImgDbId
	wish.CvmfsRepo = cvmfsRepo
	return
}
