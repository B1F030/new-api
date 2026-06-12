import assert from 'node:assert/strict'
import { describe, test } from 'node:test'
import { createToggleLegendSelectionHandler } from './vchart-legend'

describe('dashboard VChart legend interaction', () => {
  test('clicking the same single-selected legend again restores all series', () => {
    const selectedCalls: Array<Array<string | number>> = []
    const chart = {
      getLegendDataByIndex: () => [{ key: 'A' }, { key: 'B' }],
      getLegendSelectedDataByIndex: () => ['A', 'B'],
      setLegendSelectedDataByIndex: (
        _index: number,
        selected: Array<string | number>
      ) => {
        selectedCalls.push(selected)
      },
    }
    const handler = createToggleLegendSelectionHandler()

    handler(chart, {
      value: ['A'],
      event: { detail: { item: { label: 'A' } } },
    })
    handler(chart, {
      value: ['A'],
      event: { detail: { item: { label: 'A' } } },
    })

    assert.deepEqual(selectedCalls, [['A', 'B']])
  })
})
