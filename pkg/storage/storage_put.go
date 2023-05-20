package storage

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/pyroscope-io/pyroscope/pkg/model/appmetadata"
	"github.com/pyroscope-io/pyroscope/pkg/storage/dimension"
	"github.com/pyroscope-io/pyroscope/pkg/storage/metadata"
	"github.com/pyroscope-io/pyroscope/pkg/storage/segment"
	"github.com/pyroscope-io/pyroscope/pkg/storage/tree"
)

type PutInput struct {
	AccountUID      string
	StartTime       time.Time
	EndTime         time.Time
	Key             *segment.Key
	Val             *tree.Tree
	SpyName         string
	SampleRate      uint32
	Units           metadata.Units
	AggregationType metadata.AggregationType
}

func (s *Storage) Put(ctx context.Context, pi *PutInput) error {
	if s.hc.IsOutOfDiskSpace() {
		return errOutOfSpace
	}
	if pi.StartTime.Before(s.retentionPolicy().LowerTimeBoundary()) {
		return errRetention
	}

	// fmt.Println("storage.Put >>>>> Pi.Key >>>>>", pi.Key)

	if err := segment.ValidateKey(pi.Key); err != nil {
		return err
	}

	// fmt.Println("s.appsvc >>>>>", reflect.TypeOf(s.appSvc))
	if err := s.appSvc.CreateOrUpdate(ctx, appmetadata.ApplicationMetadata{
		AccountUID:      pi.AccountUID,
		FQName:          pi.Key.AppName(),
		SpyName:         pi.SpyName,
		SampleRate:      pi.SampleRate,
		Units:           pi.Units,
		AggregationType: pi.AggregationType,
	}); err != nil {
		s.logger.Error("error saving metadata", err)
	}

	s.putTotal.Inc()

	// fmt.Println("--- pi.Key.HasProfileID()", pi.Key.HasProfileID())

	if pi.Key.HasProfileID() {
		if err := s.ensureAppSegmentExists(pi); err != nil {
			return err
		}
		return s.exemplars.insert(ctx, pi)
	}

	s.logger.WithFields(logrus.Fields{
		"startTime":       pi.StartTime.String(),
		"endTime":         pi.EndTime.String(),
		"key":             pi.Key.Normalized(),
		"samples":         pi.Val.Samples(),
		"units":           pi.Units,
		"aggregationType": pi.AggregationType,
	}).Debug("storage.Put")

	if err := s.labels.PutLabels(pi.Key.Labels()); err != nil {
		return fmt.Errorf("unable to write labels: %w", err)
	}

	sk := pi.Key.SegmentKey()
	for k, v := range pi.Key.Labels() {
		key := k + ":" + v
		r, err := s.dimensions.GetOrCreate(key)
		if err != nil {
			s.logger.Errorf("dimensions cache for %v: %v", key, err)
			continue
		}
		r.(*dimension.Dimension).Insert([]byte(sk))

		// mwdebug: reflect.TypeOf(s.dimensions) => *storage.db
		// mwdebug: key => __name__:k8s.ebpf.cpu

		// fmt.Println("dimension r", r)
		s.dimensions.Put(key, r)
	}

	r, err := s.segments.GetOrCreate(sk)
	if err != nil {
		return fmt.Errorf("segments cache for %v: %v", sk, err)
	}

	st := r.(*segment.Segment)
	st.SetMetadata(metadata.Metadata{
		SpyName:         pi.SpyName,
		SampleRate:      pi.SampleRate,
		Units:           pi.Units,
		AggregationType: pi.AggregationType,
	})

	// mwdebug: returns count of samples
	samples := pi.Val.Samples()

	err = st.Put(pi.StartTime, pi.EndTime, samples, func(depth int, t time.Time, r *big.Rat, addons []segment.Addon) {

		// mwdebug: returns something like  k:depth:unixtime
		tk := pi.Key.TreeKey(depth, t)

		res, err := s.trees.GetOrCreate(tk)
		if err != nil {
			s.logger.Errorf("trees cache for %v: %v", tk, err)
			return
		}
		cachedTree := res.(*tree.Tree)
		treeClone := pi.Val.Clone(r)
		for _, addon := range addons {
			if res, ok := s.trees.Lookup(pi.Key.TreeKey(addon.Depth, addon.T)); ok {
				ta := res.(*tree.Tree)
				ta.RLock()
				treeClone.Merge(ta)
				ta.RUnlock()
			}
		}
		cachedTree.Lock()
		cachedTree.Merge(treeClone)
		cachedTree.Unlock()
		s.trees.Put(tk, cachedTree)
	})
	if err != nil {
		return err
	}

	s.segments.Put(sk, st)
	return nil
}
