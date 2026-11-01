import { Screen } from '@/components';
import { openPublicSanction } from '@/features/sanctions/navigation';
import { SanctionBrowser } from '@/features/sanctions/SanctionBrowser';

export default function PublicSanctionsScreen() {
  return (
    <Screen title="Sanciones" subtitle="Registro oficial de las sanciones de la FIA a pilotos y escuderías.">
      <SanctionBrowser onOpen={openPublicSanction} />
    </Screen>
  );
}
