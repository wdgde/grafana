import { css } from '@emotion/css';
import { useMemo, useRef } from 'react';

import {
  GrafanaTheme2,
  classicColors,
  colorManipulator,
  getColorByStringHash,
  fieldColorModeRegistry,
  Field,
} from '@grafana/data';

import { useStyles2 } from '../../../../themes/ThemeContext';
import { TableCellRendererProps } from '../types';

interface Pill {
  value: string;
  key: string;
  bgColor: string;
  color: string;
}

const WHITE = '#FFFFFF' as const;
const BLACK = '#000000' as const;
const LUMINANCE_THRESHOLD = 4.5; // WCAG 2.0 AA contrast ratio threshold for text readability

type ContrastRatioCache = Record<string, typeof WHITE | typeof BLACK>;

function getTextColor(bgColor: string, contrastRatios: ContrastRatioCache): string {
  if (contrastRatios[bgColor] === undefined) {
    contrastRatios[bgColor] = colorManipulator.getContrastRatio(WHITE, bgColor) >= LUMINANCE_THRESHOLD ? WHITE : BLACK;
  }
  return contrastRatios[bgColor];
}

function createPills(pillValues: string[], colors: string[], contrastRatios: ContrastRatioCache): Pill[] {
  return pillValues.map((pill, index) => {
    const bgColor = getColorByStringHash(colors, pill);
    const textColor = getTextColor(bgColor, contrastRatios);
    return {
      value: pill,
      key: `${pill}-${index}`,
      bgColor,
      color: textColor,
    };
  });
}

export function PillCell({ value, field, theme }: TableCellRendererProps) {
  // cache luminance calculations to avoid recalculating the same color for the same background.
  const contrastRatios = useRef<ContrastRatioCache>({});
  const styles = useStyles2(getStyles);

  const pillValues = useMemo(() => inferPills(String(value)), [value]);
  const colors = useMemo(() => getColors(field, theme, pillValues), [field, theme, pillValues]);
  const pills = useMemo(
    () => createPills(pillValues, colors, contrastRatios.current ?? {}),
    [colors, pillValues, contrastRatios]
  );

  return pills.map((pill) => (
    <span
      key={pill.key}
      className={styles.pill}
      style={{
        backgroundColor: pill.bgColor,
        color: pill.color,
      }}
    >
      {pill.value}
    </span>
  ));
}

const SPLIT_RE = /\s*,\s*/;

function inferPills(value: string): string[] {
  if (value === '') {
    return [];
  }

  if (value[0] === '[') {
    try {
      return JSON.parse(value);
    } catch {
      return value.trim().split(SPLIT_RE);
    }
  }

  return value.trim().split(SPLIT_RE);
}

function getColors(field: Field, theme: GrafanaTheme2, pillValues: string[]): string[] {
  let colors = classicColors;
  const configuredColor = field.config.color;
  if (configuredColor) {
    const mode = fieldColorModeRegistry.get(configuredColor.mode);
    if (mode) {
      if (mode.getColors) {
        colors = mode.getColors(theme);
      } else {
        colors = [];
        // only generate a maximum of 20 colors for a set of pills.
        for (let i = 0; i < Math.min(pillValues.length, 20); i++) {
          // spoof the series index changing to get a range of colors from the color mode for these pills.
          colors[i] = mode.getCalculator({ ...field, state: { ...field.state, seriesIndex: i } }, theme)(0, 0);
        }
      }
    }
  }
  return colors;
}

const getStyles = (theme: GrafanaTheme2) => ({
  pill: css({
    display: 'inline-block',
    padding: theme.spacing(0.25, 0.75),
    marginInlineEnd: theme.spacing(0.25),
    marginBlock: theme.spacing(0.25),
    borderRadius: theme.shape.radius.default,
    fontSize: theme.typography.bodySmall.fontSize,
    lineHeight: theme.typography.bodySmall.lineHeight,
    whiteSpace: 'nowrap',
  }),
});
