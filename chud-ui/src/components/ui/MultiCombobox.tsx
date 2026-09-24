import { useEffect, useId, useRef, useState, type KeyboardEvent } from 'react'
import { DrawablyInput } from 'drawably/react'
import { Card } from '@/components/ui/Card'

export interface ComboOption {
  id: string
  label: string
  color?: string
}

interface MultiComboboxProps {
  label: string
  placeholder: string
  emptyText: string
  options: ComboOption[]
  selected: string[]
  onChange: (selected: string[]) => void
}

const normalize = (text: string) => text.toLocaleLowerCase('pl').normalize('NFD').replace(/\p{Diacritic}/gu, '')

export function MultiCombobox({ label, placeholder, emptyText, options, selected, onChange }: MultiComboboxProps) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [activeIndex, setActiveIndex] = useState(0)
  const root = useRef<HTMLDivElement>(null)
  const input = useRef<HTMLInputElement>(null)
  const listId = useId()
  const inputId = useId()

  const matches = options.filter((o) => normalize(o.label).includes(normalize(query)))
  const chosen = selected.map((id) => options.find((o) => o.id === id)).filter((o) => o !== undefined)
  const optionId = (index: number) => `${listId}-${index}`

  const close = () => {
    setOpen(false)
    setQuery('')
  }

  useEffect(() => {
    if (!open) return
    const closeOutside = (e: PointerEvent) => {
      if (!root.current?.contains(e.target as Node)) {
        setOpen(false)
        setQuery('')
      }
    }
    document.addEventListener('pointerdown', closeOutside)
    return () => document.removeEventListener('pointerdown', closeOutside)
  }, [open])

  const toggle = (id: string) =>
    onChange(selected.includes(id) ? selected.filter((s) => s !== id) : [...selected, id])

  const show = (index: number) => {
    setOpen(true)
    setActiveIndex(Math.max(0, Math.min(index, matches.length - 1)))
  }

  const onKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      show(open ? activeIndex + 1 : 0)
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      show(activeIndex - 1)
    } else if (e.key === 'Enter' && open && matches[activeIndex]) {
      e.preventDefault()
      toggle(matches[activeIndex].id)
    } else if (e.key === 'Escape' && open) {
      e.preventDefault()
      close()
    } else if (e.key === 'Backspace' && query === '' && selected.length > 0) {
      onChange(selected.slice(0, -1))
    }
  }

  return (
    <div
      ref={root}
      className="flex flex-col gap-1"
      onBlur={(e) => {
        if (!root.current?.contains(e.relatedTarget as Node)) close()
      }}
    >
      <label htmlFor={inputId} className="text-[14px] text-muted">
        {label}
      </label>
      <div className="relative">
        <DrawablyInput
          ref={input}
          id={inputId}
          role="combobox"
          aria-expanded={open}
          aria-controls={listId}
          aria-autocomplete="list"
          aria-activedescendant={open && matches[activeIndex] ? optionId(activeIndex) : undefined}
          autoComplete="off"
          className="w-full [&_input]:w-full [&_input]:cursor-pointer [&_input]:pr-8"
          placeholder={chosen.length > 0 ? 'Dodaj…' : placeholder}
          value={query}
          onChange={(e) => {
            setQuery(e.target.value)
            show(0)
          }}
          onFocus={() => setOpen(true)}
          onClick={() => setOpen(true)}
          onKeyDown={onKeyDown}
        />
        <span
          aria-hidden="true"
          className={`pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-[12px] text-muted transition-transform duration-150 ${open ? 'rotate-180' : ''}`}
        >
          ▼
        </span>

        {open && (
          <div className="absolute top-full right-0 left-0 z-20 mt-1">
            <Card filled size="sm">
              <ul id={listId} role="listbox" aria-multiselectable="true" aria-label={label} className="m-0 max-h-64 list-none overflow-y-auto overscroll-contain p-0">
                {matches.length === 0 && <li className="px-2.5 py-2 text-[14px] text-muted">{emptyText}</li>}
                {matches.map((option, index) => {
                  const isSelected = selected.includes(option.id)
                  return (
                    <li
                      key={option.id}
                      id={optionId(index)}
                      role="option"
                      aria-selected={isSelected}
                      className={`flex min-h-10 cursor-pointer items-center gap-2 rounded-md px-2.5 text-[15px] ${index === activeIndex ? 'bg-ink/8' : ''}`}
                      onPointerEnter={() => setActiveIndex(index)}
                      onPointerDown={(e) => e.preventDefault()}
                      onClick={() => {
                        toggle(option.id)
                        input.current?.focus()
                      }}
                    >
                      <span
                        aria-hidden="true"
                        className={`grid size-4 flex-none place-items-center rounded-[4px] border-[1.5px] border-ink text-[11px] leading-none ${isSelected ? 'bg-ink text-paper' : ''}`}
                      >
                        {isSelected ? '✓' : ''}
                      </span>
                      {option.color && <span className="size-2.5 flex-none rounded-full" style={{ background: option.color }} />}
                      <span style={option.color ? { color: option.color } : undefined} className={option.color ? 'font-semibold' : ''}>
                        {option.label}
                      </span>
                    </li>
                  )
                })}
              </ul>
            </Card>
          </div>
        )}
      </div>

      {chosen.length > 0 && (
        <ul className="m-0 flex list-none flex-wrap gap-1.5 p-0" aria-label={`Wybrane: ${label.toLowerCase()}`}>
          {chosen.map((option) => (
            <li key={option.id}>
              <button
                type="button"
                className="inline-flex cursor-pointer items-center gap-1.5 rounded-full border border-rule bg-transparent px-2.5 py-0.5 text-[13px] hover:border-ink"
                aria-label={`Usuń ${option.label} z filtra`}
                onClick={() => toggle(option.id)}
              >
                {option.color && <span className="size-2 rounded-full" style={{ background: option.color }} />}
                <span style={option.color ? { color: option.color } : undefined} className="font-semibold">
                  {option.label}
                </span>
                <span className="text-muted" aria-hidden="true">
                  ×
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
