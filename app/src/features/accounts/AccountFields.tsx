import { StyleSheet, View } from 'react-native';

import type { Role } from '@/api/auth';
import type { FieldErrors } from '@/api/formFailure';
import { roleLabel } from '@/auth/roles';
import { AppText, ChipGroup, FieldGroup, TextField } from '@/components';
import { TeamPicker } from '@/features/TeamPicker';
import { spacing } from '@/theme';

export type AccountFieldValues = {
  email: string;
  role: Role;
  teamId: number | null;
};

export type AccountField = 'email' | 'role' | 'teamId';

type AccountFieldsProps = {
  values: AccountFieldValues;
  onChange: (values: AccountFieldValues) => void;
  errors: FieldErrors<AccountField>;
  isRoleLocked?: boolean;
};

const roleOptions = (Object.keys(roleLabel) as Role[]).map((role) => ({ value: role, label: roleLabel[role] }));

export function AccountFields({ values, onChange, errors, isRoleLocked = false }: AccountFieldsProps) {
  return (
    <View style={styles.fields}>
      <TextField
        label="Email (opcional)"
        icon="mail-outline"
        value={values.email}
        onChangeText={(email) => onChange({ ...values, email })}
        error={errors.email}
        autoCapitalize="none"
        autoCorrect={false}
        keyboardType="email-address"
      />
      <FieldGroup label="Rol" error={errors.role}>
        {isRoleLocked ? (
          <AppText color="textMuted">{roleLabel[values.role]} · no podés cambiar tu propio rol.</AppText>
        ) : (
          <ChipGroup
            options={roleOptions}
            selected={values.role}
            onSelect={(role) => onChange({ ...values, role, teamId: role === 'fia_admin' ? null : values.teamId })}
            accessibilityLabel="Elegir rol"
          />
        )}
      </FieldGroup>
      {values.role === 'team_admin' ? (
        <TeamPicker teamId={values.teamId} onChange={(teamId) => onChange({ ...values, teamId })} error={errors.teamId} />
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  fields: { gap: spacing.lg },
});
