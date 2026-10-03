import { Pressable, ScrollView, StyleSheet, View } from 'react-native';

import { colors, radius, spacing } from '@/theme';

import { AppText } from './AppText';

type ChipOption<T extends string> = { value: T; label: string };

type ChipGroupProps<T extends string> = {
  options: readonly ChipOption<T>[];
  selected: T;
  onSelect: (value: T) => void;
  accessibilityLabel: string;
  variant?: 'pill' | 'underline';
};

export function ChipGroup<T extends string>({
  options,
  selected,
  onSelect,
  accessibilityLabel,
  variant = 'pill',
}: ChipGroupProps<T>) {
  const isUnderline = variant === 'underline';

  return (
    <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={isUnderline ? styles.underlineGroup : styles.pillGroup}>
      <View style={isUnderline ? styles.underlineRow : styles.pillRow} accessibilityRole="tablist" accessibilityLabel={accessibilityLabel}>
        {options.map((option) => {
          const isSelected = option.value === selected;
          return (
            <Pressable
              key={option.value}
              accessibilityRole="tab"
              accessibilityState={{ selected: isSelected }}
              onPress={() => onSelect(option.value)}
              style={isUnderline ? [styles.underlineTab, isSelected && styles.underlineTabSelected] : [styles.pill, isSelected && styles.pillSelected]}
            >
              <AppText variant={isUnderline ? 'bodyStrong' : 'caption'} color={isSelected ? 'text' : 'textMuted'}>
                {option.label}
              </AppText>
            </Pressable>
          );
        })}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  pillGroup: { flexGrow: 1 },
  pillRow: { flexDirection: 'row', gap: spacing.sm },
  pill: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
    borderRadius: radius.pill,
    borderWidth: 1,
    borderColor: colors.border,
    backgroundColor: colors.surface,
  },
  pillSelected: { backgroundColor: colors.accent, borderColor: colors.accent },
  underlineGroup: { flexGrow: 1, borderBottomWidth: 1, borderBottomColor: colors.border },
  underlineRow: { flexDirection: 'row', gap: spacing.xl },
  underlineTab: { paddingVertical: spacing.md, borderBottomWidth: 2, borderBottomColor: 'transparent', marginBottom: -1 },
  underlineTabSelected: { borderBottomColor: colors.accent },
});
