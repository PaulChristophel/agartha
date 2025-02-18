import { useEffect } from 'react';
import PropTypes from 'prop-types';
import { jwtDecode } from 'jwt-decode';
import { Navigate } from 'react-router-dom';

import useFetchAndStoreSaltAuth from 'src/hooks/auth/useFetchAndStoreSaltAuth.ts';

interface AuthWrapperProps {
  children: React.ReactNode;
}

interface AuthSalt {
  token: string;
  eauth: string;
  start: number;
  expire: number;
  perms: string[];
}

interface JWTToken {
  exp: number;
}

// Authentication wrapper component
const AuthWrapper: React.FC<AuthWrapperProps> = ({ children }) => {
  const authToken = localStorage.getItem('authToken');
  const authSaltString = localStorage.getItem('authSalt');
  const { postSaltAuth } = useFetchAndStoreSaltAuth();

  const isAuthSaltExpired = (authSaltVar: AuthSalt): boolean =>
    authSaltVar.expire * 1000 < Date.now();

  const isTokenExpired = (token: string): boolean => {
    try {
      const decodedToken: JWTToken = jwtDecode(token);
      return decodedToken.exp * 1000 < Date.now();
    } catch (error) {
      return true; // If token cannot be decoded, consider it expired
    }
  };

  useEffect(() => {
    if (authToken && authSaltString) {
      const parsedAuthSalt: AuthSalt = JSON.parse(authSaltString);

      // Only attempt to refresh the salt token if the authToken is not expired
      if (!isTokenExpired(authToken) && isAuthSaltExpired(parsedAuthSalt)) {
        postSaltAuth(authToken);
      }
    }
  }, [authSaltString, postSaltAuth, authToken]);

  // Redirect if no token or token is expired
  if (!authToken || isTokenExpired(authToken)) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

AuthWrapper.propTypes = {
  children: PropTypes.element.isRequired,
};

export default AuthWrapper;
