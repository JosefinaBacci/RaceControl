import { useState } from 'react';
import { ScrollView, StyleSheet, View } from 'react-native';

import { spacing, colors, radius } from '@/theme';
import { AppText } from './AppText';
import { Button } from './Button';
import { TextField } from './TextField';

type SelectFieldProps = {
  options: readonly { value: string; label: string }[];
  selected: string;
  onSelect: (value: string) => void;
  accessibilityLabel: string;
};

export function SelectField({ options, selected, onSelect, accessibilityLabel }: SelectFieldProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [search, setSearch] = useState('');
  const current = options.find((option) => option.value === selected);
  const visible = options.filter((option) => option.label.toLocaleLowerCase().includes(search.trim().toLocaleLowerCase()));

  return (
    <View style={styles.container}>
      <Button label={current?.label ?? accessibilityLabel} variant="secondary"
        trailingIcon={isOpen ? 'chevron-up' : 'chevron-down'}
        accessibilityLabel={`${accessibilityLabel}: ${current?.label ?? 'sin selección'}`}
        accessibilityState={{ expanded: isOpen }} onPress={() => { setSearch(''); setIsOpen(!isOpen); }} />
      {isOpen ? (
        <View style={styles.menu}>
          <TextField label="Buscar" value={search} onChangeText={setSearch} autoCorrect={false} />
          <ScrollView style={styles.options} nestedScrollEnabled keyboardShouldPersistTaps="handled">
            {visible.map((option) => (
              <Button key={option.value} label={option.label} variant={option.value === selected ? 'primary' : 'ghost'}
                accessibilityState={{ selected: option.value === selected }}
                onPress={() => { onSelect(option.value); setIsOpen(false); setSearch(''); }} />
            ))}
            {visible.length === 0 ? <AppText color="textMuted">Sin coincidencias.</AppText> : null}
          </ScrollView>
        </View>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: spacing.sm },
  menu: { gap: spacing.sm, padding: spacing.sm, borderWidth: 1, borderColor: colors.border, borderRadius: radius.md },
  options: { maxHeight: 260 },
});
