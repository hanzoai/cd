import {expect, test as base} from '@playwright/test';

// Each test fails on any uncaught page error or console error, so a UI that
// renders nothing because it threw at load, or renders a page that 404s on its
// own API calls, cannot pass.
const test = base.extend<{signedIn: boolean}>({
    signedIn: [true, {option: true}],
    page: async ({page, context, signedIn, baseURL}, use) => {
        if (signedIn) {
            await context.addCookies([{name: 'cd.token', value: 'smoke', url: baseURL}]);
        }
        const errors: string[] = [];
        page.on('pageerror', e => errors.push(`pageerror: ${e.message}`));
        // React 19 warns once per render that a dependency (argo-ui's dropdowns and
        // tooltips) reads element.ref. It breaks nothing; every other console error
        // still fails the page.
        const ref = /^Accessing element\.ref was removed in React 19/;
        page.on('console', m => m.type() === 'error' && !ref.test(m.text()) && errors.push(`console.error: ${m.text()}`));
        await use(page);
        expect(errors).toEqual([]);
    }
});

test.describe('signed out', () => {
    test.use({signedIn: false});

    test('a page redirects to the login page', async ({page}) => {
        await page.goto('/applications');
        await expect(page).toHaveURL(/\/login\?return_url=/);
        await expect(page.locator('.login__name')).toHaveText('Hanzo CD');
        await expect(page.getByRole('button', {name: 'Log in via Hanzo'})).toBeVisible();
        await expect(page.getByRole('button', {name: 'Sign In'})).toBeVisible();
    });

    test('signing in returns to the page asked for', async ({page}) => {
        await page.goto('/settings/repos');
        await expect(page).toHaveURL(/\/login\?return_url=/);
        await page.locator('input[name=username]').fill('z');
        await page.locator('input[name=password]').fill('smoke');
        await page.getByRole('button', {name: 'Sign In'}).click();
        await expect(page).toHaveURL(/\/settings\/repos(\?|$)/);
        await expect(page.getByText('https://github.com/hanzoai/universe').first()).toBeVisible();
    });
});

test('the root redirects to the applications list', async ({page}) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/applications(\?|$)/);
    await expect(page.getByText('web', {exact: true}).first()).toBeVisible();
    await expect(page.getByText('api', {exact: true}).first()).toBeVisible();
});

test('an application opens to its resource tree', async ({page}) => {
    await page.goto('/applications');
    await page.getByText('web', {exact: true}).first().click();
    await expect(page).toHaveURL(/\/applications\/hanzo-cd\/web(\?|$)/);
    await expect(page.getByText('web-5d9f7c-x7k2p').first()).toBeVisible();
    // A group outside the built-in icon set resolves through the wildcard (minimatch) lookup.
    await expect(page.getByText('servicemonitor', {exact: true}).first()).toBeVisible();
    await page.getByText('web-5d9f7c-x7k2p').first().click();
    await expect(page).toHaveURL(/node=/);
    await expect(page.getByText('Summary', {exact: true}).first()).toBeVisible();
});

test('an application by name alone renders', async ({page}) => {
    await page.goto('/applications/api');
    await expect(page.getByText('api-5d9f7c-x7k2p').first()).toBeVisible();
});

test('the application sets list renders', async ({page}) => {
    await page.goto('/applicationsets');
    await expect(page.getByText('fleet', {exact: true}).first()).toBeVisible();
});

// One test a page: a page is a full load of the bundle, and a runner under load
// takes seconds for each, so all of them in one test ran out its minute.
const settings: [string, string, boolean][] = [
    ['/settings', 'Repositories', true],
    ['/settings/repos', 'https://github.com/hanzoai/universe', false],
    ['/settings/certs', 'github.com', true],
    ['/settings/gpgkeys', '4AEE18F83AFDEB23', false],
    ['/settings/clusters', 'in-cluster', false],
    ['/settings/projects', 'default', true],
    ['/settings/projects/default', 'General', true],
    [`/settings/clusters/${encodeURIComponent('https://kubernetes.default.svc')}`, 'in-cluster', false]
];
for (const [path, text, exact] of settings) {
    test(`settings page ${path} renders`, async ({page}) => {
        await page.goto(path);
        await expect(page.getByText(text, {exact}).first()).toBeVisible();
    });
}

test('an unknown settings page returns to settings', async ({page}) => {
    await page.goto('/settings/nowhere');
    await expect(page).toHaveURL(/\/settings(\?|$)/);
});

test('the sidebar navigates between sections', async ({page}) => {
    await page.goto('/applications');
    await expect(page.getByText('web', {exact: true}).first()).toBeVisible();
    await page.locator('.sidebar').getByText('Settings', {exact: true}).click();
    await expect(page).toHaveURL(/\/settings(\?|$)/);
    await expect(page.getByText('Repositories', {exact: true}).first()).toBeVisible();
    await page.locator('.sidebar').getByText('User info', {exact: true}).click();
    await expect(page).toHaveURL(/\/user-info(\?|$)/);
    await expect(page.getByText('Username: z')).toBeVisible();
});

test('every font the page uses loads', async ({page}) => {
    await page.goto('/applications');
    await expect(page.getByText('web', {exact: true}).first()).toBeVisible();
    const fonts = await page.evaluate(async () => {
        await document.fonts.ready;
        return [...document.fonts].filter(f => f.status !== 'unloaded').map(f => `${f.family} ${f.status}`);
    });
    expect(fonts.filter(f => f.endsWith(' error'))).toEqual([]);
    expect(fonts).toContain('Font Awesome 7 Free loaded');
});

test('the sidebar carries one mark, the product name and the version', async ({page}) => {
    await page.goto('/applications');
    await expect(page.locator('.sidebar__name')).toHaveText('Hanzo CD');
    await expect(page.locator('.sidebar .hanzo-mark')).toHaveCount(1);
    await expect(page.locator('.sidebar img')).toHaveCount(0);
    await expect(page.locator('.sidebar__version')).toHaveText('v0.0.0-smoke');
});

test('an unstamped build shows no version', async ({page}) => {
    await page.route('**/v1/version', route => route.fulfill({json: {Version: 'v3.7.0+unknown'}}));
    await page.goto('/applications');
    await expect(page.locator('.sidebar__name')).toHaveText('Hanzo CD');
    await expect(page.locator('.sidebar__version')).toHaveCount(0);
});

test('no applications reads as one line and a quiet action', async ({page}) => {
    await page.route('**/v1/applications?*', route => route.fulfill({json: {metadata: {resourceVersion: '7'}, items: []}}));
    await page.route('**/v1/applications', route => route.fulfill({json: {metadata: {resourceVersion: '7'}, items: []}}));
    await page.goto('/applications');
    await expect(page.getByText('No applications you can access', {exact: true})).toBeVisible();
    await expect(page.locator('.empty-state h5')).toHaveCount(0);
    await expect(page.getByRole('button', {name: 'New application'})).toHaveClass(/kit-button--base-o/);
});

test.describe('on a phone', () => {
    test.use({viewport: {width: 390, height: 844}});

    test('the sidebar is a rail that opens over the page', async ({page}) => {
        await page.goto('/applications');
        await expect(page.getByText('web', {exact: true}).first()).toBeVisible();
        await expect(page.locator('.sidebar')).toHaveClass(/sidebar--collapsed/);
        const width = await page.evaluate(() => document.documentElement.scrollWidth);
        expect(width).toBeLessThanOrEqual(390);
        await page.getByRole('button', {name: 'Expand sidebar'}).click();
        await expect(page.locator('.sidebar__name')).toBeVisible();
        await page.locator('.sidebar').getByText('Settings', {exact: true}).click();
        await expect(page).toHaveURL(/\/settings(\?|$)/);
        await expect(page.locator('.sidebar')).toHaveClass(/sidebar--collapsed/);
    });
});

test('a filter suggestion opens over the sidebar and applies', async ({page}) => {
    await page.goto('/applications');
    await expect(page.getByText('web', {exact: true}).first()).toBeVisible();
    await page.locator('.sidebar input[placeholder="Labels"]').fill('team');
    await page.locator('.autocomplete__items__item', {hasText: 'team=platform'}).click();
    await expect(page.locator('.sidebar .filter__item--selected', {hasText: 'team=platform'})).toBeVisible();
});

test('with the sidebar collapsed the toolbar starts at the rail', async ({page}) => {
    await page.goto('/applications');
    await page.getByRole('button', {name: 'Collapse sidebar'}).click();
    await expect(page.locator('.sidebar')).toHaveClass(/sidebar--collapsed/);
    const bar = await page.locator('.flex-top-bar').boundingBox();
    const rail = await page.locator('.sidebar').boundingBox();
    expect(bar.x).toBe(rail.width);
});

test('the network view renders', async ({page}) => {
    await page.goto('/applications/hanzo-cd/web?view=network');
    await expect(page.locator('.application-resource-tree')).toBeVisible();
});

// The session ending is a 401 by design, so this one runs without the
// console-error guard.
base('a session that ends lands on the login card in its own colours', async ({page, context, baseURL}) => {
    await context.addCookies([{name: 'cd.token', value: 'smoke', url: baseURL}]);
    await page.addInitScript(() => localStorage.setItem('view_preferences', JSON.stringify({version: 999, theme: 'light'})));
    await page.goto('/applications');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
    await context.clearCookies();
    await page.locator('.sidebar').getByText('ApplicationSets', {exact: true}).click();
    await expect(page).toHaveURL(/\/login\?return_url=/);
    await expect(page.locator('html')).not.toHaveAttribute('data-theme', /.*/);
});
