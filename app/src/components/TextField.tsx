import Ionicons from '@expo/vector-icons/Ionicons';
import { forwardRef, useState, type ComponentProps } from 'react';
import { Pressable, StyleSheet, TextInput, View, type TextInputProps } from 'react-native';

import { colors, layout, radius, spacing, typography } from '@/theme';

import { AppText } from './AppText';

type TextFieldProps = TextInputProps & {
  label: string;
  icon?: ComponentProps<typeof Ionicons>['name'];
  error?: string | null;
};

export const TextField = forwardRef<TextInput, TextFieldProps>(function TextField(
  { label, icon, error, secureTextEntry, style, ...rest },
  ref,
) {
  const [isFocused, setIsFocused] = useState(false);
  const [isRevealed, setIsRevealed] = useState(false);
  const borderColor = error ? colors.danger : isFocused ? colors.accent : colors.border;
  const canReveal = Boolean(secureTextEntry);

  return (
    <View style={styles.container}>
      <View style={[styles.field, { borderColor }]}>
        {icon ? <Ionicons name={icon} size={18} color={isFocused ? colors.text : colors.textMuted} /> : null}
        <TextInput
          ref={ref}
          accessibilityLabel={label}
          placeholder={label}
          placeholderTextColor={colors.textMuted}
          secureTextEntry={canReveal && !isRevealed}
          onFocus={() => setIsFocused(true)}
          onBlur={() => setIsFocused(false)}
          style={[styles.input, style]}
          {...rest}
        />
        {canReveal ? (
          <Pressable
            accessibilityRole="button"
            accessibilityLabel={isRevealed ? 'Ocultar contraseña' : 'Mostrar contraseña'}
            onPress={() => setIsRevealed((revealed) => !revealed)}
            hitSlop={10}
          >
            <Ionicons name={isRevealed ? 'eye-off-outline' : 'eye-outline'} size={18} color={colors.textMuted} />
          </Pressable>
        ) : null}
      </View>
      {error ? (
        <AppText variant="caption" color="danger" accessibilityLiveRegion="polite">
          {error}
        </AppText>
      ) : null}
    </View>
  );
});

const styles = StyleSheet.create({
  container: { gap: spacing.xs },
  field: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.md,
    minHeight: layout.minTouchTarget + 6,
    paddingHorizontal: spacing.lg,
    borderWidth: 1,
    borderRadius: radius.md,
    backgroundColor: colors.surfaceRaised,
  },
  input: {
    ...typography.body,
    flex: 1,
    alignSelf: 'stretch',
    color: colors.text,
    outlineStyle: 'none',
  } as object,
});
