interface Config {
  SaltAPIEndpoint: string;
  Version: string;
  ForgotPasswordURL: string;
  GetStartedURL: string;
  CASServiceURL: string;
}

declare global {
  interface Window {
    config?: Config;
  }
}

// Define default values
const defaultConfig: Config = {
  SaltAPIEndpoint: '',
  Version: 'devel',
  ForgotPasswordURL: '',
  GetStartedURL: '',
  CASServiceURL: '',
};

// Placeholder values that indicate the config has not been set properly
const placeholderConfig: Config = {
  SaltAPIEndpoint: '{{.SaltAPIEndpoint}}',
  Version: '{{.Version}}',
  ForgotPasswordURL: '{{.ForgotPasswordURL}}',
  GetStartedURL: '{{.GetStartedURL}}',
  CASServiceURL: '{{.CASServiceURL}}',
};

// Function to determine if the current config is a placeholder
function isPlaceholderConfig(config: Config): boolean {
  return (
    config.SaltAPIEndpoint === placeholderConfig.SaltAPIEndpoint &&
    config.Version === placeholderConfig.Version &&
    config.ForgotPasswordURL === placeholderConfig.ForgotPasswordURL &&
    config.GetStartedURL === placeholderConfig.GetStartedURL &&
    config.CASServiceURL === placeholderConfig.CASServiceURL
  );
}

// Use the default values if window.config is not defined or if it contains placeholder values
const { SaltAPIEndpoint, Version, ForgotPasswordURL, GetStartedURL, CASServiceURL }: Config =
  window.config && !isPlaceholderConfig(window.config) ? window.config : defaultConfig;

export { Version, GetStartedURL, CASServiceURL, SaltAPIEndpoint, ForgotPasswordURL };
