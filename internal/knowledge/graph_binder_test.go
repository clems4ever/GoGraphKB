package knowledge

import (
	"testing"

	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/stretchr/testify/assert"
)

func TestShouldRelateAssets(t *testing.T) {
	g := NewGraph()

	binder := NewGraphBinder(g)

	relation := schema.RelationType{
		FromType: "from_type",
		ToType:   "to_type",
		Type:     "rel_type",
	}
	binder.Relate("from", relation, "to")

	assert.Len(t, g.Assets(), 2)
	assert.Len(t, g.Relations(), 1)

	assert.Equal(t, g.Assets(), map[Asset]graphEntryAction{
		{Type: "from_type", Key: "from"}: graphEntryAdd,
		{Type: "to_type", Key: "to"}:     graphEntryAdd,
	})
}

func TestShouldBindAsset(t *testing.T) {
	g := NewGraph()

	binder := NewGraphBinder(g)
	binder.Bind("from", "from_type")

	assert.Len(t, g.Assets(), 1)
	assert.Len(t, g.Relations(), 0)

	assert.Equal(t, g.Assets(), map[Asset]graphEntryAction{
		{Type: "from_type", Key: "from"}: graphEntryAdd,
	})
}
