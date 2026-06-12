/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useEffect, useMemo, useRef, useState } from 'react'
import { VChart } from '@visactor/react-vchart'
import type { IVChart } from '@visactor/vchart'
import { KeyRound } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { computeTimeRange } from '@/lib/time'
import { VCHART_OPTION } from '@/lib/vchart'
import { useThemeCustomization } from '@/context/theme-customization-provider'
import { useTheme } from '@/context/theme-provider'
import { getTokenUsageDailyData } from '@/features/dashboard/api'
import { DEFAULT_TIME_GRANULARITY } from '@/features/dashboard/constants'
import {
  buildQueryParams,
  getDefaultDays,
  processTokenUsageChartData,
} from '@/features/dashboard/lib'
import type {
  DashboardFilters,
  TokenUsageDataItem,
} from '@/features/dashboard/types'
import { createToggleLegendSelectionHandler } from '../../lib/vchart-legend'

let themeManagerPromise: Promise<
  (typeof import('@visactor/vchart'))['ThemeManager']
> | null = null

interface TokenUsageChartProps {
  filters?: DashboardFilters
  isAdmin?: boolean
}

export function TokenUsageChart(props: TokenUsageChartProps) {
  const { t } = useTranslation()
  const { resolvedTheme } = useTheme()
  const { customization } = useThemeCustomization()
  const [data, setData] = useState<TokenUsageDataItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [themeReady, setThemeReady] = useState(false)
  const chartRef = useRef<IVChart | null>(null)
  const themeManagerRef = useRef<
    (typeof import('@visactor/vchart'))['ThemeManager'] | null
  >(null)
  const handleLegendSelection = useMemo(
    () => createToggleLegendSelectionHandler(),
    []
  )
  const granularity =
    props.filters?.time_granularity ?? DEFAULT_TIME_GRANULARITY

  const timeRange = useMemo(
    () =>
      computeTimeRange(
        getDefaultDays(granularity),
        props.filters?.start_timestamp,
        props.filters?.end_timestamp
      ),
    [granularity, props.filters?.start_timestamp, props.filters?.end_timestamp]
  )

  useEffect(() => {
    const updateTheme = async () => {
      setThemeReady(false)

      if (!themeManagerPromise) {
        themeManagerPromise = import('@visactor/vchart').then(
          (m) => m.ThemeManager
        )
      }

      const ThemeManager = await themeManagerPromise
      themeManagerRef.current = ThemeManager
      ThemeManager.setCurrentTheme(resolvedTheme === 'dark' ? 'dark' : 'light')
      setThemeReady(true)
    }

    updateTheme()
  }, [resolvedTheme])

  useEffect(() => {
    let cancelled = false
    const queryTimeRange = {
      start_timestamp: timeRange.start_timestamp,
      end_timestamp: timeRange.end_timestamp,
    }

    void Promise.resolve().then(() => {
      if (cancelled) return
      setLoading(true)
      setError(false)
    })

    getTokenUsageDailyData(
      buildQueryParams(queryTimeRange, {
        time_granularity: granularity,
        username: props.filters?.username,
      }),
      props.isAdmin
    )
      .then((res) => {
        if (cancelled) return
        if (!res?.success) {
          setData([])
          setError(true)
          return
        }
        setData(res.data || [])
      })
      .catch(() => {
        if (cancelled) return
        setData([])
        setError(true)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [
    granularity,
    props.filters?.username,
    props.isAdmin,
    timeRange.start_timestamp,
    timeRange.end_timestamp,
  ])

  const chartData = useMemo(
    () =>
      processTokenUsageChartData(
        loading || error ? [] : data,
        t,
        customization.preset,
        timeRange.start_timestamp,
        timeRange.end_timestamp
      ),
    [
      data,
      error,
      loading,
      t,
      customization.preset,
      timeRange.start_timestamp,
      timeRange.end_timestamp,
    ]
  )
  const spec = chartData.spec_token_usage
  const specType = typeof spec?.type === 'string' ? spec.type : 'token-usage'
  const chartKey = [
    specType,
    loading ? 'loading' : 'ready',
    error ? 'error' : 'ok',
    data.length,
    timeRange.start_timestamp,
    timeRange.end_timestamp,
    resolvedTheme,
    customization.preset,
  ].join('-')

  return (
    <div className='overflow-hidden rounded-lg border'>
      <div className='flex w-full flex-col gap-1.5 border-b px-3 py-2 sm:gap-3 sm:px-5 sm:py-3 lg:flex-row lg:items-center lg:justify-between'>
        <div className='flex items-center gap-2'>
          <KeyRound className='text-muted-foreground/60 size-4' />
          <div className='text-sm font-semibold'>{t('API Token Cost')}</div>
          <span className='text-muted-foreground text-xs'>
            {t('Cost')}: {chartData.totalCostDisplay}
          </span>
        </div>
        <div className='text-muted-foreground text-xs'>
          {t('Daily cost by API key')}
        </div>
      </div>

      <div className='h-[300px] p-1.5 sm:h-96 sm:p-2'>
        {themeReady && spec && (
          <VChart
            key={chartKey}
            spec={{
              ...spec,
              theme: resolvedTheme === 'dark' ? 'dark' : 'light',
              background: 'transparent',
            }}
            option={VCHART_OPTION}
            onReady={(instance: IVChart) => {
              chartRef.current = instance
              handleLegendSelection.reset(instance)
            }}
            onLegendItemClick={(event: unknown) =>
              handleLegendSelection(chartRef.current, event)
            }
          />
        )}
      </div>

      {chartData.tokenCostSummaries.length > 0 && (
        <div className='border-t px-3 py-3 sm:px-5'>
          <div className='mb-2 flex items-center justify-between gap-2'>
            <div className='text-muted-foreground text-xs font-medium'>
              {t('Token Cost Summary')}
            </div>
            <div className='text-muted-foreground text-xs'>
              {t('Total Cost')}: {chartData.totalCostDisplay}
            </div>
          </div>
          <div className='grid max-h-40 gap-2 overflow-y-auto pr-1 sm:grid-cols-2 xl:grid-cols-3'>
            {chartData.tokenCostSummaries.map((item) => (
              <div
                key={item.tokenName}
                className='bg-muted/30 flex min-w-0 items-center justify-between gap-2 rounded-md border px-2.5 py-2'
              >
                <span className='truncate text-xs' title={item.tokenName}>
                  {item.tokenName}
                </span>
                <span className='text-xs font-semibold tabular-nums'>
                  {item.costDisplay}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
