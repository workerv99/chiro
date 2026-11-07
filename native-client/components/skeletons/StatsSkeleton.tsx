import { View } from 'react-native';
import { Skeleton, SkeletonGroup } from '../Skeleton';
import { SkeletonCard } from './parts';
import { CONTENT_MAX_W, useLayout } from '../../utils/responsive';
import { R, S } from '../../utils/theme';

function Heading() {
  return <Skeleton width={96} height={10} style={{ marginTop: S['2xl'] - 4, marginBottom: S.sm }} />;
}

/** Account-style row: square chip, name + subtitle, trailing amount. */
function ListRow() {
  return (
    <View style={{ flexDirection: 'row', alignItems: 'center', padding: S.md, gap: S.sm }}>
      <Skeleton width={36} height={36} radius={R.xs} />
      <View style={{ flex: 1, gap: 5 }}>
        <Skeleton width="45%" height={13} />
        <Skeleton height={8} radius={4} />
      </View>
      <Skeleton width={56} height={13} />
    </View>
  );
}

/** Mirrors the stats screen: year selector, three totals, then section cards. */
export default function StatsSkeleton({ label }: { label: string }) {
  const l = useLayout();
  const desktop = l.isDesktop ? { maxWidth: CONTENT_MAX_W, width: '100%' as const, alignSelf: 'center' as const } : null;
  return (
    <SkeletonGroup label={label} style={{ flex: 1, overflow: 'hidden', paddingHorizontal: l.px, paddingTop: l.px }}>
      <View style={desktop}>
        <View style={{ alignItems: 'center', marginBottom: S.xl }}>
          <Skeleton width={72} height={32} />
        </View>
        <View style={{ flexDirection: 'row', gap: S.sm }}>
          {[0, 1, 2].map(i => (
            <SkeletonCard key={i} style={{ flex: 1, padding: S.md, alignItems: 'center' }}>
              <Skeleton width={40} height={10} style={{ marginBottom: 6 }} />
              <Skeleton width="80%" height={14} />
            </SkeletonCard>
          ))}
        </View>
        <Heading />
        <SkeletonCard>
          {[0, 1, 2].map(i => <ListRow key={i} />)}
        </SkeletonCard>
        <Heading />
        <SkeletonCard>
          {[0, 1, 2, 3, 4].map(i => <ListRow key={i} />)}
        </SkeletonCard>
      </View>
    </SkeletonGroup>
  );
}
