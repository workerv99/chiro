/**
 * Skeleton — quiet placeholder blocks shown while a screen loads.
 *
 * `SkeletonGroup` owns ONE looping opacity animation (native driver) shared by every
 * `Skeleton` below it, so cost does not grow with the number of blocks. It also carries
 * the single accessibility announcement; the blocks themselves are hidden from assistive tech.
 * With reduced motion enabled the blocks stay static.
 */
import { createContext, ReactNode, useContext, useEffect, useRef, useState } from 'react';
import { AccessibilityInfo, Animated, DimensionValue, Easing, StyleProp, View, ViewStyle } from 'react-native';
import { C, R } from '../utils/theme';

const OPACITY_LOW = 0.45;
const OPACITY_HIGH = 0.9;
const STATIC_OPACITY = 0.7;
const HALF_CYCLE_MS = 800;

// null = not inside a group (or reduced motion): blocks render static.
const PulseContext = createContext<Animated.Value | null>(null);

export function SkeletonGroup({
  label,
  style,
  children,
}: {
  label: string;
  style?: StyleProp<ViewStyle>;
  children: ReactNode;
}) {
  const pulse = useRef(new Animated.Value(OPACITY_HIGH)).current;
  const [reduceMotion, setReduceMotion] = useState(false);

  useEffect(() => {
    let mounted = true;
    AccessibilityInfo.isReduceMotionEnabled()
      .then(v => { if (mounted) setReduceMotion(v); })
      .catch(() => {});
    const sub = AccessibilityInfo.addEventListener('reduceMotionChanged', setReduceMotion);
    return () => { mounted = false; sub.remove(); };
  }, []);

  useEffect(() => {
    if (reduceMotion) return;
    const step = (toValue: number) =>
      Animated.timing(pulse, { toValue, duration: HALF_CYCLE_MS, easing: Easing.inOut(Easing.ease), useNativeDriver: true });
    const loop = Animated.loop(Animated.sequence([step(OPACITY_LOW), step(OPACITY_HIGH)]));
    loop.start();
    return () => loop.stop();
  }, [pulse, reduceMotion]);

  return (
    <PulseContext.Provider value={reduceMotion ? null : pulse}>
      <View accessible accessibilityRole="progressbar" accessibilityLabel={label} style={style}>
        {children}
      </View>
    </PulseContext.Provider>
  );
}

export function Skeleton({
  width = '100%',
  height,
  radius = R.xs,
  style,
}: {
  width?: DimensionValue;
  height: number;
  radius?: number;
  style?: StyleProp<ViewStyle>;
}) {
  const pulse = useContext(PulseContext);
  return (
    <Animated.View
      accessibilityElementsHidden
      importantForAccessibility="no-hide-descendants"
      style={[
        { width, height, borderRadius: radius, backgroundColor: C.surfaceHigh, opacity: pulse ?? STATIC_OPACITY },
        style,
      ]}
    />
  );
}
