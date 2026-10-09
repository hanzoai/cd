import {defineConfig} from '@playwright/test';

// The built UI (dist/app) against smoke/stub.mjs, which answers /v1 from fixtures.
const port = 4173;

export default defineConfig({
    testDir: '.',
    testMatch: 'smoke.ts',
    timeout: 60_000,
    retries: 0,
    workers: 1,
    reporter: 'list',
    use: {baseURL: `http://127.0.0.1:${port}`, browserName: 'chromium', trace: 'retain-on-failure'},
    webServer: {command: `node stub.mjs ${port}`, url: `http://127.0.0.1:${port}/v1/settings`, reuseExistingServer: false, stdout: 'pipe', stderr: 'pipe'}
});
