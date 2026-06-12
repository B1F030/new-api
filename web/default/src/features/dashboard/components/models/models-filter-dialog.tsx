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
import { useState } from 'react'
import { Filter, RotateCcw, Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Dialog } from '@/components/dialog'
import { cleanFilters } from '@/features/dashboard/lib'
import type { DashboardFilters } from '@/features/dashboard/types'

interface ModelsFilterProps {
  filters: DashboardFilters
  onFilterChange: (filters: DashboardFilters) => void
}

export function ModelsFilter(props: ModelsFilterProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [username, setUsername] = useState(props.filters.username ?? '')

  const handleOpenChange = (nextOpen: boolean) => {
    if (nextOpen) setUsername(props.filters.username ?? '')
    setOpen(nextOpen)
  }

  const handleApply = () => {
    props.onFilterChange(
      cleanFilters({
        ...props.filters,
        username,
      }) as DashboardFilters
    )
    setOpen(false)
  }

  const handleReset = () => {
    setUsername('')
    props.onFilterChange({
      ...props.filters,
      username: '',
    })
    setOpen(false)
  }

  return (
    <Dialog
      open={open}
      onOpenChange={handleOpenChange}
      trigger={
        <Button variant='outline' size='sm'>
          <Filter className='mr-2 h-4 w-4' />
          {t('Filter')}
        </Button>
      }
      title={t('Model Analytics Filters')}
      description={t('Filter the model analytics view by user.')}
      contentClassName='sm:max-w-md'
      contentHeight='auto'
      footerClassName='grid grid-cols-2 gap-2 sm:flex'
      footer={
        <>
          <Button onClick={handleReset} variant='outline' type='button'>
            <RotateCcw className='mr-2 h-4 w-4' />
            {t('Reset')}
          </Button>
          <Button onClick={handleApply} type='submit'>
            <Search className='mr-2 h-4 w-4' />
            {t('Apply Filters')}
          </Button>
        </>
      }
    >
      <div className='grid gap-2 py-2'>
        <Label htmlFor='username'>{t('Username')}</Label>
        <Input
          id='username'
          placeholder={t('Filter by username')}
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
      </div>
    </Dialog>
  )
}
