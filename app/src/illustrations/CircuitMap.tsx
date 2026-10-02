import Svg, { Circle, Path } from 'react-native-svg';

export type CircuitKey = 'cota' | 'marina_bay' | 'monza' | 'lusail' | 'las_vegas' | 'generic';

const circuitPaths: Record<CircuitKey, { path: string; start: [number, number] }> = {
  cota: {
    path: 'M28 96 L62 96 Q70 96 72 88 L76 72 Q78 64 86 68 L98 76 Q104 80 110 74 L118 64 Q124 58 130 64 L138 72 Q144 78 152 70 L176 40 Q182 32 174 26 L160 22 Q150 20 146 28 L140 40 Q136 48 128 46 L112 40 Q104 38 100 44 L94 54 Q90 60 82 56 L70 50 Q60 46 54 54 L40 74 Q34 82 26 84 Q18 88 22 94 Q24 96 28 96 Z',
    start: [44, 96],
  },
  marina_bay: {
    path: 'M24 92 L24 40 Q24 30 34 30 L70 30 L76 22 L120 22 Q130 22 130 32 L130 44 L164 44 Q176 44 176 56 L176 70 Q176 80 166 80 L140 80 L132 92 L96 92 L90 100 L34 100 Q24 100 24 92 Z',
    start: [24, 70],
  },
  monza: {
    path: 'M30 86 L150 86 Q176 86 178 62 Q180 40 160 34 L140 30 L134 36 L120 30 L60 26 Q40 26 34 40 L30 50 L38 56 L30 62 Q22 72 24 80 Q26 86 30 86 Z',
    start: [90, 86],
  },
  lusail: {
    path: 'M40 90 Q20 90 22 70 Q24 54 40 50 L70 44 Q80 42 84 34 Q90 22 104 24 L150 30 Q172 34 174 54 Q176 76 156 82 L120 88 Q110 90 104 98 Q98 104 86 100 L60 92 Q50 90 40 90 Z',
    start: [130, 87],
  },
  las_vegas: {
    path: 'M26 34 L150 34 L162 46 L176 46 L176 66 L150 66 L150 88 L60 88 Q50 88 46 80 L36 62 L26 62 Z',
    start: [100, 34],
  },
  generic: {
    path: 'M30 80 Q20 60 40 44 L80 30 Q110 20 140 30 L170 48 Q180 60 166 72 L130 84 Q100 94 70 90 Z',
    start: [100, 91],
  },
};

const circuitByName: Record<string, CircuitKey> = {
  'Circuit of the Americas': 'cota',
  'Marina Bay': 'marina_bay',
  Monza: 'monza',
  Lusail: 'lusail',
  'Las Vegas Strip': 'las_vegas',
};

export function circuitKeyFor(circuitName: string): CircuitKey {
  return circuitByName[circuitName] ?? 'generic';
}

type CircuitMapProps = {
  circuit: CircuitKey;
  width?: number;
  color?: string;
  strokeWidth?: number;
  showStart?: boolean;
  halo?: boolean;
};

export function CircuitMap({ circuit, width = 80, color = '#FFFFFF', strokeWidth = 4, showStart = false, halo = false }: CircuitMapProps) {
  const { path, start } = circuitPaths[circuit];
  return (
    <Svg width={width} height={width * 0.6} viewBox="0 0 200 120" accessibilityRole="image">
      {halo ? (
        <>
          <Path d={path} fill="none" stroke="rgba(0,0,0,0.55)" strokeWidth={strokeWidth * 3} strokeLinejoin="round" strokeLinecap="round" />
          <Path d={path} fill="none" stroke="#E10600" strokeWidth={strokeWidth * 1.7} strokeLinejoin="round" strokeLinecap="round" strokeOpacity={0.9} />
        </>
      ) : null}
      <Path d={path} fill="none" stroke={color} strokeWidth={strokeWidth} strokeLinejoin="round" strokeLinecap="round" />
      {showStart ? <Circle cx={start[0]} cy={start[1]} r={strokeWidth * 1.4} fill="#E10600" stroke={color} strokeWidth={strokeWidth / 2} /> : null}
    </Svg>
  );
}
