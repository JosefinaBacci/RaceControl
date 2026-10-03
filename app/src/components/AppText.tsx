import { Text, type TextProps } from 'react-native';

import { colors, typography, type ColorName, type TypographyVariant } from '@/theme';

type AppTextProps = TextProps & {
  variant?: TypographyVariant;
  color?: ColorName;
};

export function AppText({ variant = 'body', color = 'text', style, ...rest }: AppTextProps) {
  return <Text style={[typography[variant], { color: colors[color] }, style]} {...rest} />;
}
