import { format, parseISO } from 'date-fns';
import { fromZonedTime } from 'date-fns-tz';

export default function formatTime(time: string | undefined): string {
  if (!time) {
    return 'Invalid Date'; // Handle the case when time is undefined or null
  }

  const { timeZone } = Intl.DateTimeFormat().resolvedOptions();
  const date = parseISO(time);
  const zonedDate = fromZonedTime(date, timeZone);

  const formattedDate = format(zonedDate, 'eee, dd MMM yyyy HH:mm:ss');

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
