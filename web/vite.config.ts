import path from 'path';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';
import checker from 'vite-plugin-checker';
import tsconfigPaths from 'vite-tsconfig-paths';

// ----------------------------------------------------------------------

export default defineConfig({
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
  // build: {
  //   rollupOptions: {
  //     output: {
  //       manualChunks(id) {
  //         if (id.includes('node_modules')) {
  //           if (id.includes('react')) {
  //             return 'react-vendor';
  //           }
  //           if (id.includes('@mui')) {
  //             return 'mui-vendor';
  //           }
  //           if (id.includes('codemirror')) {
  //             return 'codemirror-vendor';
  //           }
  //           if (id.includes('xterm')) {
  //             return 'xterm-vendor';
  //           }
  //           if (id.includes('lodash')) {
  //             return 'lodash-vendor';
  //           }
  //           if (id.includes('axios')) {
  //             return 'axios-vendor';
  //           }
  //           if (id.includes('date-fns')) {
  //             return 'date-fns-vendor';
  //           }
  //           if (id.includes('jwt-decode')) {
  //             return 'jwt-decode-vendor';
  //           }
  //           if (id.includes('react-hook-form')) {
  //             return 'react-hook-form-vendor';
  //           }
  //           if (id.includes('react-router-dom')) {
  //             return 'react-router-dom-vendor';
  //           }
  //           if (id.includes('react-syntax-highlighter')) {
  //             return 'react-syntax-highlighter-vendor';
  //           }
  //           return 'vendor';
  //         }
  //         if (id.includes('src/sections')) {
  //           return 'sections';
  //         }
  //         if (id.includes('src/components')) {
  //           return 'components';
  //         }
  //         if (id.includes('src/hooks')) {
  //           return 'hooks';
  //         }
  //         if (id.includes('src/pages')) {
  //           return 'pages';
  //         }
  //         return 'common';
  //       },
  //       chunkFileNames: 'chunks/[name]-[hash].js',
  //     },
  //   },
  //   chunkSizeWarningLimit: 500,
  // },
});
