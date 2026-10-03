import { colors } from './colors';

const styleElementId = 'racecontrol-web-styles';

const css = `
input:-webkit-autofill,
input:-webkit-autofill:hover,
input:-webkit-autofill:focus,
input:-webkit-autofill:active {
  -webkit-box-shadow: 0 0 0 1000px ${colors.surfaceRaised} inset !important;
  box-shadow: 0 0 0 1000px ${colors.surfaceRaised} inset !important;
  -webkit-text-fill-color: ${colors.text} !important;
  caret-color: ${colors.text};
  transition: background-color 9999s ease-out 0s;
}
input::placeholder { color: ${colors.textMuted}; }
input::-ms-reveal, input::-ms-clear { display: none; }
html, body { background-color: ${colors.background}; color-scheme: dark; }
`;

export function installWebStyles() {
  if (typeof document === 'undefined' || document.getElementById(styleElementId)) {
    return;
  }
  const style = document.createElement('style');
  style.id = styleElementId;
  style.textContent = css;
  document.head.appendChild(style);
}
