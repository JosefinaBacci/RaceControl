import { Linking, StyleSheet, View } from 'react-native';

import type { EventFilter } from '@/api/calendar';
import { Button } from '@/components';
import { calendarDownloadUrl } from '@/data/calendar';
import { spacing } from '@/theme';

// The download carries the same filter as the list on screen, so the file holds
// exactly the events the user is looking at.
export function CalendarDownloads({ filter }: { filter: EventFilter }) {
  const download = (format: 'ics' | 'pdf') => Linking.openURL(calendarDownloadUrl(format, filter));

  return (
    <View style={styles.downloads}>
      <Button label=".ics" compact variant="secondary" onPress={() => download('ics')} />
      <Button label="PDF" compact variant="secondary" onPress={() => download('pdf')} />
    </View>
  );
}

const styles = StyleSheet.create({
  downloads: { flexDirection: 'row', gap: spacing.sm },
});
