import { View } from 'react-native';
import { SkeletonGroup, Skeleton } from '../Skeleton';
import { SkeletonCard, SkeletonRow, SkeletonSummary } from './parts';
import { useLayout } from '../../utils/responsive';
import { R, S } from '../../utils/theme';

/** Mirrors the home screen: month selector, summary card, search box and transaction rows. */
export default function ExpensesSkeleton({ label }: { label: string }) {
  const layout = useLayout();
  const isWide = !layout.isPhone;
  return (
    <SkeletonGroup label={label} style={{ flex: 1, overflow: 'hidden' }}>
      <View style={{ paddingHorizontal: layout.px, paddingTop: S.lg, paddingBottom: 4 }}>
        <View style={{ alignItems: 'center', marginBottom: S.md }}>
          <Skeleton width={160} height={20} />
        </View>
        <SkeletonCard style={{ flexDirection: 'row', alignItems: 'center', padding: isWide ? 20 : S.lg, marginBottom: S.sm }}>
          <SkeletonSummary items={3} />
        </SkeletonCard>
        <Skeleton height={42} radius={R.sm} style={{ marginBottom: 4 }} />
      </View>
      <View style={{ paddingTop: 4 }}>
        {Array.from({ length: 10 }, (_, i) => (
          <SkeletonRow key={i} paddingHorizontal={isWide ? 20 : 16} />
        ))}
      </View>
    </SkeletonGroup>
  );
}
