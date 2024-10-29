import path from 'path';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';
import checker from 'vite-plugin-checker';

export default defineConfig(async ({ mode }) => {
  // Dynamically import tsconfigPaths to avoid `require` issues
  const { default: tsconfigPaths } = await import('vite-tsconfig-paths');

  return {
    plugins: [
      react(),
      tsconfigPaths(),
      checker({
        eslint: {
          lintCommand: 'eslint "./src/**/*.{js,jsx,ts,tsx}"',
        },
      }),
    ],
    resolve: {
      alias: [
        {
          find: /^~(.+)/,
          replacement: path.join(process.cwd(), 'node_modules/$1'),
        },
        {
          find: /^src(.+)/,
          replacement: path.join(process.cwd(), 'src/$1'),
        },
      ],
    },
    server: {
      host: '0.0.0.0',
      port: 3030,
      proxy: {
        '/api/v1': {
          target: 'http://localhost:8080/',
          changeOrigin: true,
        },
        '/version': {
          target: 'http://localhost:8080/',
          changeOrigin: true,
        },
        '/auth': {
          target: 'http://localhost:8080/',
          changeOrigin: true,
        },
        '/docs': {
          target: 'http://localhost:8080/',
          changeOrigin: true,
        },
      },
    },
    preview: {
      host: '0.0.0.0',
      port: 3030,
      proxy: {
        '/api/v1': {
          target: 'http://backend:8080/',
          changeOrigin: true,
        },
        '/version': {
          target: 'http://backend:8080/',
          changeOrigin: true,
        },
        '/auth': {
          target: 'http://backend:8080/',
          changeOrigin: true,
        },
        '/docs': {
          target: 'http://backend:8080/',
          changeOrigin: true,
        },
      },
    },
    build: mode === 'development' ? {
      sourcemap: true,
      minify: false,
      rollupOptions: {
        output: {
          manualChunks(id: string) {
            if (id.includes('node_modules')) return 'vendor';
          },
        },
      },
      esbuild: {
        keepNames: true,
      },
    } : undefined,
  };
});
