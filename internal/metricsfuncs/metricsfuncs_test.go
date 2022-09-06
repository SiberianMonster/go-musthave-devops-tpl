package metricsfuncs_test

import (
	"testing"
	"errors"
	"github.com/stretchr/testify/assert"
	"go-musthave-devops-tpl/internal/metricsfuncs"
)

func TestReportUpdate(t *testing.T) {
    type want struct {
        errvalue    error
    }
    tests := []struct {
		name string
        pollduration int
        reportduration int
        want want
    }{
        {
            name: "equal poll report test #1",
            pollduration: 3,
        	reportduration: 3,
            want: want{
                errvalue: errors.New("reportduration needs to be larger than pollduration"),
            },
        },
		{
            name: "large poll test #2",
            pollduration: 6,
        	reportduration: 3,
            want: want{
                errvalue: errors.New("reportduration needs to be larger than pollduration"),
            },
		},
	}
	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            v := metricsfuncs.ReportUpdate(tt.pollduration, tt.reportduration)

            assert.Equal(t, tt.want.errvalue, v)
        })
    }
}

