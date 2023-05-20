package appmetadata

import "github.com/pyroscope-io/pyroscope/pkg/storage/metadata"

type ApplicationMetadata struct {
	FQName          string                   `gorm:"index:idx_fqname_accountuid,unique;not null;default:null" json:"name"`
	SpyName         string                   `json:"spyName,omitempty"`
	SampleRate      uint32                   `json:"sampleRate,omitempty"`
	Units           metadata.Units           `json:"units,omitempty"`
	AggregationType metadata.AggregationType `json:"-"`
	// AccountUID      string                   `gorm:"index:idx_fqname_accountuid" json:"accountUid,omitempty" `
	AccountUID string `gorm:"column:accountUid;json:"accountUid,omitempty;index:idx_fqname_accountuid,unique;not null;default:null"`
}
