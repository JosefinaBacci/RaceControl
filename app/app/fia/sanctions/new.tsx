import { router } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { hasFieldErrors } from '@/api/formFailure';
import { issueSanction } from '@/api/sanctions';
import { Button, Card, Notice, Screen } from '@/components';
import { useSanctionCatalogs, type SanctionCatalogs } from '@/data/sanctions';
import { useFormSubmission } from '@/data/useFormSubmission';
import { RemoteContent } from '@/features/RemoteContent';
import { SubjectFields, TermsFields } from '@/features/sanctions/SanctionFields';
import {
  emptySanctionForm,
  sanctionFormFields,
  subjectCategory,
  toNewSanction,
  validateNewSanction,
  type SanctionFormErrors,
} from '@/features/sanctions/sanctionForm';
import { spacing } from '@/theme';

export default function NewSanctionScreen() {
  const { state, reload } = useSanctionCatalogs();

  return (
    <Screen
      title="Registrar sanción"
      subtitle="La sanción queda registrada con la fecha y tu usuario, y se notifica a las escuderías del campeonato."
      headerAction={<Button label="Volver" icon="arrow-back" variant="ghost" compact onPress={() => router.back()} />}
    >
      <RemoteContent state={state} onRetry={reload}>
        {(catalogs) => <NewSanctionForm catalogs={catalogs} />}
      </RemoteContent>
    </Screen>
  );
}

function NewSanctionForm({ catalogs }: { catalogs: SanctionCatalogs }) {
  const [values, setValues] = useState(emptySanctionForm);
  const [localErrors, setLocalErrors] = useState<SanctionFormErrors>({});
  const { fieldErrors, formError, isSubmitting, submit } = useFormSubmission(sanctionFormFields);

  const save = async () => {
    const errors = validateNewSanction(values, catalogs.penaltyTypes);
    setLocalErrors(errors);
    if (hasFieldErrors(errors)) {
      return;
    }
    const issued = await submit(() => issueSanction(toNewSanction(values, catalogs.penaltyTypes)));
    if (issued) {
      router.replace({ pathname: '/fia/sanctions/[id]', params: { id: String(issued.id) } });
    }
  };

  const errors = { ...fieldErrors, ...localErrors };
  return (
    <Card title="Datos de la sanción">
      {formError ? <Notice tone="danger" message={formError} /> : null}
      <SubjectFields values={values} onChange={setValues} errors={errors} drivers={catalogs.drivers} />
      <TermsFields
        values={values}
        onChange={setValues}
        errors={errors}
        penaltyTypes={catalogs.penaltyTypes}
        categoryCode={subjectCategory(values, catalogs.drivers, catalogs.teams)}
      />
      <View style={styles.actions}>
        <Button label="Registrar sanción" icon="save-outline" onPress={save} loading={isSubmitting} />
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.md },
});
