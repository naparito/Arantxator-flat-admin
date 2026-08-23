type IconProps = { size?: number }

export function IconCasa({ size = 19 }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none">
      <rect x="3.5" y="3" width="8" height="14" rx="1" stroke="currentColor" strokeWidth="1.6" />
      <path d="M11.5 8h5v9h-5" stroke="currentColor" strokeWidth="1.6" />
      <path
        d="M6 6h1M6 9h1M6 12h1M8.5 6h1M8.5 9h1M8.5 12h1"
        stroke="currentColor"
        strokeWidth="1.3"
        strokeLinecap="round"
      />
    </svg>
  )
}

export function IconInquilinos({ size = 19 }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none">
      <circle cx="10" cy="6.5" r="3" stroke="currentColor" strokeWidth="1.6" />
      <path d="M4 17c0-3.3 2.7-6 6-6s6 2.7 6 6" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
    </svg>
  )
}

export function IconContratos({ size = 19 }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none">
      <path d="M5.5 2.5h6l3 3v12h-9v-15Z" stroke="currentColor" strokeWidth="1.6" strokeLinejoin="round" />
      <path d="M11.5 2.5v3h3" stroke="currentColor" strokeWidth="1.6" strokeLinejoin="round" />
      <path d="M7.5 10h5M7.5 12.5h5M7.5 15h3" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
    </svg>
  )
}

export function IconGastos({ size = 19 }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none">
      <rect x="2.5" y="5.5" width="15" height="10.5" rx="2" stroke="currentColor" strokeWidth="1.6" />
      <path d="M2.5 8.5h15" stroke="currentColor" strokeWidth="1.6" />
      <circle cx="14" cy="12" r="1.1" fill="currentColor" />
    </svg>
  )
}

export function IconResumen({ size = 19 }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none">
      <path d="M3 9.5 10 3l7 6.5" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M5 8.5V17h10V8.5" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

export function IconPlus({ size = 15 }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none">
      <path d="M10 4v12M4 10h12" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    </svg>
  )
}

export function IconM2({ size = 13 }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none">
      <rect x="2.5" y="4" width="15" height="12" rx="1.5" stroke="currentColor" strokeWidth="1.5" />
    </svg>
  )
}

export function IconHabitacion({ size = 13 }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none">
      <path d="M3 15V6l7-3 7 3v9" stroke="currentColor" strokeWidth="1.5" />
    </svg>
  )
}

export function IconChevronRight({ size = 12 }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none">
      <path d="m7.5 4.5 6 5.5-6 5.5" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

export function IconSearch({ size = 15 }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none">
      <circle cx="9" cy="9" r="5.5" stroke="currentColor" strokeWidth="1.6" />
      <path d="m17 17-3.5-3.5" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
    </svg>
  )
}

export function IconDoc({ size = 16 }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none">
      <path d="M5.5 2.5h6l3 3v12h-9v-15Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" />
      <path d="M11.5 2.5v3h3" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" />
    </svg>
  )
}
