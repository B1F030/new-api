import assert from 'node:assert/strict'
import { describe, test } from 'node:test'
import { TIME_RANGE_PRESETS } from '../constants'
import { processTokenUsageChartData } from './charts'

describe('dashboard chart data processing', () => {
  test('uses 30 days instead of 29 days in quick presets', () => {
    assert.equal(
      TIME_RANGE_PRESETS.some((preset) => preset.days === 30),
      true
    )
    assert.equal(
      TIME_RANGE_PRESETS.some((preset) => Number(preset.days) === 29),
      false
    )
  })

  test('fills API token cost days from the selected local start date', () => {
    const startTimestamp =
      new Date('2026-06-01T00:00:00+08:00').getTime() / 1000
    const endTimestamp = new Date('2026-06-12T23:59:59+08:00').getTime() / 1000
    const chartData = processTokenUsageChartData(
      [
        {
          token_id: 1,
          token_name: 'prod-key',
          created_at: new Date('2026-06-01T12:00:00+08:00').getTime() / 1000,
          quota: 500000,
          token_used: 100,
          count: 1,
        },
      ],
      (key) => key,
      undefined,
      startTimestamp,
      endTimestamp
    )

    const values = chartData.spec_token_usage.data[0].values as Array<{
      Time: string
    }>

    assert.equal(values[0].Time, '06-01')
    assert.equal(values.at(-1)?.Time, '06-12')
  })

  test('returns per-token total cost summaries', () => {
    const chartData = processTokenUsageChartData(
      [
        {
          token_id: 1,
          token_name: 'prod-key',
          created_at: new Date('2026-06-01T12:00:00+08:00').getTime() / 1000,
          quota: 500000,
          token_used: 100,
          count: 1,
        },
        {
          token_id: 2,
          token_name: 'batch-key',
          created_at: new Date('2026-06-01T12:00:00+08:00').getTime() / 1000,
          quota: 250000,
          token_used: 50,
          count: 1,
        },
      ],
      (key) => key
    )

    assert.deepEqual(
      chartData.tokenCostSummaries.map((item) => ({
        tokenName: item.tokenName,
        quota: item.quota,
      })),
      [
        { tokenName: 'prod-key', quota: 500000 },
        { tokenName: 'batch-key', quota: 250000 },
      ]
    )
  })

  test('keeps API tokens with the same name separate by token id', () => {
    const chartData = processTokenUsageChartData(
      [
        {
          token_id: 1,
          token_name: 'shared-key',
          created_at: new Date('2026-06-01T12:00:00+08:00').getTime() / 1000,
          quota: 500000,
          token_used: 100,
          count: 1,
        },
        {
          token_id: 2,
          token_name: 'shared-key',
          created_at: new Date('2026-06-01T12:00:00+08:00').getTime() / 1000,
          quota: 250000,
          token_used: 50,
          count: 1,
        },
      ],
      (key) => key
    )

    assert.deepEqual(
      chartData.tokenCostSummaries.map((item) => item.tokenName),
      ['shared-key (#1)', 'shared-key (#2)']
    )
  })
})
