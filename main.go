package main

import (
	"github.com/nireitdev/go-ssh-block-ip-bruteforce/logparser"
	"log"
)

func main() {

	//SSHd block
	ssh_block := logparser.Logfile{Filename: "/var/log/auth.log",
		Searchreg: "Failed",
		Filterreg: "\\d+\\.\\d+.\\d+\\.\\d+",
	}

	//Postfix block:
	smtpd_block := logparser.Logfile{Filename: "/var/log/mail.info",
		Searchreg: "SASL LOGIN authentication failed",
		Filterreg: "\\d+\\.\\d+.\\d+\\.\\d+",
	}

	banSSH := ssh_block.Run()
	banSTMPD := smtpd_block.Run()

	for {
		select {
		case ip := <-banSSH:
			log.Println("Ban SSH Ip: ", ip)

		case ip := <-banSTMPD:
			log.Println("Ban STMPD Ip: ", ip)
		}

	}

}
