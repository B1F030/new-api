import assert from 'node:assert/strict'
import { describe, test } from 'node:test'
import { getDashboardMonthRanges } from './filters'

describe('dashboard date filters', () => {
  test('builds month ranges with year-specific labels', () => {
    const now = new Date('2026-06-12T10:30:00+08:00')
    const ranges = getDashboardMonthRanges(now)

    const may = ranges.find((range) => range.label === '2026-05')
    const june = ranges.find((range) => range.label === '2026-06')

    assert.equal(may?.start.toISOString(), '2026-04-30T16:00:00.000Z')
    assert.equal(may?.end.toISOString(), '2026-05-31T15:59:59.999Z')
    assert.equal(june?.start.toISOString(), '2026-05-31T16:00:00.000Z')
    assert.equal(june?.end.toISOString(), now.toISOString())
  })
})
