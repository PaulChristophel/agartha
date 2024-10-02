make configure-web
DEBUG=vite:* NODE_ENV=development pnpm --dir ./web run build --mode=development
NODE_ENV=development pnpm --dir ./web start --mode=development