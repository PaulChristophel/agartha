declare module 'src/routes/hooks' {
  export const useRouter: () => { push: (path: string) => void };
}

declare module 'src/theme/css' {
  export const bgGradient: (params: { color: string; imgUrl: string }) => object;
}

declare module 'src/components/logo' {
  const Logo: React.FC<{ sx: object }>;
  export default Logo;
}

declare module 'src/components/iconify' {
  const Iconify: React.FC<{ icon: string }>;
  export default Iconify;
}
