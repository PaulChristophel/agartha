import PropTypes from 'prop-types'; // Make sure to import PropTypes
import { Navigate } from 'react-router-dom';
import { Helmet } from 'react-helmet-async';

import { useSession } from 'src/api/session.ts';

import { UserView } from 'src/sections/user/view';

// ----------------------------------------------------------------------

// Authentication wrapper component
const AuthWrapper = ({ children }) => {
  const { authToken: token } = useSession();

  // Redirect if no token
  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return children;
};

AuthWrapper.propTypes = {
  children: PropTypes.node, // Define the expected type for children
};

export default function UserPage() {
  return (
    <>
      <Helmet>
        <title> User | Agartha </title>
      </Helmet>

      <AuthWrapper>
        <UserView />
      </AuthWrapper>
    </>
  );
}
