package appmetadata

import "github.com/pyroscope-io/pyroscope/pkg/storage/metadata"

type ApplicationMetadata struct {
	// Fully Qualified Name. Eg app.cpu ({__name__}.{profile_type})
	FQName string `gorm:"column:fq_name;index,unique;not null;default:null" json:"name"`

	SpyName         string                   `gorm:"column:spy_name;json:"spyName,omitempty"`
	SampleRate      uint32                   `gorm:"column:sample_rate;json:"sampleRate,omitempty"`
	Units           metadata.Units           `gorm:"column:units;json:"units,omitempty"`
	AggregationType metadata.AggregationType `gorm:"column:aggregation_type;json:"-"`
}

// type ApplicationMetadata struct {
// 	// Fully Qualified Name. Eg app.cpu ({__name__}.{profile_type})
// 	FQName string `gorm:"index,unique;not null;default:null" json:"name"`

// 	SpyName         string                   `json:"spyName,omitempty"`
// 	SampleRate      uint32                   `json:"sampleRate,omitempty"`
// 	Units           metadata.Units           `json:"units,omitempty"`
// 	AggregationType metadata.AggregationType `json:"-"`
// }
