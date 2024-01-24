package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// Config we need
type Config struct {
	Slack          SlackConfig      `yaml:"slack"`
	ServiceNow     ServiceNowConfig `yaml:"servicenow"`
	Global         GlobalConfig     `yaml:"global"`
	Jobs           JobsConfig       `yaml:"jobs"`
	ConfigFilePath string
}

// GlobalConfig Options passed via cmd line
type GlobalConfig struct {
	// loglevel
	LogLevel string `yaml:"logLevel"`

	// write
	Write bool `yaml:"write"`

	RecheckInterval time.Duration
	// if true all task run at start
	RunAtStart bool `yaml:"runAtStart"`
}

// SlackConfig Struct
type SlackConfig struct {
	// Token to authenticate
	BotSecurityToken  string
	UserSecurityToken string
	InfoChannelID     string `yaml:"infoChannelID"`
	Workspace         string `yaml:"workspaceForChatLinks"`
}

// ServiceNowConfig Struct
type ServiceNowConfig struct {
	// Token to authenticate
	PfxCertFile       string
	PfxCertBase64     string
	PfxCertPassword   string
	APIendpoint       string `yaml:"apiEndpoint"`
	APIGetShifts      string
	APIGetGroupMember string
	APIGetWhoIsOnCall string
	APIGetSpans       string

	TLSconfig *tls.Config
}

// JobsConfig Real Work Definition
type JobsConfig struct {
	ScheduleSyncs []ScheduleSync `yaml:"servicenow-schedules-on-duty-to-slack-group"`
}

type ScheduleSync struct {
	CrontabExpressionForRepetition string              `yaml:"crontabExpressionForRepetition"`
	SyncOptions                    ScheduleSyncOptions `yaml:"syncOptions"`
	SyncObjects                    SyncObject          `yaml:"syncObjects"`
}

// SyncObjects Struct
type SyncObject struct {
	SlackGroupHandle string `yaml:"slackGroupHandle"`
	GroupId          string `yaml:"groupId"`
}

// ScheduleSyncOptions SyncOptions Struct
type ScheduleSyncOptions struct {
	DisableSlackHandleTemporaryIfNoneOnShift bool      `yaml:"disableSlackHandleTemporaryIfNoneOnShift"`
	SyncStyle                                SyncStyle `yaml:"syncStyle"`
}

// SyncStyle Type of which Layer (or combination) is used
type SyncStyle string

const (
	OnlyPrimary     = "OnlyPrimary"
	AllActiveLayers = "AllActiveLayers"
)

// NewConfig reads the configuration from the given filePath.
func NewConfig(configFilePath string) (cfg Config, err error) {
	if configFilePath == "" {
		return cfg, fmt.Errorf("path to configuration file not provided")
	}

	cfgBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return cfg, fmt.Errorf("reading configuration file failed: %w", err)
	}
	err = yaml.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("parsing configuration failed: %w", err)
	}
	err = loadEnvVars(&cfg)
	if err != nil {
		return cfg, err
	}

	cfg.ServiceNow.APIGetShifts = "/api/now/on_call_rota/getrotasbygroup/%s"
	cfg.ServiceNow.APIGetGroupMember = "/api/now/on_call_rota/group/members/%s"
	cfg.ServiceNow.APIGetWhoIsOnCall = "/api/now/on_call_rota/whoisoncall?rota_ids=%s&group_ids=%s"
	cfg.ServiceNow.APIGetSpans = "/api/now/on_call_rota/spans?from=%s&group_ids=%s&to=%s&target_tz=UTC"

	return cfg, nil
}

// loadEnvVars fills credentials in the config from env vars
func loadEnvVars(cfg *Config) error {
	cfg.Slack.BotSecurityToken = os.Getenv("SLACK_BOT_TOKEN")
	if cfg.Slack.BotSecurityToken == "" {
		return fmt.Errorf("env variable `SLACK_BOT_TOKEN` is not set")
	}

	cfg.Slack.UserSecurityToken = os.Getenv("SLACK_USER_TOKEN")
	if cfg.Slack.UserSecurityToken == "" {
		return fmt.Errorf("env variable `SLACK_USER_TOKEN` is not set")
	}

	cfg.ServiceNow.PfxCertFile = os.Getenv("SERVICENOW_API_CERT_PKC12")
	/*if cfg.ServiceNow.PfxCertFile == "" {
		return fmt.Errorf("env variable `SERVICENOW_API_CERT_PKC12` is not set")
	}*/

	cfg.ServiceNow.PfxCertBase64 = os.Getenv("SERVICENOW_API_CERT_PKC12_B64")
	if cfg.ServiceNow.PfxCertFile == "" && cfg.ServiceNow.PfxCertBase64 == "" {
		return fmt.Errorf("env variable `SERVICENOW_API_CERT_PKC12` or `SERVICENOW_API_CERT_PKC12_B64` has to be set")
	}

	cfg.ServiceNow.PfxCertPassword = os.Getenv("SERVICENOW_API_CERT_PKC12_PWD")
	if cfg.ServiceNow.PfxCertPassword == "" {
		return fmt.Errorf("env variable `SERVICENOW_API_CERT_PKC12_PWD` is not set")
	}

	return nil
}
