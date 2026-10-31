import { router } from 'expo-router';

import { Button, Screen } from '@/components';
import { openFiaSanction } from '@/features/sanctions/navigation';
import { SanctionBrowser } from '@/features/sanctions/SanctionBrowser';

export default function FiaSanctionsScreen() {
  return (
    <Screen
      title="Sanciones"
      subtitle="Registrá, modificá o revocá sanciones. Cada cambio se notifica a las escuderías del campeonato."
      headerAction={<Button label="Registrar sanción" icon="add" compact onPress={() => router.push('/fia/sanctions/new')} />}
    >
      <SanctionBrowser onOpen={openFiaSanction} />
    </Screen>
  );
}
