import type { IVChart } from '@visactor/vchart'

type LegendValue = string | number

type LegendDatum = {
  key?: LegendValue
  label?: LegendValue
  originalKey?: LegendValue
}

type LegendClickEvent = {
  value?: LegendValue[]
  event?: {
    detail?: {
      currentSelected?: LegendValue[]
      item?: {
        label?: LegendValue
      }
    }
  }
}

type LegendChart = Pick<
  IVChart,
  | 'getLegendDataByIndex'
  | 'getLegendSelectedDataByIndex'
  | 'setLegendSelectedDataByIndex'
>

function getSelectedValues(event: unknown): LegendValue[] {
  const legendEvent = event as LegendClickEvent
  const selected =
    legendEvent.event?.detail?.currentSelected ?? legendEvent.value
  return Array.isArray(selected) ? selected : []
}

function getAllLegendValues(chart: LegendChart): LegendValue[] {
  return chart
    .getLegendDataByIndex(0)
    .map((item) => {
      const datum = item as LegendDatum
      return datum.key ?? datum.label ?? datum.originalKey
    })
    .filter((value): value is LegendValue => value !== undefined)
}

function sameSelection(left: LegendValue[] | null, right: LegendValue[]) {
  if (!left || left.length !== right.length) return false
  return left.every((value, index) => value === right[index])
}

export function createToggleLegendSelectionHandler() {
  let previousSelected: LegendValue[] | null = null

  const handler = (chart: LegendChart | null, event: unknown) => {
    if (!chart) return

    const selected = getSelectedValues(event)
    if (selected.length !== 1) {
      previousSelected = selected
      return
    }

    if (sameSelection(previousSelected, selected)) {
      const allValues = getAllLegendValues(chart)
      chart.setLegendSelectedDataByIndex(0, allValues)
      previousSelected = allValues
      return
    }

    previousSelected = selected
  }

  handler.reset = (chart?: LegendChart | null) => {
    previousSelected = chart?.getLegendSelectedDataByIndex(0) ?? null
  }

  return handler
}
