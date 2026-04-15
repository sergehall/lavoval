import { defineConfig } from 'vitest/config';
import path from 'node:path';

export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./vitest.setup.ts']
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@lavoval/registry': path.resolve(__dirname, '../../packages/registry/src/index.ts'),
      '@lavoval/contracts': path.resolve(__dirname, '../../packages/contracts/src/index.ts'),
      '@lavoval/engine': path.resolve(__dirname, '../../packages/engine/src/index.ts'),
      '@lavoval/sdk': path.resolve(__dirname, '../../packages/sdk/src/index.ts'),
    }
  }
});
