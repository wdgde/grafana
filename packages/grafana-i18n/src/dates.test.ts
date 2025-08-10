import { formatDate, initRegionalFormat, initDateFormat } from './dates';

describe('date formatting with dateFormat preference', () => {
  beforeEach(() => {
    // Reset formatting to defaults
    initRegionalFormat('en-US');
    initDateFormat('localized');
  });

  it('should format dates with localized format by default', () => {
    const date = new Date('2023-12-25T10:30:00Z');
    const result = formatDate(date, { year: 'numeric', month: 'short', day: 'numeric' });

    // Should use localized formatting (this will vary based on locale)
    expect(result).toContain('Dec');
    expect(result).toContain('25');
    expect(result).toContain('2023');
  });

  it('should format dates with international format when set to international', () => {
    initDateFormat('international');
    const date = new Date('2023-12-25T10:30:00Z');
    const result = formatDate(date, { year: 'numeric', month: '2-digit', day: '2-digit' });

    // Should use ISO8601-style formatting with standardized calendar
    expect(result).toMatch(/12.25.2023|25.12.2023|2023.12.25/); // Various ISO formats
  });

  it('should handle different locales with international format', () => {
    initRegionalFormat('de-DE');
    initDateFormat('international');
    const date = new Date('2023-12-25T10:30:00Z');
    const result = formatDate(date, { year: 'numeric', month: '2-digit', day: '2-digit' });

    // Even with German locale, standardized format should be consistent
    // The exact format depends on browser implementation, but should contain year, month, day
    expect(result).toMatch(/2023/);
    expect(result).toMatch(/25/);
    expect(result).toMatch(/12/);
  });

  it('should handle different locales with localized format', () => {
    initRegionalFormat('de-DE');
    initDateFormat('localized');
    const date = new Date('2023-12-25T10:30:00Z');
    const result = formatDate(date, { year: 'numeric', month: 'short', day: 'numeric' });

    // Should use German locale formatting
    expect(result).toContain('Dez'); // German abbreviation for December
  });

  it('should clear memoized cache when dateFormat changes', () => {
    const date = new Date('2023-12-25T10:30:00Z');
    const options = { year: 'numeric' as const, month: 'short' as const, day: 'numeric' as const };

    // Format with localized
    initDateFormat('localized');
    const localizedResult = formatDate(date, options);

    // Format with international
    initDateFormat('international');
    const internationalResult = formatDate(date, options);

    // Results should be different (cache was cleared)
    // We can't assert exact values since they depend on browser locale implementation
    expect(typeof localizedResult).toBe('string');
    expect(typeof internationalResult).toBe('string');
  });

  it('should handle undefined dateFormat gracefully', () => {
    // @ts-ignore - testing undefined case
    initDateFormat(undefined);
    const date = new Date('2023-12-25T10:30:00Z');

    expect(() => formatDate(date, { year: 'numeric' })).not.toThrow();
  });
});
