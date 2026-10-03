import { LinearGradient } from 'expo-linear-gradient';
import { Redirect, router } from 'expo-router';
import { useRef, useState } from 'react';
import { Image, KeyboardAvoidingView, Platform, ScrollView, StyleSheet, View, useWindowDimensions, type TextInput } from 'react-native';

import {
  describeLoginFailure,
  hasErrors,
  normalizeUsername,
  validateLoginForm,
  type LoginFieldErrors,
} from '@/auth/loginForm';
import { roleHome } from '@/auth/roles';
import { useSession } from '@/auth/SessionProvider';
import { AppText, Button, Logo, Notice, TextField } from '@/components';
import { FullScreenSpinner } from '@/navigation/RoleGate';
import { loginPhoto } from '@/photos';
import { colors, spacing } from '@/theme';

const splitLayoutBreakpoint = 900;

export default function LoginScreen() {
  const session = useSession();
  const { width, height } = useWindowDimensions();
  const isSplitLayout = width >= splitLayoutBreakpoint;
  const passwordRef = useRef<TextInput>(null);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [fieldErrors, setFieldErrors] = useState<LoginFieldErrors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (session.status === 'loading') {
    return <FullScreenSpinner />;
  }
  if (session.status === 'authenticated' && !isSubmitting) {
    return <Redirect href={roleHome[session.user.role]} />;
  }

  const submit = async () => {
    const normalizedUsername = normalizeUsername(username);
    const errors = validateLoginForm(normalizedUsername, password);
    setFieldErrors(errors);
    setFormError(null);
    if (hasErrors(errors)) {
      return;
    }

    setIsSubmitting(true);
    try {
      const user = await session.signIn(normalizedUsername, password);
      router.replace(roleHome[user.role]);
    } catch (error) {
      const failure = describeLoginFailure(error);
      setFieldErrors(failure.fieldErrors);
      setFormError(failure.message);
      setPassword('');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <View style={styles.screen}>
      {isSplitLayout ? <BrandPanel /> : null}
      <KeyboardAvoidingView style={styles.formSide} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
        <ScrollView
          contentContainerStyle={isSplitLayout ? styles.formScroll : styles.formScrollCompact}
          keyboardShouldPersistTaps="handled"
        >
          {isSplitLayout ? null : <MobileHero height={Math.min(height * 0.44, 380)} />}
          <View style={[styles.form, !isSplitLayout && styles.formCompact]}>
            {isSplitLayout ? <Logo size="lg" /> : null}
            <View style={styles.heading}>
              <AppText variant="title" accessibilityRole="header">
                Iniciar sesión
              </AppText>
              <AppText color="textMuted">Accedé a tu cuenta de la FIA o de tu escudería.</AppText>
            </View>

            {formError ? <Notice tone="danger" message={formError} /> : null}

            <TextField
              label="Usuario"
              icon="person-outline"
              value={username}
              onChangeText={setUsername}
              error={fieldErrors.username}
              autoCapitalize="none"
              autoCorrect={false}
              autoComplete="username"
              textContentType="username"
              returnKeyType="next"
              onSubmitEditing={() => passwordRef.current?.focus()}
              maxLength={32}
              editable={!isSubmitting}
            />
            <TextField
              ref={passwordRef}
              label="Contraseña"
              icon="lock-closed-outline"
              value={password}
              onChangeText={setPassword}
              error={fieldErrors.password}
              secureTextEntry
              autoCapitalize="none"
              autoComplete="current-password"
              textContentType="password"
              returnKeyType="go"
              onSubmitEditing={submit}
              editable={!isSubmitting}
            />

            <Button label="Ingresar" onPress={submit} loading={isSubmitting} />

            <View style={styles.divider}>
              <View style={styles.line} />
              <AppText variant="caption" color="textMuted">
                o seguí sin cuenta
              </AppText>
              <View style={styles.line} />
            </View>

            <Button
              label="Ver calendario y resultados"
              variant="secondary"
              icon="flag-outline"
              onPress={() => router.replace('/')}
              disabled={isSubmitting}
            />
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </View>
  );
}

function PhotoBackdrop({ fadesIntoForm }: { fadesIntoForm: boolean }) {
  return (
    <>
      <Image source={loginPhoto.source} style={styles.photo} resizeMode="cover" accessibilityIgnoresInvertColors />
      <LinearGradient
        colors={['rgba(90,0,0,0.25)', 'rgba(170,0,0,0.45)', 'rgba(70,0,0,0.88)', colors.background]}
        locations={[0, 0.4, 0.72, 1]}
        start={{ x: 0.5, y: 0 }}
        end={{ x: 0.5, y: 1 }}
        style={StyleSheet.absoluteFill}
      />
      <LinearGradient
        colors={['rgba(225,6,0,0.35)', 'transparent']}
        start={{ x: 0, y: 1 }}
        end={{ x: 0.8, y: 0.2 }}
        style={StyleSheet.absoluteFill}
      />
      {fadesIntoForm ? (
        <LinearGradient
          colors={['transparent', colors.background]}
          start={{ x: 0.85, y: 0.5 }}
          end={{ x: 1, y: 0.5 }}
          style={StyleSheet.absoluteFill}
        />
      ) : null}
      <View style={styles.stripes} pointerEvents="none">
        {[0, 1, 2, 3, 4].map((index) => (
          <View key={index} style={[styles.stripe, { opacity: 0.04 + index * 0.02 }]} />
        ))}
      </View>
      <AppText variant="caption" style={styles.credit}>
        {loginPhoto.credit}
      </AppText>
    </>
  );
}

function BrandPanel() {
  return (
    <View style={styles.brandPanel}>
      <PhotoBackdrop fadesIntoForm />
      <View style={styles.brandContent}>
        <AppText variant="overline" style={styles.tagline}>
          Speed · Data · Passion
        </AppText>
        <AppText variant="hero" style={[styles.brandHeadline, styles.shadowed]}>
          Calendario, puntajes y sanciones en un solo lugar.
        </AppText>
        <AppText style={styles.shadowed}>F1 · F2 · F3 · F1 Academy</AppText>
      </View>
    </View>
  );
}

function MobileHero({ height }: { height: number }) {
  return (
    <View style={[styles.mobileHero, { height }]}>
      <PhotoBackdrop fadesIntoForm={false} />
      <LinearGradient
        colors={['transparent', 'rgba(11,11,15,0.75)', colors.background]}
        locations={[0.35, 0.75, 1]}
        style={StyleSheet.absoluteFill}
      />
      <View style={styles.mobileHeroContent}>
        <Logo size="lg" />
        <AppText variant="overline" style={styles.tagline}>
          Speed · Data · Passion
        </AppText>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, flexDirection: 'row', backgroundColor: colors.background },
  brandPanel: { flex: 1, overflow: 'hidden', justifyContent: 'flex-end' },
  photo: { position: 'absolute', top: 0, left: 0, width: '100%', height: '100%' },
  stripes: { position: 'absolute', top: -80, right: 40, flexDirection: 'row', gap: 22, transform: [{ skewX: '-24deg' }] },
  stripe: { width: 46, height: 1400, backgroundColor: colors.text },
  mobileHero: { width: '100%', overflow: 'hidden', justifyContent: 'flex-end' },
  mobileHeroContent: { padding: spacing.xl, paddingBottom: spacing.xxl, gap: spacing.sm },
  formScrollCompact: { flexGrow: 1 },
  formCompact: { marginTop: -spacing.lg, paddingHorizontal: spacing.xl, paddingBottom: spacing.xxl },
  credit: { position: 'absolute', top: spacing.md, left: spacing.lg, color: 'rgba(255,255,255,0.7)', fontSize: 10 },
  shadowed: { textShadowColor: 'rgba(0,0,0,0.7)', textShadowRadius: 10, textShadowOffset: { width: 0, height: 1 } },
  brandContent: { padding: spacing.xxl * 1.5, gap: spacing.lg, maxWidth: 560 },
  brandHeadline: { lineHeight: 40 },
  tagline: { color: colors.text, letterSpacing: 6, opacity: 0.85 },
  formSide: { flex: 1, maxWidth: 640 },
  formScroll: { flexGrow: 1, justifyContent: 'center', padding: spacing.xl },
  form: { width: '100%', maxWidth: 400, alignSelf: 'center', gap: spacing.lg },
  heading: { gap: spacing.xs, marginTop: spacing.lg },
  divider: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
  line: { flex: 1, height: StyleSheet.hairlineWidth, backgroundColor: colors.border },
});
