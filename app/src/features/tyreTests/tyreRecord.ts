import type { HistoryType } from '@/api/history';
import type { TyreTestRevision } from '@/api/tyreTests';

// What the FIA screen of one tyre test shows: per compound, its record and the
// result in force, plus the compounds that still have none.

export type CompoundRecord = {
  compound: HistoryType;
  revisions: TyreTestRevision[];
  current: TyreTestRevision;
};

export function compoundRecords(revisions: readonly TyreTestRevision[]): CompoundRecord[] {
  const byCompound = new Map<string, TyreTestRevision[]>();
  for (const revision of revisions) {
    byCompound.set(revision.type.code, [...(byCompound.get(revision.type.code) ?? []), revision]);
  }
  return [...byCompound.values()].map((chain) => ({
    compound: chain[0].type,
    revisions: chain,
    current: chain.find((revision) => revision.isCurrent) ?? chain[chain.length - 1],
  }));
}

export function untestedCompounds(compounds: readonly HistoryType[], revisions: readonly TyreTestRevision[]): HistoryType[] {
  const tested = new Set(revisions.map((revision) => revision.type.code));
  return compounds.filter((compound) => !tested.has(compound.code));
}

export function compoundError(compoundCode: string): string | undefined {
  return compoundCode ? undefined : 'Elegí el compuesto que se probó';
}
