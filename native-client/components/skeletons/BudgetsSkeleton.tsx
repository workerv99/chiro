import { View } from 'react-native';
import { Skeleton, SkeletonGroup } from '../Skeleton';
import { SkeletonCard, SkeletonSummary } from './parts';
import { CONTENT_MAX_W, useLayout } from '../../utils/responsive';
import { R, S } from '../../utils/theme';

function SectionRow() {
  return (
    <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginBottom: S.sm }}>
      <Skeleton width={110} height={14} />
      <Skeleton width={32} height={32} radius={R.full} />
    </View>
  );
}

/** Budget / piggy-bank card: avatar + two lines, progress bar and a footer line. */
function ItemCard({ wide }: { wide: boolean }) {
  return (
    <SkeletonCard style={[{ padding: S.md, marginBottom: S.sm }, wide && { width: '49%' }]}>
      <View style={{ flexDirection: 'row', alignItems: 'center', gap: S.sm }}>
        <Skeleton width={38} height={38} radius={R.xs} />
        <View style={{ flex: 1, gap: 4 }}>
          <Skeleton width="45%" height={14} />
          <Skeleton width="60%" height={10} />
        </View>
        <Skeleton width={16} height={16} />
      </View>
      <Skeleton height={10} radius={5} style={{ marginTop: 8 }} />
      <Skeleton width="50%" height={12} style={{ marginTop: 6 }} />
    </SkeletonCard>
  );
}

/** Mirrors the budgets screen: month selector, summary card, budget cards and piggy banks. */
export default function BudgetsSkeleton({ label }: { label: string }) {
  const l = useLayout();
  const isWide = !l.isPhone;
  const desktop = l.isDesktop ? { maxWidth: CONTENT_MAX_W, width: '100%' as const, alignSelf: 'center' as const } : null;
  const grid = isWide ? { flexDirection: 'row' as const, flexWrap: 'wrap' as const, gap: S.sm } : null;
  return (
    <SkeletonGroup label={label} style={{ flex: 1, overflow: 'hidden', paddingHorizontal: l.px, paddingTop: l.px }}>
      <View style={desktop}>
        <View style={{ alignItems: 'center', marginBottom: S.lg }}>
          <Skeleton width={160} height={20} />
        </View>
        <SkeletonCard style={{ flexDirection: 'row', alignItems: 'center', padding: S.lg, marginBottom: S.xl }}>
          <SkeletonSummary items={3} />
        </SkeletonCard>
        <SectionRow />
        <View style={grid}>
          {[0, 1, 2].map(i => <ItemCard key={i} wide={isWide} />)}
        </View>
        <View style={{ marginTop: S.xl }}>
          <SectionRow />
        </View>
        <View style={grid}>
          {[0, 1].map(i => <ItemCard key={i} wide={isWide} />)}
        </View>
      </View>
    </SkeletonGroup>
  );
}
