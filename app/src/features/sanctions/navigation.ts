import { router } from 'expo-router';

import type { Sanction } from '@/api/sanctions';

export function openPublicSanction(sanction: Sanction) {
  router.push({ pathname: '/sanctions/[id]', params: { id: String(sanction.id) } });
}

export function openFiaSanction(sanction: Sanction) {
  router.push({ pathname: '/fia/sanctions/[id]', params: { id: String(sanction.id) } });
}
