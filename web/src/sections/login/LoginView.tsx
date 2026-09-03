import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';

import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Card from '@mui/material/Card';
import Stack from '@mui/material/Stack';
import Select from '@mui/material/Select';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import { alpha, useTheme } from '@mui/material/styles';
import InputAdornment from '@mui/material/InputAdornment';

import { useRouter } from 'src/routes/hooks';

import useFetchAndStoreSaltAuth from 'src/hooks/auth/useFetchAndStoreSaltAuth.ts';

import { bgGradient } from 'src/theme/css';
import { queryKeys } from 'src/api/queryKeys.ts';
import { sessionStore } from 'src/api/session.ts';
import { login, getSession, getAuthMethods } from 'src/api/auth.ts';
import { Version, GetStartedURL, ForgotPasswordURL } from 'src/config.ts';

import Logo from 'src/components/logo';
import Iconify from 'src/components/iconify';

const LoginView: React.FC = () => {
  const theme = useTheme();
  const router = useRouter();
  const queryClient = useQueryClient();
  const [showPassword, setShowPassword] = useState(false);
  const [selectedAuthMethod, setSelectedAuthMethod] = useState('');
  const [loginError, setLoginError] = useState('');
  const { postSaltAuth } = useFetchAndStoreSaltAuth();
  const { data: authMethods = [] } = useQuery({
    queryKey: queryKeys.auth.methods(),
    queryFn: ({ signal }) => getAuthMethods(signal),
    select: (methods) => {
      const order = ['ldap', 'local', 'cas'];
      return [...methods].sort((a, b) => order.indexOf(a) - order.indexOf(b));
    },
  });

  const effectiveAuthMethod = selectedAuthMethod || authMethods[0] || '';

  const handleClick = async () => {
    const username = (document.querySelector('input[name="username"]') as HTMLInputElement).value;
    const password = (document.querySelector('input[name="password"]') as HTMLInputElement).value;

    try {
      setLoginError('');
      await login({
        username,
        password,
        method: effectiveAuthMethod,
      });
      const authUser = await queryClient.fetchQuery({
        queryKey: queryKeys.auth.session(),
        queryFn: ({ signal }) => getSession(signal),
        staleTime: 0,
      });
      sessionStore.setAuthenticated(authUser);
      await postSaltAuth();

      // Redirect to the home page
      router.push('/');
    } catch (err) {
      console.error('Login error:', err);
      setLoginError('Unable to sign in. Check your credentials and try again.');
    }
  };

  const renderForm = (
    <>
      <Stack spacing={3}>
        <TextField name="username" label="Username" />

        <TextField
          name="password"
          label="Password"
          type={showPassword ? 'text' : 'password'}
          InputProps={{
            endAdornment: (
              <InputAdornment position="end">
                <IconButton onClick={() => setShowPassword(!showPassword)} edge="end">
                  <Iconify icon={showPassword ? 'eva:eye-fill' : 'eva:eye-off-fill'} />
                </IconButton>
              </InputAdornment>
            ),
          }}
        />

        <Select
          value={effectiveAuthMethod}
          onChange={(e) => setSelectedAuthMethod(e.target.value)}
          displayEmpty
          inputProps={{ 'aria-label': 'Select Auth Method' }}
        >
          {authMethods.map((method) => (
            <MenuItem key={method} value={method}>
              {method}
            </MenuItem>
          ))}
        </Select>
      </Stack>

      <Stack direction="row" alignItems="center" justifyContent="flex-end" sx={{ my: 3 }}>
        <Link variant="subtitle2" underline="hover" href={ForgotPasswordURL}>
          Forgot password?
        </Link>
      </Stack>

      <Button
        fullWidth
        size="large"
        type="submit"
        variant="contained"
        color="inherit"
        onClick={handleClick}
      >
        Login
      </Button>
      {loginError && (
        <Typography role="alert" color="error" sx={{ mt: 2 }}>
          {loginError}
        </Typography>
      )}
    </>
  );

  return (
    <Box
      sx={{
        ...bgGradient({
          color: alpha(theme.palette.background.default, 0.9),
          imgUrl: '/assets/background/overlay_4.jpg',
        }),
        height: 1,
      }}
    >
      <Logo
        sx={{
          position: 'fixed',
          top: { xs: 16, md: 24 },
          left: { xs: 16, md: 24 },
        }}
      />

      <Stack alignItems="center" justifyContent="center" sx={{ height: 1 }}>
        <Card
          sx={{
            p: 5,
            width: 1,
            maxWidth: 420,
          }}
        >
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Typography variant="h4">Sign in to Agartha</Typography>
            <img src="/assets/logo.svg" alt="Agartha logo" style={{ width: 96, height: 96 }} />
          </Stack>

          <Typography variant="body2" sx={{ mt: 2, mb: 5 }}>
            Don’t have a login?
            <Link variant="subtitle2" sx={{ ml: 0.5 }} href={GetStartedURL}>
              Get started
            </Link>
          </Typography>

          <Divider sx={{ my: 3 }}>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              OR
            </Typography>
          </Divider>

          {renderForm}

          <Divider sx={{ my: 3 }} />

          <Typography variant="body2" sx={{ color: 'text.secondary' }}>
            Version: {Version}
          </Typography>
        </Card>
      </Stack>
    </Box>
  );
};

export default LoginView;
