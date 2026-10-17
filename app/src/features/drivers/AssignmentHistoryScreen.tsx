import { useState } from 'react';
import { listAssignments } from '@/api/assignments';
import { useSession } from '@/auth/SessionProvider';
import { AppText, Badge, Card, Screen, SelectField, TextField } from '@/components';
import { useTeams } from '@/data/accounts';
import { useRemote } from '@/data/useRemote';
import { useReloadOnFocus } from '@/data/useReloadOnFocus';
import { RemoteContent } from '@/features/RemoteContent';
import { formatDateTime } from '@/format';

export function AssignmentHistoryScreen() {
 const isFia = useSession().user?.role === 'fia_admin';
 const [teamId, setTeamId] = useState('');
 const [search, setSearch] = useState('');
 const teams = useTeams();
 const history = useRemote(() => listAssignments(teamId ? Number(teamId) : undefined), [teamId]);
 useReloadOnFocus(history.reload);
 return <Screen title="Historial de asignaciones" subtitle={isFia ? 'Altas, cambios de rol, transferencias y bajas de todas las escuderías.' : 'Movimientos históricos de tu escudería.'}>
  {isFia ? <RemoteContent state={teams.state} onRetry={teams.reload}>{(items) =>
   <SelectField options={[{ value: '', label: 'Todas las escuderías' }, ...items.map((team) => ({ value: String(team.id), label: team.name }))]}
    selected={teamId} onSelect={setTeamId} accessibilityLabel="Filtrar escudería" />
  }</RemoteContent> : null}
  <TextField label="Buscar piloto" icon="search" value={search} onChangeText={setSearch} />
  <RemoteContent state={history.state} onRetry={history.reload}>{(items) => {
   const rows = items.filter((item) => `${item.firstName} ${item.lastName}`.toLocaleLowerCase().includes(search.trim().toLocaleLowerCase()));
   return rows.length === 0 ? <AppText color="textMuted">No hay asignaciones para esta búsqueda.</AppText> : rows.map((item) =>
    <Card key={item.id} title={`${item.firstName} ${item.lastName}`} action={<Badge label={item.teamId === null ? 'Baja' : item.isStarter ? 'Titular' : 'Suplente'} tone={item.teamId === null ? 'neutral' : 'accent'} />}>
     <AppText>{item.previousTeamName ? `${item.previousTeamName} → ` : ''}{item.teamName}</AppText>
     <AppText color="textMuted">Desde {formatDateTime(item.effectiveFrom)}</AppText>
     <AppText color="textMuted">{item.effectiveUntil ? `Hasta ${formatDateTime(item.effectiveUntil)}` : 'Sin fecha de fin'}</AppText>
    </Card>);
  }}</RemoteContent>
 </Screen>;
}
