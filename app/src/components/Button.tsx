import Ionicons from '@expo/vector-icons/Ionicons';
import type { ComponentProps } from 'react';
import { ActivityIndicator, Pressable, StyleSheet, type PressableProps } from 'react-native';

import { colors, layout, radius, spacing, type ColorName } from '@/theme';

import { AppText } from './AppText';

type ButtonVariant = 'primary' | 'secondary' | 'ghost';

type ButtonProps = Omit<PressableProps, 'children'> & {
  label: string;
  variant?: ButtonVariant;
  loading?: boolean;
  compact?: boolean;
  icon?: ComponentProps<typeof Ionicons>['name'];
  trailingIcon?: ComponentProps<typeof Ionicons>['name'];
};

const variantStyles: Record<ButtonVariant, { background: ColorName; pressed: ColorName; text: ColorName; border: ColorName }> = {
  primary: { background: 'accent', pressed: 'accentPressed', text: 'textInverse', border: 'accent' },
  secondary: { background: 'surfaceRaised', pressed: 'surfaceMuted', text: 'text', border: 'border' },
  ghost: { background: 'background', pressed: 'surfaceRaised', text: 'textMuted', border: 'background' },
};

export function Button({
  label,
  variant = 'primary',
  loading = false,
  compact = false,
  icon,
  trailingIcon,
  disabled,
  style,
  ...rest
}: ButtonProps) {
  const palette = variantStyles[variant];
  const isDisabled = disabled || loading;
  const iconSize = compact ? 16 : 18;

  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ disabled: isDisabled, busy: loading }}
      disabled={isDisabled}
      style={(state) => [
        styles.base,
        compact && styles.compact,
        {
          backgroundColor: state.pressed ? colors[palette.pressed] : variant === 'ghost' ? 'transparent' : colors[palette.background],
          borderColor: variant === 'ghost' ? 'transparent' : colors[palette.border],
          opacity: isDisabled ? 0.5 : 1,
        },
        typeof style === 'function' ? style(state) : style,
      ]}
      {...rest}
    >
      {loading ? (
        <ActivityIndicator color={colors[palette.text]} />
      ) : (
        <>
          {icon ? <Ionicons name={icon} size={iconSize} color={colors[palette.text]} /> : null}
          <AppText variant="bodyStrong" color={palette.text}>
            {label}
          </AppText>
          {trailingIcon ? <Ionicons name={trailingIcon} size={iconSize} color={colors[palette.text]} /> : null}
        </>
      )}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  base: {
    minHeight: layout.minTouchTarget + 4,
    paddingHorizontal: spacing.xl,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.sm,
  },
  compact: { minHeight: 38, paddingHorizontal: spacing.lg },
});
