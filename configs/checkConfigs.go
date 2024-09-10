package configs

import "log"

func CheckIfFieldsAreEmpty(c Config) {
	if c.Applications.BackEndURlWithDomain == "" {
		log.Fatal("BackEndURlWithDomain is empty")
	}
	if len(c.Applications.AllowedCorsOrigins) == 0 {
		log.Fatal("AllowedCorsOrigins is empty")
	}
	if c.Authentication.GoogleOAuth && c.Authentication.GoogleOAuthAppID == "" {
		log.Fatal("GoogleOAuthAppID is empty")
	}
	if c.Authentication.GoogleOAuth && c.Authentication.GoogleOAuthAppSecret == "" {
		log.Fatal("GoogleOAuthAppSecret is empty")
	}
	if c.Authentication.GithubOAuth && c.Authentication.GithubOAuthAppID == "" {
		log.Fatal("GithubOAuthAppID is empty")
	}
	if c.Authentication.GithubOAuth && c.Authentication.GithubOAuthAppSecret == "" {
		log.Fatal("GithubOAuthAppSecret is empty")
	}
	if c.SMTPConfigurations.SMTPEnabled {
		if c.SMTPConfigurations.SMTPServerAddress == "" {
			log.Fatal("SMTPServerAddress is empty")
		}
		if c.SMTPConfigurations.SMTPServerPORT == "" {
			log.Fatal("SMTPServerPORT is empty")
		}
		if c.SMTPConfigurations.SMTPEmailAddrss == "" {
			log.Fatal("SMTPEmailAddrss is empty")
		}
		if c.SMTPConfigurations.SMTPEmailPassword == "" {
			log.Fatal("SMTPEmailPassword is empty")
		}
	}
	if c.DatabaseConfigurations.PrimaryDB == "" {
		log.Fatal("PrimaryDB is empty")
	}
	if len(c.DatabaseConfigurations.RunningDatabases) == 0 {
		log.Fatal("RunningDatabases is empty")
	}
	if c.DatabaseConfigurations.MongoDBRootPassword == "" && contains(c.DatabaseConfigurations.RunningDatabases, "mongodb") {
		log.Fatal("MongoDBRootPassword is empty")
	}
	if c.DatabaseConfigurations.MariaDBRootPassword == "" && contains(c.DatabaseConfigurations.RunningDatabases, "mariadb") {
		log.Fatal("MariaDBRootPassword is empty")
	}
	if c.DatabaseConfigurations.RedisDBRootPassword == "" && contains(c.DatabaseConfigurations.RunningDatabases, "redis") {
		log.Fatal("RedisDBRootPassword is empty")
	}
	if c.Features.ChatFunctions && !contains(c.DatabaseConfigurations.RunningDatabases, "redis") {
		log.Fatal("ChatFunctions is enabled but Redis is not present in RunningDatabases")
	}
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
