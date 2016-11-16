package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/mxk/go-imap/imap"
)

type smtpCfg struct {
	server       *string
	serverPort   *int
	helloName    *string
	username     *string
	password     *string
	cram         *bool
	sTARTLS      *bool
	mailFrom     *string
	rcpt         *string
	subject      *string
	SSLonConnect *bool
}

type imapCfg struct {
	server     *string
	serverPort *int
	username   *string
	password   *string
	imapTLS    *bool
	debug      *bool
	silent     *bool
}

func parseCmd(smtpcfg *smtpCfg, imapcfg *imapCfg) {

	smtpcfg.server = flag.String("smtpServer", "", "SMTP server string. No Default")
	smtpcfg.serverPort = flag.Int("smtpServerPort", 25, "Port of the SMTP server. Default is 25")
	currentHost, _ := os.Hostname()
	smtpcfg.helloName = flag.String("smtpHelloName", currentHost, "Hostname for the hello string. Default is current hostname")
	smtpcfg.username = flag.String("smtpUsername", "", "User to authenticate to SMTP server. If omited, then no authentication will be used")
	smtpcfg.password = flag.String("smtpPassword", "", "Password to authenticate to SMTP server. You must specify it, if smtpUsername is specified.")
	smtpcfg.cram = flag.Bool("smtpCram", false, "Use this flag to use CRAMMD5 authentication. Default is no CRAMMD5")
	smtpcfg.sTARTLS = flag.Bool("smtpSTARTTLS", false, "Use this flag to enable STARTTLS for smtp. By default it is off")
	smtpcfg.SSLonConnect = flag.Bool("SSLonConnect", false, "Use SSL during SMTP connection. By default is off")
	smtpcfg.mailFrom = flag.String("smtpMailFrom", "", "Address of the message sender")
	smtpcfg.rcpt = flag.String("smtpRcpt", "", "Mail recipient address")
	smtpcfg.subject = flag.String("smtpSubject", "Delivery quality monitoring", "Subject of message to send. Default is 'Delivery quality monitoring'")

	imapcfg.server = flag.String("imapServer", "", "IMAP server. No default value")
	imapcfg.serverPort = flag.Int("imapServerPort", 143, "Port of the IMAP server. Default is 143")
	imapcfg.username = flag.String("imapUsername", "", "User to authenticate to IMAP server. No default value")
	imapcfg.password = flag.String("imapPassword", "", "Password to authenticate to IMAP server. No default value")
	imapcfg.imapTLS = flag.Bool("imapTLS", false, "Use this flag to enable TLS connection. By default this setting is off")
	imapcfg.debug = flag.Bool("debug", false, "Show additional information. By default is off.")
	imapcfg.silent = flag.Bool("silent", false, "If enabled, then only delivery time is displayed, without additional words. By default is off.")

	flag.Parse()

	if *smtpcfg.username != "" && *smtpcfg.password == "" {
		fmt.Fprintf(os.Stderr, "You must specify smtpPassword")
		os.Exit(1)
	}

	if *smtpcfg.server == "" {
		fmt.Fprintf(os.Stderr, "You must specify smtp server name")
		os.Exit(1)
	}

	if *smtpcfg.mailFrom == "" {
		fmt.Fprintf(os.Stderr, "You must specify message sender address!")
		os.Exit(1)
	}

	if *smtpcfg.rcpt == "" {
		fmt.Fprintf(os.Stderr, "You must specify recipient address!")
		os.Exit(1)
	}

	if *imapcfg.server == "" {
		fmt.Fprintf(os.Stderr, "You must specify IMAP server address!")
		os.Exit(1)
	}

	if *imapcfg.username == "" {
		fmt.Fprintf(os.Stderr, "You must specify IMAP username")
		os.Exit(1)
	}

	if *imapcfg.password == "" {
		fmt.Fprintf(os.Stderr, "You must specify IMAP password")
		os.Exit(1)
	}

}

func main() {

	var smtpconf smtpCfg
	var imapconf imapCfg

	parseCmd(&smtpconf, &imapconf)

	var imapClient *imap.Client
	var imapError error

	if *imapconf.imapTLS {
		tlscfg := &tls.Config{InsecureSkipVerify: true}
		imapClient, imapError = imap.DialTLS(*imapconf.server+":"+strconv.Itoa(*imapconf.serverPort), tlscfg)
		if imapError != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to IMAP server (%s): %s\n", *imapconf.server, imapError)
			os.Exit(1)
		}
	} else {
		imapClient, imapError = imap.Dial(*imapconf.server + ":" + strconv.Itoa(*imapconf.serverPort))
		if imapError != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to IMAP server (%s): %s\n", *imapconf.server, imapError)
			os.Exit(1)
		}
	}

	_, imapError = imapClient.Login(*imapconf.username, *imapconf.password)
	if imapError != nil {
		fmt.Fprintf(os.Stderr, "Error logging in to IMAP server (%s): %s\n", *imapconf.server, imapError)
		os.Exit(1)
	}

	set, _ := imap.NewSeqSet("")
	var uidslice []uint32

	start := time.Now()

	var smtpClient *smtp.Client
	var smtpError error

	if *smtpconf.SSLonConnect {
		tlscfg := &tls.Config{InsecureSkipVerify: true}
		tlsConnection, tlsError := tls.Dial("tcp", *smtpconf.server+":"+strconv.Itoa(*smtpconf.serverPort), tlscfg)
		if tlsError != nil {
			fmt.Fprintf(os.Stderr, "Error establishing SSL connection to SMTP server (%s): %s\n", *smtpconf.server, tlsError)
			os.Exit(1)
		}
		smtpClient, smtpError = smtp.NewClient(tlsConnection, *smtpconf.server)
		if smtpError != nil {
			fmt.Fprintf(os.Stderr, "Error establishing connection to SMTP server (%s): %s\n", *smtpconf.server, smtpError)
			os.Exit(1)
		}

	} else {
		smtpClient, smtpError = smtp.Dial(*smtpconf.server + ":" + strconv.Itoa(*smtpconf.serverPort))
		if smtpError != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to SMTP server (%s): %s\n", *smtpconf.server, smtpError)
			os.Exit(1)
		}
	}

	smtpError = smtpClient.Hello(*smtpconf.helloName)
	if smtpError != nil {
		fmt.Fprintf(os.Stderr, "Error sending hello command (%s): %s\n", *smtpconf.server, smtpError)
		os.Exit(1)
	}

	if *smtpconf.sTARTLS {
		tlscfg := &tls.Config{InsecureSkipVerify: true}
		smtpError = smtpClient.StartTLS(tlscfg)
		if smtpError != nil {
			fmt.Fprintf(os.Stderr, "Error starting STARTTLS session (%s): %s\n", *smtpconf.server, smtpError)
			os.Exit(1)
		}
	}

	var a smtp.Auth

	if *smtpconf.username != "" {
		if *smtpconf.cram {
			a = smtp.CRAMMD5Auth(*smtpconf.username, *smtpconf.password)
		} else {
			a = smtp.PlainAuth("", *smtpconf.username, *smtpconf.password, *smtpconf.server)
		}

		smtpError = smtpClient.Auth(a)
		if smtpError != nil {
			fmt.Fprintf(os.Stderr, "Error authenticating to SMTP server (%s): %s\n", *smtpconf.server, smtpError)
			os.Exit(1)
		}
	}

	smtpError = smtpClient.Mail(*smtpconf.mailFrom)

	if smtpError != nil {
		fmt.Fprintf(os.Stderr, "Error sending MAIL FROM command (%s): %s\n", *smtpconf.server, smtpError)
		os.Exit(1)
	}

	smtpError = smtpClient.Rcpt(*smtpconf.rcpt)

	if smtpError != nil {
		fmt.Fprintf(os.Stderr, "Error sending Rcpt to command (%s): %s\n", *smtpconf.server, smtpError)
		os.Exit(1)
	}

	smtpWriter, smtpError := smtpClient.Data()

	if smtpError != nil {
		fmt.Fprintf(os.Stderr, "Error opening DATA writer (%s): %s\n", *smtpconf.server, smtpError)
		os.Exit(1)
	}

	smtpMsg := `From: ` + *smtpconf.mailFrom + `
To: ` + *smtpconf.rcpt + `
Subject: ` + start.Format("20060102150405") + *smtpconf.subject + `

Test message to monitor mail delivery
.

`

	if _, smtpError := smtpWriter.Write([]byte(smtpMsg)); smtpError != nil {
		fmt.Fprintf(os.Stderr, "Error sending data to server (%s): %s\n", *smtpconf.server, smtpError)
		os.Exit(1)
	}

	smtpError = smtpWriter.Close()
	if smtpError != nil {
		fmt.Fprintf(os.Stderr, "Error sending data to SMTP server (%s): %s\n", *smtpconf.server, smtpError)
		os.Exit(1)
	}

	smtpClient.Quit()

	if *imapconf.debug {
		elapsed := (time.Since(start)).Seconds()
		fmt.Printf("Time to send message from SMTP server (%s): %d\n", *smtpconf.server, int(elapsed))
	}

	wait := 0

	for ; (len(uidslice) == 0) && wait < 200; wait++ {
		_, imapError = imapClient.Select("INBOX", false)

		if imapError != nil {
			fmt.Fprintf(os.Stderr, "Error selecting INBOX folder (%s): %s\n", *imapconf.server, imapError)
			os.Exit(1)
		}

		cmd, imapError := imap.Wait(imapClient.Send("UID SEARCH", `SUBJECT "`+start.Format("20060102150405")+*smtpconf.subject+`"`))
		if imapError != nil {
			fmt.Fprintf(os.Stderr, "Error sending search command to IMAP session (%s): %s\n", *imapconf.server, imapError)
			os.Exit(1)
		}

		for _, rsp := range cmd.Data {

			uidslice = rsp.SearchResults()
			if *imapconf.debug {
				fmt.Println("Found messages: ", uidslice)
			}
			for i := range uidslice {
				set.AddNum(uidslice[i])
			}
		}
		time.Sleep(time.Second)
	}

	if !(*imapconf.silent) {
		fmt.Printf("Time to send and recieve message with IMAP (%s): %d\n", *imapconf.server, int((time.Since(start)).Seconds()))
	} else {
		fmt.Println(int((time.Since(start)).Seconds()))
	}

	_, imapError = imap.Wait(imapClient.UIDStore(set, `+FLAGS`, `(\Deleted)`))

	if imapError != nil {
		fmt.Fprintf(os.Stderr, "Error setting Deleted flag for messages (%s): %s\n ", *imapconf.server, imapError)
	}
	_, imapError = imap.Wait(imapClient.Expunge(nil))

	if imapError != nil {
		fmt.Fprintf(os.Stderr, "Error deleting messages (%s): %s\n", *imapconf.server, imapError)
	}

	imapClient.Logout(1)

}
