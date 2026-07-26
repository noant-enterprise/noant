export function arr<T>(v: T[] | null | undefined): T[] {
  return v ?? []
}

export function num(v: number | null | undefined): number {
  return v ?? 0
}
