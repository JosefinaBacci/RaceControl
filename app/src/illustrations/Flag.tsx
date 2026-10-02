import type { ReactNode } from 'react';
import Svg, { Circle, G, Path, Rect } from 'react-native-svg';

export type FlagCode = 'us' | 'sg' | 'it' | 'qa' | 'at' | 'de' | 'gb' | 'fr' | 'es';

const verticalTricolor = (left: string, middle: string, right: string) => (
  <>
    <Rect x={0} y={0} width={10} height={20} fill={left} />
    <Rect x={10} y={0} width={10} height={20} fill={middle} />
    <Rect x={20} y={0} width={10} height={20} fill={right} />
  </>
);

const horizontalBands = (bands: { color: string; height: number }[]) => {
  let offset = 0;
  return bands.map((band) => {
    const rect = <Rect key={`${band.color}-${offset}`} x={0} y={offset} width={30} height={band.height} fill={band.color} />;
    offset += band.height;
    return rect;
  });
};

const usStripes = Array.from({ length: 7 }, (_, index) => (
  <Rect key={index} x={0} y={index * (20 / 6.5)} width={30} height={20 / 13} fill="#B22234" />
));

const usStars = Array.from({ length: 12 }, (_, index) => (
  <Circle key={index} cx={1.6 + (index % 4) * 3} cy={1.6 + Math.floor(index / 4) * 3.4} r={0.55} fill="#FFFFFF" />
));

const qatarSerration = Array.from({ length: 9 }, (_, index) => {
  const top = (index * 20) / 9;
  const middle = top + 20 / 18;
  return `L10 ${top} L13 ${middle}`;
}).join(' ');

const flags: Record<FlagCode, ReactNode> = {
  us: (
    <>
      <Rect x={0} y={0} width={30} height={20} fill="#FFFFFF" />
      {usStripes}
      <Rect x={0} y={0} width={12} height={10.8} fill="#3C3B6E" />
      {usStars}
    </>
  ),
  sg: (
    <>
      <Rect x={0} y={0} width={30} height={10} fill="#EF3340" />
      <Rect x={0} y={10} width={30} height={10} fill="#FFFFFF" />
      <Circle cx={6} cy={5} r={3.4} fill="#FFFFFF" />
      <Circle cx={7.3} cy={5} r={3.2} fill="#EF3340" />
      <G fill="#FFFFFF">
        <Circle cx={9.6} cy={3} r={0.55} />
        <Circle cx={11.4} cy={4.2} r={0.55} />
        <Circle cx={10.8} cy={6.3} r={0.55} />
        <Circle cx={8.6} cy={6.3} r={0.55} />
        <Circle cx={8.1} cy={4.2} r={0.55} />
      </G>
    </>
  ),
  it: verticalTricolor('#009246', '#FFFFFF', '#CE2B37'),
  fr: verticalTricolor('#0055A4', '#FFFFFF', '#EF4135'),
  qa: (
    <>
      <Rect x={0} y={0} width={30} height={20} fill="#8A1538" />
      <Path d={`M0 0 ${qatarSerration} L10 20 L0 20 Z`} fill="#FFFFFF" />
    </>
  ),
  at: horizontalBands([
    { color: '#ED2939', height: 20 / 3 },
    { color: '#FFFFFF', height: 20 / 3 },
    { color: '#ED2939', height: 20 / 3 },
  ]),
  de: horizontalBands([
    { color: '#000000', height: 20 / 3 },
    { color: '#DD0000', height: 20 / 3 },
    { color: '#FFCE00', height: 20 / 3 },
  ]),
  es: horizontalBands([
    { color: '#AA151B', height: 5 },
    { color: '#F1BF00', height: 10 },
    { color: '#AA151B', height: 5 },
  ]),
  gb: (
    <>
      <Rect x={0} y={0} width={30} height={20} fill="#012169" />
      <Path d="M0 0 L30 20 M30 0 L0 20" stroke="#FFFFFF" strokeWidth={4} />
      <Path d="M0 0 L30 20 M30 0 L0 20" stroke="#C8102E" strokeWidth={1.4} />
      <Path d="M15 0 V20 M0 10 H30" stroke="#FFFFFF" strokeWidth={6} />
      <Path d="M15 0 V20 M0 10 H30" stroke="#C8102E" strokeWidth={3.4} />
    </>
  ),
};

export function Flag({ code, width = 24 }: { code: FlagCode; width?: number }) {
  return (
    <Svg width={width} height={(width * 2) / 3} viewBox="0 0 30 20" accessibilityRole="image">
      {flags[code]}
      <Rect x={0.25} y={0.25} width={29.5} height={19.5} rx={1.5} fill="none" stroke="rgba(255,255,255,0.18)" strokeWidth={0.5} />
    </Svg>
  );
}
