package knowledge

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SourceUpdatesSuite struct {
	suite.Suite
}

func (s *SourceUpdatesSuite) TestShouldUpsertForCreatingGraph() {
	g := NewGraph()
	ip1 := g.AddAsset("ip", "127.0.0.1")
	ip2 := g.AddAsset("ip", "192.168.0.1")

	rel := g.AddRelation(ip1, "linked", ip2)

	bulk := GenerateGraphUpdatesBulk(nil, g)

	s.Require().Len(bulk.GetAssetUpserts(), 2)
	s.Require().Len(bulk.GetRelationUpserts(), 1)
	s.Require().Len(bulk.GetAssetRemovals(), 0)
	s.Require().Len(bulk.GetRelationRemovals(), 0)

	s.Assert().ElementsMatch(bulk.GetAssetUpserts(), []Asset{Asset(ip2), Asset(ip1)})
	s.Assert().ElementsMatch(bulk.GetRelationUpserts(), []Relation{rel})
}

func (s *SourceUpdatesSuite) TestShouldUpsertAssets() {
	g1 := NewGraph()
	ip1 := g1.AddAsset("ip", "127.0.0.1")
	ip2 := g1.AddAsset("ip", "192.168.0.1")

	g1.AddRelation(ip1, "linked", ip2)
	g2 := g1.Copy()

	ip3 := g2.AddAsset("ip", "10.0.0.1")
	ip4 := g2.AddAsset("ip", "10.0.0.2")

	bulk := GenerateGraphUpdatesBulk(g1, g2)

	s.Require().Len(bulk.GetAssetUpserts(), 2)
	s.Require().Len(bulk.GetRelationUpserts(), 0)
	s.Require().Len(bulk.GetAssetRemovals(), 0)
	s.Require().Len(bulk.GetRelationRemovals(), 0)

	s.Assert().ElementsMatch(bulk.GetAssetUpserts(), []Asset{Asset(ip3), Asset(ip4)})
}

func (s *SourceUpdatesSuite) TestShouldUpsertRelations() {
	g1 := NewGraph()
	ip1 := g1.AddAsset("ip", "127.0.0.1")
	ip2 := g1.AddAsset("ip", "192.168.0.1")

	g1.AddRelation(ip1, "linked", ip2)
	g2 := g1.Copy()

	ip3 := g2.AddAsset("ip", "10.0.0.1")
	r1 := g2.AddRelation(ip3, "linked", ip1)
	r2 := g2.AddRelation(ip3, "linked", ip2)

	bulk := GenerateGraphUpdatesBulk(g1, g2)

	s.Require().Len(bulk.GetAssetUpserts(), 1)
	s.Require().Len(bulk.GetRelationUpserts(), 2)
	s.Require().Len(bulk.GetAssetRemovals(), 0)
	s.Require().Len(bulk.GetRelationRemovals(), 0)

	s.Assert().ElementsMatch(bulk.GetAssetUpserts(), []Asset{Asset(ip3)})
	s.Assert().ElementsMatch(bulk.GetRelationUpserts(), []Relation{r1, r2})
}

func (s *SourceUpdatesSuite) TestShouldRemoveGraph() {
	g1 := NewGraph()
	ip1 := g1.AddAsset("ip", "127.0.0.1")
	ip2 := g1.AddAsset("ip", "192.168.0.1")
	r := g1.AddRelation(ip1, "linked", ip2)

	bulk := GenerateGraphUpdatesBulk(g1, nil)

	s.Require().Len(bulk.GetAssetUpserts(), 0)
	s.Require().Len(bulk.GetRelationUpserts(), 0)
	s.Require().Len(bulk.GetAssetRemovals(), 2)
	s.Require().Len(bulk.GetRelationRemovals(), 1)

	s.Assert().ElementsMatch(bulk.GetAssetRemovals(), []Asset{Asset(ip1), Asset(ip2)})
	s.Assert().ElementsMatch(bulk.GetRelationRemovals(), []Relation{r})
}

func (s *SourceUpdatesSuite) TestShouldGenerateBulkOfSubgraph() {
	g1 := NewGraph()
	ip1 := g1.AddAsset("ip", "127.0.0.1")
	ip2 := g1.AddAsset("ip", "192.168.0.1")
	r := g1.AddRelation(ip1, "linked", ip2)

	g2 := NewGraph()
	g2.AddAsset("ip", "127.0.0.1")

	bulk := GenerateGraphUpdatesBulk(g1, g2)

	s.Require().Len(bulk.GetAssetUpserts(), 0)
	s.Require().Len(bulk.GetRelationUpserts(), 0)
	s.Require().Len(bulk.GetAssetRemovals(), 1)
	s.Require().Len(bulk.GetRelationRemovals(), 1)

	s.Assert().ElementsMatch(bulk.GetAssetRemovals(), []Asset{Asset(ip2)})
	s.Assert().ElementsMatch(bulk.GetRelationRemovals(), []Relation{r})
}

func (s *SourceUpdatesSuite) TestShouldGenerateBulkForMixedAdditionsAndRemovals() {
	g1 := NewGraph()
	ip1 := g1.AddAsset("ip", "127.0.0.1")
	ip2 := g1.AddAsset("ip", "192.168.0.1")
	r := g1.AddRelation(ip1, "linked", ip2)

	g2 := NewGraph()
	g2.AddAsset("ip", "127.0.0.1")
	ip3 := g2.AddAsset("ip", "10.0.0.1")
	r2 := g2.AddRelation(ip3, "linked", ip2)

	bulk := GenerateGraphUpdatesBulk(g1, g2)

	s.Require().Len(bulk.GetAssetUpserts(), 1)
	s.Require().Len(bulk.GetRelationUpserts(), 1)
	s.Require().Len(bulk.GetAssetRemovals(), 1)
	s.Require().Len(bulk.GetRelationRemovals(), 1)

	s.Assert().ElementsMatch(bulk.GetAssetUpserts(), []Asset{Asset(ip3)})
	s.Assert().ElementsMatch(bulk.GetRelationUpserts(), []Relation{r2})
	s.Assert().ElementsMatch(bulk.GetAssetRemovals(), []Asset{Asset(ip2)})
	s.Assert().ElementsMatch(bulk.GetRelationRemovals(), []Relation{r})
}

func TestGraphUpdatesSuite(t *testing.T) {
	suite.Run(t, new(SourceUpdatesSuite))
}
