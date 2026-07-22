import { useEffect } from 'react';
import PropTypes from 'prop-types';
import { Navigate } from 'react-router-dom';

import useFetchAndStoreSaltAuth from 'src/hooks/auth/useFetchAndStoreSaltAuth.ts';

import { useSession, isJwtExpired, isSaltAuthExpired } from 'src/api/session.ts';

interface AuthWrapperProps {
  children: React.ReactNode;
}

// Authentication wrapper component
const AuthWrapper: React.FC<AuthWrapperProps> = ({ children }) => {
  const { authToken, authSalt } = useSession();
  const { postSaltAuth } = useFetchAndStoreSaltAuth();

  useEffect(() => {
    if (authToken && authSalt) {
      // Only attempt to refresh the salt token if the authToken is not expired
      if (!isJwtExpired(authToken) && isSaltAuthExpired(authSalt)) {
        void postSaltAuth().catch((error) => {
          console.error('Failed to refresh Salt authentication:', error);
        });
      }
    }
  }, [authSalt, postSaltAuth, authToken]);

  // Redirect if no token or token is expired
  if (!authToken || isJwtExpired(authToken)) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

AuthWrapper.propTypes = {
  children: PropTypes.element.isRequired,
};

export default AuthWrapper;
