function formatIsoDate(
  utcIso: string,
  options: Intl.DateTimeFormatOptions,
): string {
  if (!utcIso) return '—';
  const date = new Date(utcIso);
  if (Number.isNaN(date.getTime())) return utcIso;
  return new Intl.DateTimeFormat(undefined, options).format(date);
}

export function formatLocalDateTime(utcIso: string): string {
  return formatIsoDate(utcIso, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function formatLocalDate(utcIso: string): string {
  return formatIsoDate(utcIso, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
  });
}
