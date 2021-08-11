// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// ResolverMetrics is an autogenerated mock type for the ResolverMetrics type
type ResolverMetrics struct {
	mock.Mock
}

// DNSCacheResolution provides a mock function with given fields:
func (_m *ResolverMetrics) DNSCacheResolution() {
	_m.Called()
}

// DNSLookupDuration provides a mock function with given fields: duration
func (_m *ResolverMetrics) DNSLookupDuration(duration time.Duration) {
	_m.Called(duration)
}

// DNSLookupResolution provides a mock function with given fields:
func (_m *ResolverMetrics) DNSLookupResolution() {
	_m.Called()
}