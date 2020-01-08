package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/clems4ever/go-graphkb/internal/database"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Database the selected database
var Database *database.MariaDB

// ConfigPath string
var ConfigPath string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	rootCmd := &cobra.Command{
		Use: "go-graphkb [opts]",
	}

	listenCmd := &cobra.Command{
		Use: "listen",
		Run: listen,
	}

	cleanCmd := &cobra.Command{
		Use: "count",
		Run: count,
	}

	countCmd := &cobra.Command{
		Use: "flush",
		Run: flush,
	}

	readCmd := &cobra.Command{
		Use:  "read [source]",
		Run:  read,
		Args: cobra.ExactArgs(1),
	}

	queryCmd := &cobra.Command{
		Use:  "query [query]",
		Run:  queryFunc,
		Args: cobra.ExactArgs(1),
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "config.yml", "Provide the path to the configuration file (required)")

	cobra.OnInitialize(onInit)

	rootCmd.AddCommand(cleanCmd, listenCmd, countCmd, readCmd, queryCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func onInit() {
	viper.SetConfigFile(ConfigPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Cannot read configuration file from %s", ConfigPath))
	}

	fmt.Println("Using config file:", viper.ConfigFileUsed())

	dbName := viper.GetString("mariadb.database")
	if dbName == "" {
		log.Fatal("Please provide database_name option in your configuration file")
	}
	Database = database.NewMariaDB(
		viper.GetString("mariadb.username"),
		viper.GetString("mariadb.password"),
		viper.GetString("mariadb.host"),
		dbName)
}

func count(cmd *cobra.Command, args []string) {
	countAssets, err := Database.CountAssets()
	if err != nil {
		log.Fatal(err)
	}

	countRelations, err := Database.CountRelations()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d assets\n%d relations\n", countAssets, countRelations)
}

func flush(cmd *cobra.Command, args []string) {
	if err := Database.FlushAll(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successul flush")
}

func listen(cmd *cobra.Command, args []string) {
	eventBus := make(chan knowledge.SourceSubGraphUpdates)
	listener := knowledge.NewGraphUpdater(Database, Database)

	listener.Listen(eventBus)

	if err := Database.InitializeSchema(); err != nil {
		log.Fatal(err)
	}

	server.StartServer(Database, Database, eventBus)

	close(eventBus)
}

func read(cmd *cobra.Command, args []string) {
	g := knowledge.NewGraph()
	err := Database.ReadGraph(args[0], g)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("assets = %d\nrelations = %d\n", len(g.Assets()), len(g.Relations()))
}

func queryFunc(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	q := knowledge.NewQuerier(Database)

	r, err := q.Query(ctx, args[0])
	if err != nil {
		log.Fatal(err)
	}

	resultsCount := 0
	for r.Cursor.HasMore() {
		var m interface{}
		err := r.Cursor.Read(context.Background(), &m)
		if err != nil {
			log.Fatal(err)
		}

		doc := m.([]interface{})
		ldoc := make([]string, len(doc))
		for i, d := range doc {
			ldoc[i] = fmt.Sprintf("%v", d)
		}
		fmt.Println(ldoc)
		resultsCount++
	}

	totalTime := r.Statistics.Parsing + r.Statistics.Execution

	fmt.Printf("%d results found in %fms\n", resultsCount, float64(totalTime.Microseconds())/1000.0)
}
