package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
)

func TestCounterCheck(t *testing.T) {
	type want struct {
		errvalue error
	}
	tests := []struct {
		name           string
		pollduration   int
		reportduration int
		want           want
	}{
		{
			name:           "equal poll report test #1",
			pollduration:   3,
			reportduration: 3,
			want: want{
				errvalue: errors.New("reportduration needs to be larger than pollduration"),
			},
		},
		{
			name:           "large poll test #2",
			pollduration:   6,
			reportduration: 3,
			want: want{
				errvalue: errors.New("reportduration needs to be larger than pollduration"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := CounterCheck(tt.pollduration, tt.reportduration)

			assert.Equal(t, tt.want.errvalue, v)
		})
	}
}

func TestReportUpdateBatch(t *testing.T) {
	type want struct {
		errvalue error
	}
	tests := []struct {
		name           string
		pollduration   int
		reportduration int
		want           want
	}{
		{
			name:           "equal poll report test #1",
			pollduration:   3,
			reportduration: 3,
			want: want{
				errvalue: errors.New("reportduration needs to be larger than pollduration"),
			},
		},
		{
			name:           "large poll test #2",
			pollduration:   6,
			reportduration: 3,
			want: want{
				errvalue: errors.New("reportduration needs to be larger than pollduration"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ReportUpdateBatch(tt.pollduration, tt.reportduration)

			assert.Equal(t, tt.want.errvalue, v)
		})
	}
}

func TestCollectStats(t *testing.T) {

	tests := []struct {
		name string
	}{
		{
			name: "trial run",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CollectStats()

		})
	}
}

func TestCollectMemStats(t *testing.T) {

	tests := []struct {
		name string
	}{
		{
			name: "trial run",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CollectMemStats()

		})
	}
}

func TestReportStats(t *testing.T) {

	tests := []struct {
		name string
	}{
		{
			name: "trial run",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ReportStats()

		})
	}
}

func TestSendMemStats(t *testing.T) {

	tests := []struct {
		name           string
		urlString string
	}{
		{
			name: "trial run",
			urlString: "",
		},
	}
	for _, tt := range tests {
		metricsObj:= metrics.Metrics{}
		t.Run(tt.name, func(t *testing.T) {
			SendMemStats(metricsObj, tt.urlString)

		})
	}
}


func ExampleCollectStats() {

	CollectStats()

}

func ExampleCollectMemStats() {

	CollectMemStats()

}

func ExampleReportStats() {

	ReportStats()

}
