import { useEffect } from 'react';
import PropTypes from 'prop-types';
import { Navigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';

import { getSession } from 'src/api/auth.ts';
import { queryKeys } from 'src/api/queryKeys.ts';
import { useSession, sessionStore } from 'src/api/session.ts';

interface AuthWrapperProps {
  children: React.ReactNode;
}

// Authentication wrapper component
const AuthWrapper: React.FC<AuthWrapperProps> = ({ children }) => {
  const { status } = useSession();
  const { data, isPending, isError } = useQuery({
    queryKey: queryKeys.auth.session(),
    queryFn: ({ signal }) => getSession(signal),
  });

  useEffect(() => {
    if (data) sessionStore.setAuthenticated(data);
  }, [data]);
  if (isPending) return null;
  if (isError || status === 'anonymous') {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

AuthWrapper.propTypes = {
  children: PropTypes.element.isRequired,
};

export default AuthWrapper;
