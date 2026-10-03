import Ionicons from '@expo/vector-icons/Ionicons';
import { router } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View, useWindowDimensions } from 'react-native';

import type { Role } from '@/api/auth';
import type { Account, AccountStatus } from '@/api/users';
import { roleLabel } from '@/auth/roles';
import { useSession } from '@/auth/SessionProvider';
import { AppText, Avatar, Badge, Button, Card, ChipGroup, IconButton, Screen, TextField } from '@/components';
import { useAccounts } from '@/data/accounts';
import { useStatusToggle } from '@/features/accounts/useStatusToggle';
import { RemoteContent } from '@/features/RemoteContent';
import { teamById } from '@/mocks/catalog';
import { colors, spacing } from '@/theme';

type RoleFilter = Role | 'all';
type StatusFilter = AccountStatus | 'all';

const roleFilterOptions = [
  { value: 'all' as const, label: 'Todos' },
  { value: 'fia_admin' as const, label: 'Administradores' },
  { value: 'team_admin' as const, label: 'Escuderías' },
];

const statusFilterOptions = [
  { value: 'all' as const, label: 'Cualquier estado' },
  { value: 'active' as const, label: 'Activos' },
  { value: 'inactive' as const, label: 'Desactivados' },
];

const tableBreakpoint = 720;

export default function FiaUsersScreen() {
  const [search, setSearch] = useState('');
  const [role, setRole] = useState<RoleFilter>('all');
  const [status, setStatus] = useState<StatusFilter>('all');
  const { state, reload } = useAccounts({
    search,
    role: role === 'all' ? undefined : role,
    status: status === 'all' ? undefined : status,
  });
  const showsTable = useWindowDimensions().width >= tableBreakpoint;

  return (
    <Screen
      title="Gestión de usuarios"
      subtitle="Creá, modificá o desactivá cuentas y asigná su rol."
      headerAction={<Button label="Nuevo usuario" icon="add" compact onPress={() => router.push('/fia/users/new')} />}
    >
      <ChipGroup variant="underline" options={roleFilterOptions} selected={role} onSelect={setRole} accessibilityLabel="Filtrar por rol" />
      <ChipGroup options={statusFilterOptions} selected={status} onSelect={setStatus} accessibilityLabel="Filtrar por estado" />
      <TextField
        label="Buscar usuario o email…"
        icon="search-outline"
        value={search}
        onChangeText={setSearch}
        autoCapitalize="none"
        autoCorrect={false}
      />
      <Card flush>
        {showsTable ? <TableHeader /> : null}
        <RemoteContent state={state} onRetry={reload}>
          {(accounts) =>
            accounts.length === 0 ? (
              <AppText color="textMuted" style={styles.empty}>
                No hay usuarios que coincidan con la búsqueda.
              </AppText>
            ) : (
              accounts.map((account) => <AccountRow key={account.id} account={account} showsTable={showsTable} onChanged={reload} />)
            )
          }
        </RemoteContent>
      </Card>
    </Screen>
  );
}

function TableHeader() {
  return (
    <View style={[styles.row, styles.headerRow]}>
      {['Usuario', 'Rol', 'Estado', 'Acciones'].map((label, index) => (
        <AppText key={label} variant="overline" color="textMuted" style={columnStyleList[index]}>
          {label}
        </AppText>
      ))}
    </View>
  );
}

function AccountRow({ account, showsTable, onChanged }: { account: Account; showsTable: boolean; onChanged: () => void }) {
  const roleText = account.teamName ? `${roleLabel[account.role]} · ${account.teamName}` : roleLabel[account.role];
  const avatarColor = account.role === 'team_admin' ? teamById(account.teamId).color : colors.accent;
  const status = (
    <Badge
      label={account.isActive ? 'Activo' : 'Desactivado'}
      tone={account.isActive ? 'success' : 'neutral'}
      icon={account.isActive ? 'checkmark' : 'remove'}
    />
  );

  return (
    <View style={styles.row}>
      <View style={[columnStyleList[0], styles.userCell]}>
        <Avatar name={account.username} color={avatarColor} />
        <View style={styles.userText}>
          <AppText variant="bodyStrong" numberOfLines={1}>
            {account.username}
          </AppText>
          <AppText variant="caption" color="textMuted" numberOfLines={1}>
            {showsTable ? (account.email ?? 'Sin email') : roleText}
          </AppText>
        </View>
      </View>
      {showsTable ? (
        <>
          <View style={[columnStyleList[1], styles.roleCell]}>
            <Ionicons name={account.role === 'fia_admin' ? 'shield-checkmark-outline' : 'car-sport-outline'} size={16} color={colors.textMuted} />
            <AppText variant="caption" color="textMuted" numberOfLines={1} style={styles.userText}>
              {roleText}
            </AppText>
          </View>
          <View style={columnStyleList[2]}>{status}</View>
          <View style={columnStyleList[3]}>
            <RowActions account={account} onChanged={onChanged} />
          </View>
        </>
      ) : (
        <View style={styles.compactTrailing}>
          {status}
          <RowActions account={account} onChanged={onChanged} />
        </View>
      )}
    </View>
  );
}

function RowActions({ account, onChanged }: { account: Account; onChanged: () => void }) {
  const isSelf = useSession().user?.id === account.id;
  const status = useStatusToggle(account, onChanged);
  const openAccount = () => router.push({ pathname: '/fia/users/[id]', params: { id: String(account.id) } });

  if (status.isConfirming) {
    return (
      <View style={styles.actionsColumn}>
        <AppText variant="caption" color="danger">
          ¿Desactivar?
        </AppText>
        <View style={styles.actions}>
          <IconButton icon="close" accessibilityLabel="Cancelar" onPress={status.cancel} disabled={status.isSubmitting} />
          <IconButton
            icon="checkmark"
            tone="danger"
            accessibilityLabel={`Confirmar la desactivación de ${account.username}`}
            onPress={status.toggle}
            disabled={status.isSubmitting}
          />
        </View>
      </View>
    );
  }
  return (
    <View style={styles.actionsColumn}>
      <View style={styles.actions}>
        <IconButton icon="create-outline" accessibilityLabel={`Editar ${account.username}`} onPress={openAccount} />
        {account.isActive ? (
          <IconButton
            icon="trash-outline"
            tone="danger"
            accessibilityLabel={`Desactivar ${account.username}`}
            onPress={status.ask}
            disabled={isSelf}
          />
        ) : (
          <IconButton icon="refresh" accessibilityLabel={`Reactivar ${account.username}`} onPress={status.toggle} disabled={status.isSubmitting} />
        )}
      </View>
      {status.error ? (
        <AppText variant="caption" color="danger">
          {status.error}
        </AppText>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.md,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderTopWidth: StyleSheet.hairlineWidth,
    borderTopColor: colors.border,
  },
  headerRow: { backgroundColor: colors.surfaceRaised, borderTopWidth: 0, paddingVertical: spacing.sm },
  userCell: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
  userText: { flexShrink: 1 },
  roleCell: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm },
  actions: { flexDirection: 'row', gap: spacing.sm },
  actionsColumn: { gap: spacing.xs },
  compactTrailing: { alignItems: 'flex-end', gap: spacing.sm },
  empty: { padding: spacing.lg },
});

const columnStyles = StyleSheet.create({
  user: { flex: 3 },
  role: { flex: 3 },
  status: { flex: 1.4 },
  actions: { flex: 1.6 },
});

const columnStyleList = [columnStyles.user, columnStyles.role, columnStyles.status, columnStyles.actions];
