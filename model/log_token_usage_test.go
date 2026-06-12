package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetTokenUsageDailyAggregatesConsumeLogCostByTokenAndDay(t *testing.T) {
	truncateTables(t)

	dayOne := time.Date(2026, 1, 2, 10, 30, 0, 0, time.UTC).Unix()
	dayTwo := time.Date(2026, 1, 3, 9, 15, 0, 0, time.UTC).Unix()
	dayOneBucket := dayOne - dayOne%86400
	dayTwoBucket := dayTwo - dayTwo%86400

	logs := []*Log{
		{
			UserId:           1,
			Username:         "alice",
			CreatedAt:        dayOne,
			Type:             LogTypeConsume,
			TokenId:          101,
			TokenName:        "prod-key",
			Quota:            250,
			PromptTokens:     100,
			CompletionTokens: 30,
		},
		{
			UserId:           1,
			Username:         "alice",
			CreatedAt:        dayOne + 3600,
			Type:             LogTypeConsume,
			TokenId:          101,
			TokenName:        "prod-key",
			Quota:            75,
			PromptTokens:     20,
			CompletionTokens: 10,
		},
		{
			UserId:           1,
			Username:         "alice",
			CreatedAt:        dayOne + 7200,
			Type:             LogTypeConsume,
			TokenId:          202,
			TokenName:        "batch-key",
			Quota:            12,
			PromptTokens:     7,
			CompletionTokens: 3,
		},
		{
			UserId:           1,
			Username:         "alice",
			CreatedAt:        dayTwo,
			Type:             LogTypeConsume,
			TokenId:          101,
			TokenName:        "prod-key",
			Quota:            90,
			PromptTokens:     50,
			CompletionTokens: 5,
		},
		{
			UserId:           2,
			Username:         "bob",
			CreatedAt:        dayOne,
			Type:             LogTypeConsume,
			TokenId:          303,
			TokenName:        "bob-key",
			Quota:            1000,
			PromptTokens:     999,
			CompletionTokens: 1,
		},
		{
			UserId:           1,
			Username:         "alice",
			CreatedAt:        dayOne,
			Type:             LogTypeError,
			TokenId:          101,
			TokenName:        "prod-key",
			Quota:            9999,
			PromptTokens:     1000,
			CompletionTokens: 1000,
		},
	}
	require.NoError(t, LOG_DB.Create(&logs).Error)

	rows, err := GetTokenUsageDaily(1, "", dayOneBucket, dayTwoBucket+86399)
	require.NoError(t, err)

	require.Equal(t, []TokenUsageDaily{
		{
			TokenId:   101,
			TokenName: "prod-key",
			CreatedAt: dayOneBucket,
			Quota:     325,
			TokenUsed: 160,
			Count:     2,
		},
		{
			TokenId:   202,
			TokenName: "batch-key",
			CreatedAt: dayOneBucket,
			Quota:     12,
			TokenUsed: 10,
			Count:     1,
		},
		{
			TokenId:   101,
			TokenName: "prod-key",
			CreatedAt: dayTwoBucket,
			Quota:     90,
			TokenUsed: 55,
			Count:     1,
		},
	}, rows)
}

func TestGetTokenUsageDailyBucketsByRequestedLocalDayBoundary(t *testing.T) {
	truncateTables(t)

	location := time.FixedZone("CST", 8*3600)
	rangeStart := time.Date(2026, 6, 1, 0, 0, 0, 0, location).Unix()
	rangeEnd := time.Date(2026, 6, 1, 23, 59, 59, 0, location).Unix()
	earlyMorning := time.Date(2026, 6, 1, 0, 30, 0, 0, location).Unix()
	lateNight := time.Date(2026, 6, 1, 23, 30, 0, 0, location).Unix()

	logs := []*Log{
		{
			UserId:           1,
			Username:         "alice",
			CreatedAt:        earlyMorning,
			Type:             LogTypeConsume,
			TokenId:          101,
			TokenName:        "prod-key",
			Quota:            250,
			PromptTokens:     100,
			CompletionTokens: 30,
		},
		{
			UserId:           1,
			Username:         "alice",
			CreatedAt:        lateNight,
			Type:             LogTypeConsume,
			TokenId:          101,
			TokenName:        "prod-key",
			Quota:            75,
			PromptTokens:     20,
			CompletionTokens: 10,
		},
	}
	require.NoError(t, LOG_DB.Create(&logs).Error)

	rows, err := GetTokenUsageDaily(1, "", rangeStart, rangeEnd)
	require.NoError(t, err)

	require.Equal(t, []TokenUsageDaily{
		{
			TokenId:   101,
			TokenName: "prod-key",
			CreatedAt: rangeStart,
			Quota:     325,
			TokenUsed: 160,
			Count:     2,
		},
	}, rows)
}

func TestGetTokenUsageDailyCanFilterByUsernameForAdmin(t *testing.T) {
	truncateTables(t)

	createdAt := time.Date(2026, 1, 2, 10, 30, 0, 0, time.UTC).Unix()
	bucket := createdAt - createdAt%86400

	logs := []*Log{
		{
			UserId:           1,
			Username:         "alice",
			CreatedAt:        createdAt,
			Type:             LogTypeConsume,
			TokenId:          101,
			TokenName:        "prod-key",
			Quota:            250,
			PromptTokens:     100,
			CompletionTokens: 30,
		},
		{
			UserId:           2,
			Username:         "bob",
			CreatedAt:        createdAt,
			Type:             LogTypeConsume,
			TokenId:          202,
			TokenName:        "bob-key",
			Quota:            84,
			PromptTokens:     40,
			CompletionTokens: 2,
		},
	}
	require.NoError(t, LOG_DB.Create(&logs).Error)

	rows, err := GetTokenUsageDaily(0, "bob", bucket, bucket+86399)
	require.NoError(t, err)

	require.Equal(t, []TokenUsageDaily{
		{
			TokenId:   202,
			TokenName: "bob-key",
			CreatedAt: bucket,
			Quota:     84,
			TokenUsed: 42,
			Count:     1,
		},
	}, rows)
}
