export interface OptimisticItem {
  id: string
  _optimistic?: boolean
  _originalValue?: any
  [key: string]: any
}

export function optimisticUpdate<T extends OptimisticItem>(
  data: T[],
  update: Partial<T> & { id: string }
): T[] {
  return data.map(item => {
    if (item.id === update.id) {
      return {
        ...item,
        ...update,
        _optimistic: true,
        _originalValue: item._originalValue || { ...item },
      }
    }
    return item
  })
}

export function rollbackUpdate<T extends OptimisticItem>(
  optimisticData: T[],
  serverData: T[]
): T[] {
  return optimisticData.map(item => {
    if (item._optimistic) {
      const serverItem = serverData.find(s => s.id === item.id)
      if (serverItem) {
        const { _optimistic, _originalValue, ...cleanItem } = serverItem
        return cleanItem as T
      }
      // Fallback to original value if server data not available
      if (item._originalValue) {
        return item._originalValue
      }
    }
    return item
  })
}

export function confirmUpdate<T extends OptimisticItem>(
  data: T[],
  confirmedUpdate: Partial<T> & { id: string }
): T[] {
  return data.map(item => {
    if (item.id === confirmedUpdate.id && item._optimistic) {
      const { _optimistic, _originalValue, ...cleanItem } = {
        ...item,
        ...confirmedUpdate,
      }
      return cleanItem as T
    }
    return item
  })
}