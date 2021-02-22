package knowledge

import (
	"context"
	"fmt"
	"time"

	"github.com/clems4ever/go-graphkb/internal/history"
	"github.com/clems4ever/go-graphkb/internal/query"
	"github.com/sirupsen/logrus"
)

type Querier struct {
	GraphDB    GraphDB
	historizer history.Historizer
}

type QuerierResult struct {
	Cursor      Cursor
	Projections []Projection
	Statistics  Statistics
}

func NewQuerier(db GraphDB, historizer history.Historizer) *Querier {
	return &Querier{GraphDB: db, historizer: historizer}
}

func (q *Querier) Query(ctx context.Context, queryString string) (*QuerierResult, error) {
	qr, sql, err := q.queryInternal(ctx, queryString)
	if err != nil {
		saveErr := q.historizer.SaveFailedQuery(ctx, queryString, sql, err)
		if saveErr != nil {
			return nil, fmt.Errorf("Unable to save query error in database: %v", saveErr)
		}
		return nil, err
	}

	saveErr := q.historizer.SaveSuccessfulQuery(ctx, queryString, sql, qr.Statistics.Execution/time.Millisecond)
	if saveErr != nil {
		return nil, fmt.Errorf("Unable to save query history in database: %v", saveErr)
	}
	return qr, nil
}

func (q *Querier) queryInternal(ctx context.Context, cypherQuery string) (*QuerierResult, string, error) {
	s := Statistics{}

	var err error
	var queryCypher *query.QueryCypher

	s.Parsing = MeasureDuration(func() {
		queryCypher, err = query.TransformCypher(cypherQuery)
	})

	if err != nil {
		return nil, "", err
	}

	translation, err := NewSQLQueryTranslator().Translate(queryCypher)
	if err != nil {
		return nil, "", err
	}

	var res *GraphQueryResult
	s.Execution = MeasureDuration(func() {
		res, err = q.GraphDB.Query(ctx, *translation)
	})

	if err != nil {
		return nil, translation.Query, err
	}

	logrus.Debugf("Found results in %dms\n", s.Execution/time.Millisecond)

	result := &QuerierResult{
		Cursor:      res.Cursor,
		Projections: res.Projections,
		Statistics:  s,
	}
	return result, translation.Query, nil
}

type Statistics struct {
	Parsing   time.Duration
	Execution time.Duration
}

func MeasureDuration(Func func()) time.Duration {
	now := time.Now()
	Func()
	return time.Since(now)
}
