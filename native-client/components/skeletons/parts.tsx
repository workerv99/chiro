/** Shared skeleton building blocks that mirror the real list rows and cards. */
import { ReactNode } from 'react';
import { StyleProp, StyleSheet, View, ViewStyle } from 'react-native';
import { Skeleton } from '../Skeleton';
import { C, R, S } from '../../utils/theme';

/** Avatar + two text lines (+ optional trailing amount), like the tab list rows. */
export function SkeletonRow({
  paddingHorizontal = S.lg,
  avatarSize = 40,
  round = false,
  trailing = 'amount',
}: {
  paddingHorizontal?: number;
  avatarSize?: number;
  round?: boolean;
  trailing?: 'amount' | 'chevron';
}) {
  return (
    <View style={[row.base, { paddingHorizontal }]}>
      <Skeleton width={avatarSize} height={avatarSize} radius={round ? R.full : R.sm} />
      <View style={row.body}>
        <Skeleton width="55%" height={14} style={{ marginBottom: 3 }} />
        <Skeleton width="35%" height={11} />
      </View>
      {trailing === 'amount' ? <Skeleton width={64} height={16} /> : <Skeleton width={18} height={18} />}
    </View>
  );
}

/** Bordered surface card, like the summary / list cards. */
export function SkeletonCard({ children, style }: { children: ReactNode; style?: StyleProp<ViewStyle> }) {
  return <View style={[card.base, style]}>{children}</View>;
}

/** Row of label + value stacks separated by hairline dividers (summary card content). */
export function SkeletonSummary({ items, valueWidth = 70 }: { items: number; valueWidth?: number }) {
  return (
    <>
      {Array.from({ length: items }, (_, i) => (
        <View key={i} style={{ flex: 1, flexDirection: 'row', alignItems: 'center' }}>
          {i > 0 && <View style={card.divider} />}
          <View style={{ flex: 1, alignItems: 'center' }}>
            <Skeleton width={44} height={10} style={{ marginBottom: 5 }} />
            <Skeleton width={valueWidth} height={17} />
          </View>
        </View>
      ))}
    </>
  );
}

const row = StyleSheet.create({
  base: {
    flexDirection: 'row', alignItems: 'center', gap: S.md,
    paddingVertical: 13,
    borderBottomWidth: 1, borderBottomColor: C.borderFaint,
  },
  body: { flex: 1 },
});

const card = StyleSheet.create({
  base: {
    backgroundColor: C.surface, borderRadius: R.md,
    borderWidth: 1, borderColor: C.border,
  },
  divider: { width: 1, height: 36, backgroundColor: C.border },
});
