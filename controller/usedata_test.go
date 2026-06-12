package controller

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateDataTimeRangeAllowsNinetyDays(t *testing.T) {
	const daySeconds = int64(24 * 3600)
	end := int64(1767398400)
	start := end - 90*daySeconds

	require.NoError(t, validateDataTimeRange(start, end, 90))
}

func TestValidateDataTimeRangeRejectsMissingTime(t *testing.T) {
	require.EqualError(t, validateDataTimeRange(0, 0, 90), "开始时间和结束时间不能为空")
}

func TestValidateDataTimeRangeRejectsMoreThanNinetyDays(t *testing.T) {
	const daySeconds = int64(24 * 3600)
	end := int64(1767398400)
	start := end - 90*daySeconds - 1

	require.EqualError(t, validateDataTimeRange(start, end, 90), "时间跨度不能超过 90 天")
}

func TestValidateDataTimeRangeRejectsFutureEndTime(t *testing.T) {
	const daySeconds = int64(24 * 3600)
	end := currentUnixTime() + daySeconds
	start := end - daySeconds

	require.EqualError(t, validateDataTimeRange(start, end, 90), "结束时间不能晚于当前时间")
}
