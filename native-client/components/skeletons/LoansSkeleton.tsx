import { View } from 'react-native';
import { SkeletonGroup } from '../Skeleton';
import { SkeletonCard, SkeletonRow, SkeletonSummary } from './parts';
import { CONTENT_MAX_W, useLayout } from '../../utils/responsive';
import { S } from '../../utils/theme';

/** Mirrors the loans screen: two-figure summary card and person rows. */
export default function LoansSkeleton({ label }: { label: string }) {
  const l = useLayout();
  const desktop = l.isDesktop ? { maxWidth: CONTENT_MAX_W, width: '100%' as const, alignSelf: 'center' as const } : null;
  return (
    <SkeletonGroup label={label} style={{ flex: 1, overflow: 'hidden' }}>
      <SkeletonCard
        style={[
          { flexDirection: 'row', alignItems: 'center', marginHorizontal: S.lg, marginTop: S.lg, marginBottom: S.xs, padding: S.lg },
          desktop && { ...desktop, marginHorizontal: 0 },
        ]}
      >
        <SkeletonSummary items={2} valueWidth={90} />
      </SkeletonCard>
      <View style={[{ paddingTop: S.xs }, desktop]}>
        {Array.from({ length: 9 }, (_, i) => (
          <SkeletonRow key={i} avatarSize={42} round trailing="chevron" />
        ))}
      </View>
    </SkeletonGroup>
  );
}
