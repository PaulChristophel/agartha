package config

type LDAPOptions struct {
	Server            string `mapstructure:"server" yaml:"server"`
	User              string `mapstructure:"user" yaml:"user"`
	Password          string `mapstructure:"password" yaml:"password"`
	BaseDN            string `mapstructure:"base_dn" yaml:"base_dn"`
	Filter            string `mapstructure:"filter" yaml:"filter"`
	LDAPDomainDefault string `mapstructure:"ldap_domain_default" yaml:"ldap_domain_default"`
}
