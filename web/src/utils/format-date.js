import { format, parseISO } from 'date-fns';
import { fromZonedTime } from 'date-fns-tz';

export function formatTime(time) {
  const { timeZone } = Intl.DateTimeFormat().resolvedOptions().timeZone;
  const date = parseISO(time);
  const zonedDate = fromZonedTime(date, timeZone);

  const formattedDate = format(zonedDate, 'eee, dd MMM yyyy HH:mm:ss', { timeZone });

  // Attempt to append the timezone abbreviation
  // This may not always return a 3-letter abbreviation due to localization and browser differences
  const timeZoneAbbreviation = new Intl.DateTimeFormat('en-US', {
    timeZone,
    timeZoneName: 'short',
  })
    .format(zonedDate)
    .split(', ')[1];

  return `${formattedDate} ${timeZoneAbbreviation}`;
}
