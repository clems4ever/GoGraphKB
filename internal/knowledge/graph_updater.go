package knowledge

import (
	"context"
	"fmt"

	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/sirupsen/logrus"
)

// SourceSubGraphUpdates represents the updates to perform on a source subgraph
type SourceSubGraphUpdates struct {
	Updates GraphUpdatesBulk
	Schema  schema.SchemaGraph
	Source  string
}

// GraphUpdater represents the updater of graph
type GraphUpdater struct {
	graphDB         GraphDB
	schemaPersistor schema.Persistor
}

// NewGraphUpdater create a new instance of graph updater
func NewGraphUpdater(graphDB GraphDB, schemaPersistor schema.Persistor) *GraphUpdater {
	return &GraphUpdater{graphDB, schemaPersistor}
}

// UpdateSchema update the schema for the source with the one provided in the request
func (sl *GraphUpdater) UpdateSchema(source string, sg schema.SchemaGraph) error {
	previousSchema, err := sl.schemaPersistor.LoadSchema(context.Background(), source)
	if err != nil {
		return fmt.Errorf("Unable to read schema from DB: %v", err)
	}

	schemaEqual := previousSchema.Equal(sg)

	if !schemaEqual {
		logrus.Debug("The schema needs an update")
		if err := sl.schemaPersistor.SaveSchema(context.Background(), source, sg); err != nil {
			return fmt.Errorf("Unable to write schema in DB: %v", err)
		}
	}
	return nil
}

// InsertAssets insert multiple assets in the graph of the data source
func (sl *GraphUpdater) InsertAssets(source string, assets []Asset) error {
	if err := sl.graphDB.InsertAssets(source, assets); err != nil {
		return fmt.Errorf("Unable to insert assets from source %s: %v", source, err)
	}
	return nil
}

// InsertRelations insert multiple relations in the graph of the data source
func (sl *GraphUpdater) InsertRelations(source string, relations []Relation) error {
	if err := sl.graphDB.InsertRelations(source, relations); err != nil {
		return fmt.Errorf("Unable to insert relations from source %s: %v", source, err)
	}
	return nil
}

// RemoveAssets remove multiple assets from the graph of the data source
func (sl *GraphUpdater) RemoveAssets(source string, assets []Asset) error {
	if err := sl.graphDB.RemoveAssets(source, assets); err != nil {
		return fmt.Errorf("Unable to remove assets from source %s: %v", source, err)
	}
	return nil
}

// RemoveRelations remove multiple relations from the graph of the data source
func (sl *GraphUpdater) RemoveRelations(source string, relations []Relation) error {
	if err := sl.graphDB.RemoveRelations(source, relations); err != nil {
		return fmt.Errorf("Unable to remove relations from source %s: %v", source, err)
	}
	return nil
}
