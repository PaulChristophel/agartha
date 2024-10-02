import React from 'react';
import PropTypes from 'prop-types';

import Typography from '@mui/material/Typography';

interface ErrorType {
  message: string;
}

interface DataType {
  version: string;
}

interface VersionDataProps {
  isLoading: boolean;
  error: ErrorType | null;
  data: DataType | null;
}

const VersionData: React.FC<VersionDataProps> = ({ isLoading, error, data }) => {
  if (isLoading) {
    return <Typography variant="body2">Loading...</Typography>;
  }

  if (error) {
    return (
      <Typography variant="body2" color="error">
        Error: {error.message}
      </Typography>
    );
  }

  return <Typography variant="body2">Version: {data?.version}</Typography>;
};

VersionData.propTypes = {
  isLoading: PropTypes.bool.isRequired,
  error: PropTypes.shape({
    message: PropTypes.string.isRequired,
  }),
  data: PropTypes.shape({
    version: PropTypes.string.isRequired,
  }),
};

export default VersionData;
