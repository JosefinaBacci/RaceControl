import { StyleSheet, View } from 'react-native';

import type { FieldErrors } from '@/api/formFailure';
import { ChipGroup, FieldGroup, TextField } from '@/components';
import { spacing } from '@/theme';

import type { DriverField, DriverFieldValues } from './driverForm';

const roleOptions = [
  { value: 'starter', label: 'Titular' },
  { value: 'reserve', label: 'Suplente' },
] as const;

type DriverFieldsProps = {
  values: DriverFieldValues;
  onChange: (values: DriverFieldValues) => void;
  errors: FieldErrors<DriverField>;
};

export function DriverFields({ values, onChange, errors }: DriverFieldsProps) {
  return (
    <View style={styles.fields}>
      <TextField
        label="Nombre"
        icon="person-outline"
        value={values.firstName}
        onChangeText={(firstName) => onChange({ ...values, firstName })}
        error={errors.firstName}
        autoCorrect={false}
      />
      <TextField
        label="Apellido"
        value={values.lastName}
        onChangeText={(lastName) => onChange({ ...values, lastName })}
        error={errors.lastName}
        autoCorrect={false}
      />
      <TextField
        label="Licencia"
        icon="card-outline"
        value={values.licenseNumber}
        onChangeText={(licenseNumber) => onChange({ ...values, licenseNumber })}
        error={errors.licenseNumber}
        autoCapitalize="characters"
        autoCorrect={false}
        maxLength={32}
      />
      <TextField
        label="País (dos letras)"
        value={values.countryCode}
        onChangeText={(countryCode) => onChange({ ...values, countryCode })}
        error={errors.countryCode}
        autoCapitalize="none"
        autoCorrect={false}
        maxLength={2}
      />
      <TextField
        label="Número de temporada"
        value={values.number}
        onChangeText={(number) => onChange({ ...values, number })}
        error={errors.number}
        keyboardType="number-pad"
        maxLength={3}
      />
      <FieldGroup label="Rol">
        <ChipGroup
          options={roleOptions}
          selected={values.isStarter ? 'starter' : 'reserve'}
          onSelect={(role) => onChange({ ...values, isStarter: role === 'starter' })}
          accessibilityLabel="Elegir rol"
        />
      </FieldGroup>
    </View>
  );
}

const styles = StyleSheet.create({
  fields: { gap: spacing.lg },
});
