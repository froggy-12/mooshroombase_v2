package configs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func InitConfigs() Config {
	var configs Config
	if _, err := os.Stat("configs.json"); os.IsNotExist(err) {
		CreateDefaultConfig(&configs)
	} else {
		data, err := os.ReadFile("configs.json")
		if err != nil {
			log.Fatal("Error: ", err.Error())
		}
		err = json.Unmarshal(data, &configs)
		if err != nil {
			log.Fatal("Error: ", err.Error())
		}
	}
	return configs
}

func CreateDefaultConfig(configs *Config) {
	*configs = Config{
		Applications: Applications{
			BackEndURlWithDomain: "http://localhost:6644",
			BackEndPort:          ":6644",
			AllowedCorsOrigins:   []string{"*"},
		},
		Authentication: Authentication{
			Auth:                         true,
			OAuth:                        false,
			GoogleOAuth:                  false,
			GoogleOAuthAppID:             "",
			GoogleOAuthAppSecret:         "",
			GithubOAuth:                  false,
			GithubOAuthAppID:             "",
			GithubOAuthAppSecret:         "",
			EmailVerificationAllowed:     false,
			SetJWTTokenAfterSignUp:       false,
			RealTimeUserData:             false,
			SendEmailAfterSignUpWithCode: true,
		},
		DatabaseConfigurations: DatabaseConfigurations{
			PrimaryDB:           "mongodb",
			RunningDatabases:    []string{"mongodb", "redis", "mariadb"},
			MongoDBRootPassword: "mooshroombase",
			MongoDBServerPort:   "27018",
			MariaDBRootPassword: "mooshroombase",
			MariaDBServerPort:   "6645",
			RedisDBRootPassword: "mooshroombase",
			RedisDBServerPort:   "6656",
		},
		HttpConfigurations: HttpConfigurations{
			JWTSecret:              "SuperSecretMooshroombase",
			CorsHeaderMaxAge:       7,
			JWTTokenExpirationTime: 7,
		},
		SMTPConfigurations: SMTPConfigurations{
			SMTPEnabled:            false,
			SMTPServerAddress:      "smtp.gmail.com",
			SMTPServerPORT:         "587",
			SMTPEmailAddrss:        "",
			SMTPEmailPassword:      "",
			SMTPAllowedForEveryone: false,
		},
		ExtraConfigurations: ExtraConfigurations{
			BodySizeLimit:                     100 * 1024 * 1024,
			DefaultProfilePictureUrl:          "",
			DebugLogging:                      true,
			AfterStartDockerThreadRestartTime: 50,
		},
		Features: Features{
			FileUplaod:    true,
			ServeFile:     true,
			ChatFunctions: true,
		},
	}
	data, err := json.MarshalIndent(*configs, "", "  ")
	if err != nil {
		fmt.Println("Error creating default config:", err)
		os.Exit(1)
	}
	err = os.WriteFile("configs.json", data, 0644)
	if err != nil {
		fmt.Println("Error writing default config:", err)
		os.Exit(1)
	}
	fmt.Println("[Important]: ", "Default config file created. Please configure the settings in configs.json to avoid using default container names")
	fmt.Println("[Important]: ", "Using default container names can lead to conflicts and security issues.")
	fmt.Println("[Important]: ", "Please restart the application after configuring the settings.")
	fmt.Println("[Important]: ", "Ohh yeah remember if the containers has been created u need to delete them manually")
	os.Exit(0)
}
