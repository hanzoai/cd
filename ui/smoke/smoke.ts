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
    await expect(page.getByText('SUMMARY', {exact: true}).first()).toBeVisible();
});

test('an application by name alone renders', async ({page}) => {
    await page.goto('/applications/api');
    await expect(page.getByText('api-5d9f7c-x7k2p').first()).toBeVisible();
});

test('the application sets list renders', async ({page}) => {
    await page.goto('/applicationsets');
    await expect(page.getByText('fleet', {exact: true}).first()).toBeVisible();
});

test('settings pages render their lists', async ({page}) => {
    await page.goto('/settings');
    await expect(page.getByText('Repositories', {exact: true}).first()).toBeVisible();
    await page.goto('/settings/repos');
    await expect(page.getByText('https://github.com/hanzoai/universe').first()).toBeVisible();
    await page.goto('/settings/certs');
    await expect(page.getByText('github.com', {exact: true}).first()).toBeVisible();
    await page.goto('/settings/gpgkeys');
    await expect(page.getByText('4AEE18F83AFDEB23').first()).toBeVisible();
    await page.goto('/settings/clusters');
    await expect(page.getByText('in-cluster').first()).toBeVisible();
    await page.goto('/settings/projects');
    await expect(page.getByText('default', {exact: true}).first()).toBeVisible();
    await page.goto('/settings/projects/default');
    await expect(page.getByText('GENERAL', {exact: true}).first()).toBeVisible();
    await page.goto(`/settings/clusters/${encodeURIComponent('https://kubernetes.default.svc')}`);
    await expect(page.getByText('in-cluster').first()).toBeVisible();
    await page.goto('/settings/nowhere');
    await expect(page).toHaveURL(/\/settings(\?|$)/);
});

test('the sidebar navigates between sections', async ({page}) => {
    await page.goto('/applications');
    await expect(page.getByText('web', {exact: true}).first()).toBeVisible();
    await page.locator('.sidebar').getByText('Settings', {exact: true}).click();
    await expect(page).toHaveURL(/\/settings(\?|$)/);
    await expect(page.getByText('Repositories', {exact: true}).first()).toBeVisible();
    await page.locator('.sidebar').getByText('User Info', {exact: true}).click();
    await expect(page).toHaveURL(/\/user-info(\?|$)/);
    await expect(page.getByText('Username: z')).toBeVisible();
});
