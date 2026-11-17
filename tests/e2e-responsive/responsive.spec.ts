import { test, expect } from '@playwright/test';

// P1-19: 8-viewport responsive sweep over the public web surface.
// Fails on horizontal overflow; key content must stay reachable.
const WIDTHS = [360, 390, 768, 1024, 1280, 1440, 1920, 2560];
const PAGES = [
  { path: '/', name: 'home' },
  { path: '/search?q=cloud', name: 'search' },
  { path: '/watch/00000000-0000-0000-0000-000000000001', name: 'watch' },
];

test.describe('Responsive 8-width sweep (P1-19)', () => {
  for (const width of WIDTHS) {
    for (const page of PAGES) {
      test(`${page.name} has no horizontal overflow at ${width}px`, async ({ browser }) => {
        const context = await browser.newContext({
          viewport: { width, height: 900 },
          reducedMotion: 'reduce',
        });
        const pg = await context.newPage();
        await pg.goto(page.path, { waitUntil: 'networkidle' });
        const overflow = await pg.evaluate(() => {
          const doc = document.documentElement;
          return doc.scrollWidth - doc.clientWidth;
        });
        expect(
          overflow,
          `${page.name} overflows by ${overflow}px at ${width}px`
        ).toBeLessThanOrEqual(1);
        await context.close();
      });
    }
  }

  test('watch player exposes keyboard-operable seek slider', async ({ browser }) => {
    const context = await browser.newContext({ viewport: { width: 1280, height: 900 } });
    const pg = await context.newPage();
    await pg.goto('/watch/00000000-0000-0000-0000-000000000001', { waitUntil: 'networkidle' });
    const slider = pg.getByRole('slider', { name: 'Seek' });
    await expect(slider).toBeVisible();
    await slider.focus();
    await expect(slider).toBeFocused();
    await context.close();
  });
});

test.describe('Watch render stability (P1-19)', () => {
  test('unknown video id renders the error block instead of SSR 500', async ({ browser }) => {
    const context = await browser.newContext({ viewport: { width: 1280, height: 900 } });
    const pg = await context.newPage();
    const resp = await pg.goto('/watch/99999999-9999-9999-9999-999999999999', { waitUntil: 'networkidle' });
    expect(resp?.status()).not.toBe(500);
    await expect(pg.getByRole('alert').filter({ hasText: '동영상을 불러오지 못했습니다.' })).toBeVisible();
    await expect(pg.getByRole('slider', { name: 'Seek' })).toHaveCount(0);
    await context.close();
  });
});
