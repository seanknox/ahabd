package fixer

import (
	"errors"
	"testing"

	"github.com/juan-lee/ahabd/pkg/fixer/stats"

	"github.com/stretchr/testify/assert"
)

type fakeStats struct {
	needsFixing int
	fixed       int
	fixFail     int
}

func (s *fakeStats) IncNeedsFixing() {
	s.needsFixing++
}

func (s *fakeStats) IncFixed() {
	s.fixed++
}

func (s *fakeStats) IncFixFail() {
	s.fixFail++
}

type fakeFixer struct {
	needsFixing bool
	err         error
	fixed       bool
	stats       *fakeStats
}

func (f *fakeFixer) NeedsFixing() bool {
	return f.needsFixing
}

func (f *fakeFixer) Fix() error {
	if f.err == nil {
		f.fixed = true
	}
	return f.err
}

func (f *fakeFixer) Stats() stats.Stats {
	if f.stats != nil {
		return f.stats
	}
	return nil
}

func TestNoFixNullStats(t *testing.T) {
	ff := &fakeFixer{needsFixing: false, err: nil}
	assert.Nil(t, Fix(ff))
	assert.Equal(t, false, ff.fixed)
	assert.Nil(t, ff.stats)
}

func TestNoFix(t *testing.T) {
	ff := &fakeFixer{needsFixing: false, err: nil, stats: &fakeStats{}}
	assert.Nil(t, Fix(ff))
	assert.Equal(t, false, ff.fixed)
	assert.Equal(t, 0, ff.stats.needsFixing)
	assert.Equal(t, 0, ff.stats.fixed)
	assert.Equal(t, 0, ff.stats.fixFail)
}

func TestNeedsFixNullStats(t *testing.T) {
	ff := &fakeFixer{needsFixing: true, err: nil}
	assert.Nil(t, Fix(ff))
	assert.Equal(t, true, ff.fixed)
	assert.Nil(t, ff.stats)
}

func TestNeedsFix(t *testing.T) {
	ff := &fakeFixer{needsFixing: true, err: nil, stats: &fakeStats{}}
	assert.Nil(t, Fix(ff))
	assert.Equal(t, true, ff.fixed)
	assert.Equal(t, 1, ff.stats.needsFixing)
	assert.Equal(t, 1, ff.stats.fixed)
}

func TestNeedsFixError(t *testing.T) {
	ff := &fakeFixer{needsFixing: true, err: errors.New("Fix Failed"), stats: &fakeStats{}}
	assert.Equal(t, ff.err, Fix(ff))
	assert.Equal(t, false, ff.fixed)
	assert.Equal(t, 1, ff.stats.needsFixing)
	assert.Equal(t, 0, ff.stats.fixed)
	assert.Equal(t, 1, ff.stats.fixFail)
}
