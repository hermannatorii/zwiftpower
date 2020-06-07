package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/lizrice/zwiftpower/zp"
	"github.com/spf13/cobra"
)

func getID(args []string, defaultID int) (id int) {
	id = defaultID
	if len(args) >= 1 {
		id64, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Printf("Can't parse ID: %v", err.Error())
			os.Exit(1)
		}
		id = int(id64)
	}
	return id
}

func main() {
	riderCmd := &cobra.Command{
		Use:   "rider [ID]",
		Short: "Import data for rider ID",
		Run: func(cmd *cobra.Command, args []string) {
			riderID := getID(args, 98588)
			rider, err := zp.ImportRider(riderID)
			if err != nil {
				fmt.Printf("Error getting rider: %v", err)
			}
			fmt.Printf("%v\n", rider.Strings())
		},
	}

	rootCmd := &cobra.Command{
		Use:   "zp [ID]",
		Short: "Import data for club ID",
		Long:  `Default club ID is 2672, Revolution Velo`,
		Run: func(cmd *cobra.Command, args []string) {
			clubID := getID(args, 2672)
			riders := zp.ImportZP(clubID)

			f, err := os.Create("result.csv")
			if err != nil {
				fmt.Printf("Error creating file: %v\n", err)
				os.Exit(1)
			}
			defer f.Close()

			writer := csv.NewWriter(f)
			defer writer.Flush()

			for i, rider := range riders {
				var err error
				name := rider.Name
				riders[i], err = zp.ImportRider(rider.Zwid)
				if err != nil {
					fmt.Printf("Error loading data for %s: %v\n", name, err)
					continue
				}
				riders[i].Name = name
				// fmt.Printf("%v\n", riders[i])
				err = writer.Write(riders[i].Strings())
				if err != nil {
					fmt.Printf("Error writing to file: %v\n", err)
				}
			}
		},
	}

	rootCmd.AddCommand(riderCmd)
	rootCmd.Execute()
}
