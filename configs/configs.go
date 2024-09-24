package configs

type Applications struct {
	BackEndURlWithDomain string   `json:"back_end_url_with_domain"` // by default it will be http://localhost:6644
	BackEndPort          string   `json:"back_end_port"`            // by default it will be ":6644"
	AllowedCorsOrigins   []string `json:"allowed_cors_origins"`     // by default it set *
}

type Authentication struct {
	Auth                         bool   `json:"auth"`                        // by default true
	OAuth                        bool   `json:"oauth"`                       // by default false
	GoogleOAuth                  bool   `json:"google_oauth"`                // by default false (it will require OAuth to be true)
	GoogleOAuthAppID             string `json:"google_oauth_app_id"`         // required if Google OAuth enabled
	GoogleOAuthAppSecret         string `json:"google_oauth_app_secret"`     // required if Google OAuth enabled
	GithubOAuth                  bool   `json:"github_oauth"`                // by default false (it will require OAuth to be true)
	GithubOAuthAppID             string `json:"github_oauth_app_id"`         // required if Github OAuth enabled
	GithubOAuthAppSecret         string `json:"github_oauth_app_secret"`     // required if Github OAuth enabled
	EmailVerificationAllowed     bool   `json:"email_verification_allowed"`  // adds some latency to the server and by default false and its preference dont need to turn on you should learn more about this first
	SetJWTTokenAfterSignUp       bool   `json:"set_jwt_token_after_sign_up"` // its false by default
	RealTimeUserData             bool   `json:"real_time_user_data"`         // by default false turn true for real time user data works with only mongodb not mariadb
	SendEmailAfterSignUpWithCode bool   `json:"send_email_after_sign_up_with_code"`
}

type DatabaseConfigurations struct {
	PrimaryDB           string   `json:"primary_db"`             // either mongodb or mariadb (authentication will be handled by primary db)
	RunningDatabases    []string `json:"running_databases"`      // mongodb, mariadb, redis otherwise wont work
	MongoDBRootPassword string   `json:"mongodb_root_password"`  // by default it will be mooshroombase for root user u can change it ofc
	MongoDBServerPort   string   `json:"mongodb_server_port"`    // by default it will be 66441
	MariaDBRootPassword string   `json:"mariadb_root_password"`  // by default it will be mooshroombase for root
	MariaDBServerPort   string   `json:"mariadb_server_port"`    // by default it will be 6645
	RedisDBRootPassword string   `json:"redis_db_root_password"` // by default it will be mooshroombase
	RedisDBServerPort   string   `json:"redis_db_server_port"`   // by default it will be 6656
	// note if u dont pay attention and do everything default it gonna make problems for future so please configure everything at once that is best.
}

type HttpConfigurations struct {
	JWTSecret              string `json:"jwt_secret"`                // by default it will be SuperSecretMooshroombase
	CorsHeaderMaxAge       int    `json:"cors_header_max_age"`       // by default 7 (1 = 1 day)
	JWTTokenExpirationTime int    `json:"jwt_token_expiration_time"` // by default 7 (1 = 1 day)
}

type SMTPConfigurations struct {
	SMTPEnabled            bool   `json:"smtp_enabled"`              // by default false
	SMTPServerAddress      string `json:"smtp_server_address"`       // by default smtp.gmail.com
	SMTPServerPORT         string `json:"smtp_server_port"`          // by default 587
	SMTPEmailAddrss        string `json:"smtp_email_address"`        // required if SMTPEnabled == true
	SMTPEmailPassword      string `json:"smtp_email_password"`       // required if SMTPEnabled == true
	SMTPAllowedForEveryone bool   `json:"smtp_allowed_for_everyone"` // by default false
}

type ExtraConfigurations struct {
	BodySizeLimit                     int    `json:"body_size_limit"`             // the amount of data can be transfered by the api including file uploads by default 100 mb 100 * 1024 * 1024 = 104857600
	DefaultProfilePictureUrl          string `json:"default_profile_picture_url"` // by default empty
	DebugLogging                      bool   `json:"debug_logging"`
	AfterStartDockerThreadRestartTime int    `json:"after_start_docker_thread_restart_time"` // 1 = second
	RealTimeMainSwitch                bool   `json:"real_time_main_switch"`                  // turn this true for every real time use case
}

type Features struct {
	FileUplaod    bool `json:"file_upload"`    // by default true
	ServeFile     bool `json:"serve_file"`     // by default true
	ChatFunctions bool `json:"chat_functions"` // by default true its its enabled and there is no redis in the running database slice it will throw error
}

type Config struct {
	Applications           Applications           `json:"applications"`
	Authentication         Authentication         `json:"authentication"`
	DatabaseConfigurations DatabaseConfigurations `json:"database_configurations"`
	HttpConfigurations     HttpConfigurations     `json:"http_configurations"`
	SMTPConfigurations     SMTPConfigurations     `json:"smtp_configurations"`
	ExtraConfigurations    ExtraConfigurations    `json:"extra_configurations"`
	Features               Features               `json:"features"`
}

var Configs Config
