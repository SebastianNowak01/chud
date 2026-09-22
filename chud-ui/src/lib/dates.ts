// Dates are handled as local "YYYY-MM-DD" strings, the format the API uses for plan days.

// The UI is English, so dates are too, whatever the browser's language.
export const LOCALE = 'en-GB'

const pad = (n: number) => String(n).padStart(2, '0')

export const toDateString = (date: Date): string =>
  `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`

// The API returns DATE columns as midnight UTC timestamps; the date part is what matters.
export const apiDate = (value: string): string => value.slice(0, 10)

// Parses a "YYYY-MM-DD" string as a local date (at noon, so DST shifts never change the day).
export const parseDate = (value: string): Date => {
  const [year, month, day] = value.split('-').map(Number)
  return new Date(year, month - 1, day, 12)
}

export const addDays = (date: Date, days: number): Date => {
  const result = new Date(date)
  result.setDate(result.getDate() + days)
  return result
}

// Value for <input type="datetime-local">, in local time.
export const toDateTimeLocal = (date: Date): string =>
  `${toDateString(date)}T${pad(date.getHours())}:${pad(date.getMinutes())}`

export const formatDateTime = (iso: string): string =>
  new Date(iso).toLocaleString(LOCALE, { dateStyle: 'medium', timeStyle: 'short' })

export const formatMonth = (dateString: string): string =>
  parseDate(dateString).toLocaleDateString(LOCALE, { month: 'short' })

export const formatDate = (dateString: string): string =>
  parseDate(dateString).toLocaleDateString(LOCALE, { weekday: 'short', day: 'numeric', month: 'short' })
