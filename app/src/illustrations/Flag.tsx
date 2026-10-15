import type { ReactNode } from 'react';
import Svg, { Circle, G, Path, Rect } from 'react-native-svg';

export type FlagCode = string;

const verticalTricolor = (left: string, middle: string, right: string) => (
  <>
    <Rect x={0} y={0} width={10} height={20} fill={left} />
    <Rect x={10} y={0} width={10} height={20} fill={middle} />
    <Rect x={20} y={0} width={10} height={20} fill={right} />
  </>
);

const horizontalBands = (bands: { color: string; height: number }[]) => {
  let offset = 0;
  return bands.map((band, index) => {
    const rect = <Rect key={`${band.color}-${index}`} x={0} y={offset} width={30} height={band.height} fill={band.color} />;
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

const usFlag = (
  <>
    <Rect x={0} y={0} width={30} height={20} fill="#FFFFFF" />
    {usStripes}
    <Rect x={0} y={0} width={12} height={10.8} fill="#3C3B6E" />
    {usStars}
  </>
);

const qatarSerration = Array.from({ length: 9 }, (_, index) => {
  const top = (index * 20) / 9;
  const middle = top + 20 / 18;
  return `L10 ${top} L13 ${middle}`;
}).join(' ');

const genericFlag = (
  <>
    <Rect x={0} y={0} width={30} height={20} fill="#2F3543" />
    <Circle cx={15} cy={10} r={5.5} fill="none" stroke="#7F8596" strokeWidth={1} />
    <Path d="M15 4.5 V15.5 M9.5 10 H20.5" stroke="#7F8596" strokeWidth={0.8} />
  </>
);

const flags: Record<string, ReactNode> = {
  us: usFlag,
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
  au: horizontalBands([
    { color: '#012169', height: 20 / 3 },
    { color: '#FFFFFF', height: 20 / 3 },
    { color: '#012169', height: 20 / 3 },
  ]),
  ca: verticalTricolor('#D52B1E', '#FFFFFF', '#D52B1E'),
  mc: verticalTricolor('#CE1126', '#FFFFFF', '#CE1126'),
  jp: (
    <>
      <Rect x={0} y={0} width={30} height={20} fill="#FFFFFF" />
      <Circle cx={15} cy={10} r={5.5} fill="#BC002D" />
    </>
  ),
  cn: (
    <>
      <Rect x={0} y={0} width={30} height={20} fill="#DE2910" />
      <Path d="M4 3 L5.5 7 L2.5 4.5 L5.5 4.5 L2.5 7 Z" fill="#FFDE00" />
    </>
  ),
  bh: (
    <>
      <Rect x={0} y={0} width={30} height={20} fill="#FFFFFF" />
      <Rect x={0} y={0} width={9} height={20} fill="#DA291C" />
      <Path d="M9 0 L13 1.5 L9 3 L13 4.5 L9 6 L13 7.5 L9 9 L13 10.5 L9 12 L13 13.5 L9 15 L13 16.5 L9 18 L13 19.5 L9 20" stroke="#DA291C" strokeWidth={0.8} fill="none" />
    </>
  ),
  sa: (
    <>
      <Rect x={0} y={0} width={30} height={20} fill="#006C35" />
      <Circle cx={15} cy={10} r={3} fill="#FFFFFF" />
    </>
  ),
  be: verticalTricolor('#000000', '#FDDA24', '#EF3340'),
  hu: horizontalBands([
    { color: '#CE2939', height: 20 / 3 },
    { color: '#FFFFFF', height: 20 / 3 },
    { color: '#477050', height: 20 / 3 },
  ]),
  nl: horizontalBands([
    { color: '#AE1C28', height: 20 / 3 },
    { color: '#FFFFFF', height: 20 / 3 },
    { color: '#21468B', height: 20 / 3 },
  ]),
  az: horizontalBands([
    { color: '#00AFCA', height: 20 / 3 },
    { color: '#E30A17', height: 20 / 3 },
    { color: '#009739', height: 20 / 3 },
  ]),
  mx: verticalTricolor('#006847', '#FFFFFF', '#CE1126'),
  br: (
    <>
      <Rect x={0} y={0} width={30} height={20} fill="#009739" />
      <Path d="M15 3 L26 10 L15 17 L4 10 Z" fill="#FEDD00" />
      <Circle cx={15} cy={10} r={4} fill="#002776" />
    </>
  ),
  ae: horizontalBands([
    { color: '#00843D', height: 20 / 3 },
    { color: '#FFFFFF', height: 20 / 3 },
    { color: '#000000', height: 20 / 3 },
  ]),
};

export function Flag({ code, width = 24 }: { code?: string | null; width?: number }) {
  const key = (code ?? '').toLowerCase();
  return (
    <Svg width={width} height={(width * 2) / 3} viewBox="0 0 30 20" accessibilityRole="image">
      {flags[key] ?? genericFlag}
      <Rect x={0.25} y={0.25} width={29.5} height={19.5} rx={1.5} fill="none" stroke="rgba(255,255,255,0.18)" strokeWidth={0.5} />
    </Svg>
  );
}
