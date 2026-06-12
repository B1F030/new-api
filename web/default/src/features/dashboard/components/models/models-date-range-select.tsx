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
import { useMemo, useState } from 'react'
import { CalendarRange, RotateCcw, Search } from 'lucide-react'
import type { DateRange } from 'react-day-picker'
import { enUS, fr, ja, ru, vi, zhCN } from 'react-day-picker/locale'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import dayjs from '@/lib/dayjs'
import { getEndOfDay, getStartOfDay, type TimeGranularity } from '@/lib/time'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Label } from '@/components/ui/label'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  MAX_DASHBOARD_RANGE_DAYS,
  TIME_GRANULARITY_OPTIONS,
  TIME_RANGE_PRESETS,
} from '@/features/dashboard/constants'
import {
  buildDefaultDashboardFilters,
  getDashboardDateRange,
  getDashboardMonthRanges,
  type DashboardMonthRange,
} from '@/features/dashboard/lib'
import type {
  DashboardChartPreferences,
  DashboardFilters,
} from '@/features/dashboard/types'

const calendarLocales = {
  en: enUS,
  zh: zhCN,
  fr,
  ru,
  ja,
  vi,
} as const

interface ModelsDateRangeSelectProps {
  filters: DashboardFilters
  preferences: DashboardChartPreferences
  onFilterChange: (filters: DashboardFilters) => void
}

function getFilterRange(filters: DashboardFilters): DateRange {
  return {
    from: filters.start_timestamp,
    to: filters.end_timestamp,
  }
}

function getMinSelectableDate() {
  const min = getStartOfDay(new Date())
  min.setDate(min.getDate() - (MAX_DASHBOARD_RANGE_DAYS - 1))
  return min
}

function normalizeRange(range: DateRange): DateRange {
  if (!range.from) return range
  const to = range.to ?? range.from
  const start = getStartOfDay(range.from)
  const endOfSelection = getEndOfDay(to)
  const now = new Date()
  return {
    from: start,
    to: endOfSelection > now ? now : endOfSelection,
  }
}

function getInclusiveDayCount(range: DateRange) {
  if (!range.from) return 0
  const to = range.to ?? range.from
  const start = getStartOfDay(range.from).getTime()
  const end = getStartOfDay(to).getTime()
  return Math.floor((end - start) / (24 * 3600 * 1000)) + 1
}

export function ModelsDateRangeSelect(props: ModelsDateRangeSelectProps) {
  const { t, i18n } = useTranslation()
  const calendarLocale =
    calendarLocales[i18n.language as keyof typeof calendarLocales] ?? enUS
  const [open, setOpen] = useState(false)
  const [draftRange, setDraftRange] = useState<DateRange>(() =>
    getFilterRange(props.filters)
  )
  const [draftGranularity, setDraftGranularity] = useState<TimeGranularity>(
    props.filters.time_granularity ?? props.preferences.defaultTimeGranularity
  )
  const minSelectableDate = useMemo(() => getMinSelectableDate(), [])
  const monthRanges = useMemo(() => getDashboardMonthRanges(), [])

  const resetDraftFromFilters = () => {
    setDraftRange(getFilterRange(props.filters))
    setDraftGranularity(
      props.filters.time_granularity ?? props.preferences.defaultTimeGranularity
    )
  }

  const handleOpenChange = (nextOpen: boolean) => {
    if (nextOpen) resetDraftFromFilters()
    setOpen(nextOpen)
  }

  const handleQuickRange = (days: number) => {
    const { start, end } = getDashboardDateRange(days)
    setDraftRange({ from: start, to: end })
  }

  const handleMonthRange = (range: DashboardMonthRange) => {
    setDraftRange({
      from: new Date(range.start),
      to: new Date(range.end),
    })
  }

  const handleReset = () => {
    const nextFilters = buildDefaultDashboardFilters(props.preferences)
    const range = getFilterRange(nextFilters)
    setDraftRange(range)
    setDraftGranularity(
      nextFilters.time_granularity ?? props.preferences.defaultTimeGranularity
    )
    props.onFilterChange({
      ...nextFilters,
      username: props.filters.username,
    })
    setOpen(false)
  }

  const handleApply = () => {
    if (!draftRange.from) {
      toast.error(t('Select a start and end date'))
      return
    }

    if (getInclusiveDayCount(draftRange) > MAX_DASHBOARD_RANGE_DAYS) {
      toast.error(t('The selected range cannot exceed 90 days'))
      return
    }

    const normalized = normalizeRange(draftRange)
    props.onFilterChange({
      ...props.filters,
      start_timestamp: normalized.from,
      end_timestamp: normalized.to,
      time_granularity: draftGranularity,
    })
    setOpen(false)
  }

  const selectedPreset = TIME_RANGE_PRESETS.find(
    (preset) => getInclusiveDayCount(draftRange) === preset.days
  )?.days
  const selectedMonth = monthRanges.find((range) => {
    if (!draftRange.from) return false
    const draftTo = draftRange.to ?? draftRange.from
    return (
      getStartOfDay(draftRange.from).getTime() ===
        getStartOfDay(range.start).getTime() &&
      getStartOfDay(draftTo).getTime() === getStartOfDay(range.end).getTime()
    )
  })?.label
  const currentRange = getFilterRange(props.filters)
  const currentLabel =
    currentRange.from && currentRange.to
      ? `${dayjs(currentRange.from).format('YYYY-MM-DD')} - ${dayjs(
          currentRange.to
        ).format('YYYY-MM-DD')}`
      : t('Date Range')

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger
        render={
          <Button
            variant='outline'
            size='sm'
            className='max-w-full justify-start gap-2 font-normal'
          />
        }
      >
        <CalendarRange className='size-4 shrink-0' />
        <span className='truncate'>{currentLabel}</span>
      </PopoverTrigger>
      <PopoverContent
        align='end'
        className='w-[min(calc(100vw-2rem),620px)] p-3'
      >
        <div className='grid gap-3'>
          <div className='grid gap-2'>
            <Label>{t('Quick Range')}</Label>
            <div className='grid grid-cols-2 gap-2 sm:grid-cols-5'>
              {TIME_RANGE_PRESETS.map((range) => (
                <Button
                  key={range.days}
                  type='button'
                  size='sm'
                  variant={
                    selectedPreset === range.days ? 'default' : 'outline'
                  }
                  onClick={() => handleQuickRange(range.days)}
                  className={cn(
                    selectedPreset === range.days &&
                      'ring-ring ring-2 ring-offset-2'
                  )}
                >
                  {t(range.label)}
                </Button>
              ))}
            </div>
          </div>

          {monthRanges.length > 0 && (
            <div className='grid gap-2'>
              <Label>{t('Monthly Range')}</Label>
              <div className='grid grid-cols-2 gap-2 sm:grid-cols-4'>
                {monthRanges.map((range) => (
                  <Button
                    key={range.label}
                    type='button'
                    size='sm'
                    variant={
                      selectedMonth === range.label ? 'default' : 'outline'
                    }
                    onClick={() => handleMonthRange(range)}
                    className={cn(
                      selectedMonth === range.label &&
                        'ring-ring ring-2 ring-offset-2'
                    )}
                  >
                    {range.label}
                  </Button>
                ))}
              </div>
            </div>
          )}

          <div className='grid gap-2'>
            <Label>{t('Custom Time Range')}</Label>
            <Calendar
              mode='range'
              selected={draftRange}
              onSelect={(range) => setDraftRange(range ?? { from: undefined })}
              numberOfMonths={2}
              captionLayout='dropdown'
              locale={calendarLocale}
              disabled={(date: Date) =>
                date > new Date() || date < minSelectableDate
              }
              className='mx-auto'
            />
          </div>

          <div className='grid gap-2'>
            <Label htmlFor='dashboard-time-granularity'>
              {t('Time Granularity')}
            </Label>
            <Select
              items={[
                ...TIME_GRANULARITY_OPTIONS.map((option) => ({
                  value: option.value,
                  label: t(option.label),
                })),
              ]}
              value={draftGranularity}
              onValueChange={(value) =>
                setDraftGranularity(value as TimeGranularity)
              }
            >
              <SelectTrigger id='dashboard-time-granularity'>
                <SelectValue placeholder={t('Select time granularity')} />
              </SelectTrigger>
              <SelectContent alignItemWithTrigger={false}>
                <SelectGroup>
                  {TIME_GRANULARITY_OPTIONS.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {t(option.label)}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>

          <div className='grid grid-cols-2 gap-2 sm:flex sm:justify-end'>
            <Button type='button' variant='outline' onClick={handleReset}>
              <RotateCcw className='mr-2 size-4' />
              {t('Reset')}
            </Button>
            <Button type='button' onClick={handleApply}>
              <Search className='mr-2 size-4' />
              {t('Apply')}
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
}
