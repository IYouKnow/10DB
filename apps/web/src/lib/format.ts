export function formatDateTime(value: string) {
  return new Date(value).toLocaleString();
}

export function classNames(...values: Array<string | false | null | undefined>) {
  return values.filter(Boolean).join(" ");
}
