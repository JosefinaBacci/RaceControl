import { useId } from 'react';
import Svg, { Circle, Defs, G, LinearGradient, Path, Rect, Stop } from 'react-native-svg';

type RaceCarProps = {
  color: string;
  width?: number;
  showSpeedLines?: boolean;
};

const speedLines = [
  { y: 48, length: 120 },
  { y: 62, length: 80 },
  { y: 76, length: 150 },
  { y: 90, length: 60 },
];

function Wheel({ cx, radius }: { cx: number; radius: number }) {
  return (
    <G>
      <Circle cx={cx} cy={92} r={radius} fill="#0A0A0C" />
      <Circle cx={cx} cy={92} r={radius - 5} fill="#16161B" stroke="#34343E" strokeWidth={2} />
      <Circle cx={cx} cy={92} r={radius * 0.32} fill="#26262F" />
      <Path d={`M${cx - radius + 4} 92 A${radius - 4} ${radius - 4} 0 0 1 ${cx} ${92 - radius + 4}`} stroke="#E10600" strokeWidth={2.5} fill="none" />
    </G>
  );
}

export function RaceCar({ color, width = 360, showSpeedLines = false }: RaceCarProps) {
  const offset = showSpeedLines ? 150 : 0;
  const gradientId = useId().replace(/[^a-zA-Z0-9]/g, '');
  const bodyFill = `body${gradientId}`;
  const speedFill = `speed${gradientId}`;
  return (
    <Svg width={width} height={(width * 130) / (420 + offset)} viewBox={`0 0 ${420 + offset} 130`} accessibilityRole="image" accessibilityLabel="Auto de carreras">
      <Defs>
        <LinearGradient id={bodyFill} x1="0" y1="0" x2="0" y2="1">
          <Stop offset="0" stopColor={color} stopOpacity={1} />
          <Stop offset="1" stopColor={color} stopOpacity={0.62} />
        </LinearGradient>
        <LinearGradient id={speedFill} x1="0" y1="0" x2="1" y2="0">
          <Stop offset="0" stopColor="#FFFFFF" stopOpacity={0.55} />
          <Stop offset="1" stopColor="#FFFFFF" stopOpacity={0} />
        </LinearGradient>
      </Defs>

      {showSpeedLines
        ? speedLines.map((line) => (
            <Rect key={line.y} x={404} y={line.y} width={line.length} height={2.5} rx={1.2} fill={`url(#${speedFill})`} />
          ))
        : null}

      <Rect x={60} y={98} width={300} height={8} rx={3} fill="#0A0A0C" />
      <Path d="M2 94 L58 94 L58 101 L2 101 Z" fill="#0A0A0C" />
      <Path d="M4 80 L12 80 L12 101 L4 101 Z" fill={color} />
      <Path
        d="M10 92 L64 84 L130 74 L166 66 L186 56 L212 42 L240 42 L252 56 L300 62 L340 66 L352 92 L60 96 Z"
        fill={`url(#${bodyFill})`}
      />
      <Path d="M120 80 L330 72" stroke="#FFFFFF" strokeOpacity={0.55} strokeWidth={3} strokeLinecap="round" />
      <Path d="M216 44 L236 44 L232 56 L214 56 Z" fill="#0A0A0C" />
      <Path d="M150 66 Q182 36 214 50" stroke="#0A0A0C" strokeWidth={6} fill="none" strokeLinecap="round" />
      <Circle cx={186} cy={56} r={10} fill="#F5F5F7" />
      <Path d="M178 54 L196 54" stroke="#0A0A0C" strokeWidth={4} strokeLinecap="round" />
      <Rect x={334} y={26} width={58} height={9} rx={2} fill="#0A0A0C" />
      <Rect x={340} y={40} width={50} height={6} rx={2} fill={color} />
      <Rect x={384} y={24} width={9} height={70} rx={2} fill="#0A0A0C" />
      <Path d="M340 60 L356 44" stroke="#0A0A0C" strokeWidth={5} strokeLinecap="round" />

      <Wheel cx={96} radius={26} />
      <Wheel cx={318} radius={28} />
    </Svg>
  );
}
