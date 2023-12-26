package main

import (
	"crypto/tls"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	servicenow "github.com/sapcc/schedules2slack/internal/clients/servicenow"
	slackclient "github.com/sapcc/schedules2slack/internal/clients/slack"
	config "github.com/sapcc/schedules2slack/internal/config"
	jobs "github.com/sapcc/schedules2slack/internal/jobs"
	"golang.org/x/crypto/pkcs12"
)

var opts config.Config

func printUsage() {
	var CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	_, _ = fmt.Fprintf(CommandLine.Output(), "\n\nUsage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.StringVar(&opts.ConfigFilePath, "config", "./config.yml", "Config file path including file name.")
	flag.BoolVar(&opts.Global.Write, "write", false, "[true|false] write changes? Overrides config setting!")
	flag.Parse()

	cfg, err := config.NewConfig(opts.ConfigFilePath)
	if err != nil {
		printUsage()
		log.Fatal(err)
	}

	initLogging(cfg.Global.LogLevel)

	convert(&cfg.ServiceNow)

	slackClient, err := slackclient.NewClient(&cfg.Slack)
	if err != nil {
		log.Fatal(err)
	}

	servicenowClient, err := servicenow.NewClient(&cfg.ServiceNow)
	if err != nil {
		log.Fatal(err)
	}

	c := cron.New(cron.WithLocation(time.UTC))
	_, err = c.AddFunc("0 * * * *", func() {
		if err := slackClient.LoadMasterData(); err != nil {
			log.Warnf("loading slack masterdata failed: %s", err.Error())
		}
	})
	if err != nil {
		log.Fatalf("adding Slack masterdata loading to cron failed: %s", err.Error())
	}

	//member sync jobs
	for _, s := range cfg.Jobs.ScheduleSyncs {
		job, err := jobs.NewScheduleSyncJob(s, !cfg.Global.Write, servicenowClient, slackClient)
		if err != nil {
			log.Fatalf("creating job to sync '%s' failed: %s", s.SyncObjects.SlackGroupHandle, err.Error())
		}

		_, err = c.AddFunc(s.CrontabExpressionForRepetition, func() {
			err := job.Run()
			if err != nil {
				log.Warnf("schedule_sync failed: %s", err.Error())
			}

			if err = jobs.PostInfoMessage(slackClient, job); err != nil {
				log.Warnf("posting update to slack failed: %s", err.Error())
			}
		})
		if err != nil {
			log.Fatalf("failed to create job: %s", err.Error())
		}
	}
	go c.Start()
	defer c.Stop()

	time.Sleep(2000)
	if cfg.Global.RunAtStart {
		for rc, e := range c.Entries() {
			if rc == 0 {
				continue
			}
			log.Debugf("job %d: next run %s; valid: %v", e.ID, e.Next, e.Valid())
			if e.Valid() {
				c.Entry(e.ID).WrappedJob.Run()
			}
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		s := <-sig
		log.Infof("received %v, shutting down", s.String())
	} else {
		log.Info("cfg.Global.RunAtStart is set to: ", cfg.Global.RunAtStart)
	}
}

// initLogging configures the logger
func initLogging(logLevel string) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.SetReportCaller(false)
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Info("parsing log level failed, defaulting to info")
		level = log.InfoLevel
	}
	log.SetLevel(level)
}

func convert(cfg *config.ServiceNowConfig) {

	// read PKCS#12 file
	cert, err := os.ReadFile(cfg.PfxCertFile)
	if err != nil {
		fmt.Println("Cert Error:", err)
		return
	}

	certificates, err := pkcs12.ToPEM(cert, cfg.PfxCertPassword)
	//print(b, err)

	/*// Erstelle einen TLS-Konfigurationscontainer
	certificates, err := tls.X509KeyPair(cert, []byte(cfg.PfxCertPassword))
	if err != nil {
		fmt.Println("Creating Cert Pair failed:", err)
		return
	}
	*/
	// Füge das Zertifikat zum Wurzel-Zertifikats-Pool hinzu
	/*roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(certificates))
	if !ok {
		fmt.Println("Error on adding RootCertPool")
		return
	}*/

	var pemData []byte
	for _, b := range certificates {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}
	// then use PEM data for tls to construct tls certificate:
	pemcert, err := tls.X509KeyPair(pemData, pemData)
	if err != nil {
		panic(err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{pemcert},
	}
	/*tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      roots,
	}*/

	cfg.TLSconfig = tlsConfig

}
