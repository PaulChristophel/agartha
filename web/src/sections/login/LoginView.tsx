import axios from 'axios';
import { useState, useEffect } from 'react';

import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Card from '@mui/material/Card';
import Stack from '@mui/material/Stack';
import Select from '@mui/material/Select';
import Divider from '@mui/material/Divider';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import LoadingButton from '@mui/lab/LoadingButton';
import { alpha, useTheme } from '@mui/material/styles';
import InputAdornment from '@mui/material/InputAdornment';

import { useRouter } from 'src/routes/hooks';

import { fetchAndStoreAuthUser } from 'src/hooks/auth/fetchAndStoreAuthUser.ts';
import useFetchAndStoreSaltAuth from 'src/hooks/auth/useFetchAndStoreSaltAuth.ts';

import { bgGradient } from 'src/theme/css';
import { Version, GetStartedURL, ForgotPasswordURL } from 'src/config.ts';

import Logo from 'src/components/logo';
import Iconify from 'src/components/iconify';

const LoginView: React.FC = () => {
  const theme = useTheme();
  const router = useRouter();
  const [showPassword, setShowPassword] = useState(false);
  const [authMethods, setAuthMethods] = useState<string[]>([]);
  const [selectedAuthMethod, setSelectedAuthMethod] = useState('');
  const { postSaltAuth } = useFetchAndStoreSaltAuth();

  useEffect(() => {
    const fetchAuthMethods = async () => {
      try {
        const response = await axios.get('/auth/method');
        const methods = response.data.auth_methods;
        // Set the order of precedence:
        const orderedMethods = methods.sort((a: string, b: string) => {
          const order = ['ldap', 'local', 'cas']; // We put cas last because eventually it will be a redirect
          return order.indexOf(a) - order.indexOf(b);
        });
        setAuthMethods(orderedMethods);
        setSelectedAuthMethod(orderedMethods[0]);
      } catch (err) {
        console.error('Error fetching auth methods:', err);
      }
    };

    fetchAuthMethods();
  }, []);

  const handleClick = async () => {
    const username = (document.querySelector('input[name="username"]') as HTMLInputElement).value;
    const password = (document.querySelector('input[name="password"]') as HTMLInputElement).value;

    try {
      const response = await axios.post('/auth/token', {
        username,
        password,
        method: selectedAuthMethod,
      });
      const { token } = response.data;
      // console.log('Token:', token);

      // Store the token in localStorage
      localStorage.setItem('authToken', token);

      // Fetch and store user settings
      await fetchAndStoreAuthUser(token);
      await postSaltAuth(token);

      // Redirect to the home page
      router.push('/');
    } catch (err) {
      console.error('Login error:', err);
      // Handle errors, e.g., show an error message to the user
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
          value={selectedAuthMethod}
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

      <LoadingButton
        fullWidth
        size="large"
        type="submit"
        variant="contained"
        color="inherit"
        onClick={handleClick}
      >
        Login
      </LoadingButton>
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
