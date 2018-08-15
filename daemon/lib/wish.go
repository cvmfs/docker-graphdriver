package lib

import (
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

type WishFriendly struct {
	Id         int
	InputId    int
	InputName  string
	OutputId   int
	OutputName string
	CvmfsRepo  string
	Converted  bool
}

func (d WishFriendly) PrintWish(machineFriendly, printHeader bool) {
	if machineFriendly {
		if printHeader {
			fmt.Println("id,input_image_id,input_image_name,cvmfs_repo,output_image_id,output_image_name,converted")
		}
		fmt.Printf("%d,%d,%s,%s,%d,%s,%t\n", d.Id, d.InputId, d.InputName, d.CvmfsRepo, d.OutputId, d.OutputName, d.Converted)
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeader([]string{"Id", "Input Id",
			"Input Image Name", "CVMFS Repo",
			"Output Id", "Output Image Name", "Converted"})
		table.Append([]string{strconv.Itoa(d.Id), strconv.Itoa(d.InputId),
			d.InputName, d.CvmfsRepo,
			strconv.Itoa(d.OutputId), d.OutputName, strconv.FormatBool(d.Converted)})
		table.Render()
	}
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
