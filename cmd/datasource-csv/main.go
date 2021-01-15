package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/clems4ever/go-graphkb/graphkb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CSVSource represent a CSV source file
type CSVSource struct {
	dataPath string
	graphAPI *graphkb.GraphAPI
}

// NewCSVSource create a new instance of CSV source
func NewCSVSource(graphAPI *graphkb.GraphAPI) *CSVSource {
	csvSource := new(CSVSource)
	csvSource.dataPath = viper.GetString("path")

	if csvSource.dataPath == "" {
		log.Fatal(fmt.Errorf("Unable to detect CSV file path in configuration. Check patch configuration is provided"))
	}

	csvSource.graphAPI = graphAPI
	return csvSource
}

// Publish the graph built from CSV
func (cs *CSVSource) Publish() error {
	file, err := os.Open(cs.dataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	r := csv.NewReader(file)

	previousGraph, err := cs.graphAPI.ReadCurrentGraph()
	if err != nil {
		return fmt.Errorf("Unable to read previous graph: %v", err)
	}

	tx := cs.graphAPI.CreateTransaction(previousGraph)

	header := true

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("Unable to read data in CSV file: %v", err)
		}

		// Skip header line
		if header {
			header = false
			continue
		}

		relationType := graphkb.RelationType{
			FromType: graphkb.AssetType(record[0]),
			ToType:   graphkb.AssetType(record[3]),
			Type:     graphkb.RelationKeyType(record[2]),
		}

		tx.Relate(record[1], relationType, record[4])
	}

	_, err = tx.Commit()
	fmt.Println("CSV data has been sent successfully")
	return err
}

// ConfigPath string
var ConfigPath string

func onInit() {
	viper.SetConfigFile(ConfigPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Cannot read configuration file from %s", ConfigPath))
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cobra.OnInitialize(onInit)

	rootCmd := &cobra.Command{
		Use: "source-csv [opts]",
		Run: func(cmd *cobra.Command, args []string) {

			options := graphkb.GraphAPIOptions{
				URL:        viper.GetString("graphkb_url"),
				AuthToken:  viper.GetString("graphkb_auth_token"),
				SkipVerify: viper.GetBool("graphkb_skip_verify"),
			}

			dataSource := NewCSVSource(graphkb.NewGraphAPI(options))

			if err := dataSource.Publish(); err != nil {
				panic(err)
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "config.yml",
		"Provide the path to the configuration file (required)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
