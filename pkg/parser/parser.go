// Package parser deals with parsing various incoming formats
package parser

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/pyroscope-io/pyroscope/pkg/ingestion"
	"github.com/pyroscope-io/pyroscope/pkg/storage"
)

func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

type Parser struct {
	log      *logrus.Logger
	putter   storage.Putter
	exporter storage.MetricsExporter
}

func New(log *logrus.Logger, s storage.Putter, exporter storage.MetricsExporter) *Parser {
	return &Parser{
		log:      log,
		putter:   s,
		exporter: exporter,
	}
}

func (p *Parser) Ingest(ctx context.Context, in *ingestion.IngestInput) error {
	// fmt.Println("data ingested here ....", in.Metadata.AccountUID)

	// for self-profiling data
	if in.Metadata.AccountUID == "" {
		in.Metadata.AccountUID = "keval12"
	}
	// in.Metadata.AccountUID = RandomString(7)
	fmt.Println("in.Metadata.AccountUID", in.Metadata.AccountUID)

	// jsonData, err := json.Marshal(in)
	// if err != nil {
	// 	fmt.Println("Error marshaling struct to JSON:", err)
	// 	// return
	// }

	// // Print the JSON data
	// fmt.Println("data ingested here ....", jsonData)

	fmt.Println("in.Profile", reflect.TypeOf(in.Profile))

	updateMetrics(in)
	fmt.Println("p.putter, p.exporter, in.Metadata", p.putter, p.exporter, in.Metadata)
	return in.Profile.Parse(ctx, p.putter, p.exporter, in.Metadata)
}
