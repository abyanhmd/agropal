// Day-only dates for <input type="date">, as YYYY-MM-DD. Stored dates are shown in UTC (see
// formatOrderDate), so a day maps to UTC midnight and back.

export function todayIso(): string {
    const now = new Date();
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
}

export const toDateInput = (date: Date) => date.toISOString().slice(0, 10);

export const fromDateInput = (day: string) => new Date(`${day}T00:00:00Z`);

// For history entries, where the newest one is "current": today means right now, so an entry
// added today sorts after any added earlier the same day
export const entryTimestamp = (day: string) => (day === todayIso() ? new Date() : fromDateInput(day));
