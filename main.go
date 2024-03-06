package main

import (
	"context"
	"fmt"
	"github.com/nireitdev/go-ssh-block-ip-bruteforce/config"
	"github.com/nireitdev/go-ssh-block-ip-bruteforce/db"
	"github.com/nireitdev/go-ssh-block-ip-bruteforce/logparser"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type remoteIP struct {
	//limiter  *rate.Limiter
	attemps  int
	lastSeen time.Time
}

var remoteips = make(map[string]*remoteIP)
var mu sync.Mutex
var cfg *config.Config
var redisclient db.Redis
var ctx context.Context

func main() {

	cfg = config.ReadConfig()

	hostname, err := os.Hostname()
	if err != nil {
		log.Println(err)
	}

	redisclient = db.Redis{Addr: cfg.Redis.Addr,
		User:     cfg.Redis.User,
		Password: cfg.Redis.Pass,
	}
	ctx = context.Background()
	redisclient.Open(ctx)

	//envio data al canal
	err = redisclient.Publish("INIT SERVER NRO " + strconv.Itoa(int(redisclient.NroServer)) + " host = " + hostname)
	if err != nil {
		log.Fatalf("Error tratando de publicar", err)
	}

	//SSHd block
	ssh_block := logparser.Logfile{Filename: cfg.Logfile.Filename,
		Searchreg: cfg.Logfile.Searchreg,
		Filterreg: cfg.Logfile.Filterreg,
	}
	banSSH := ssh_block.Run()

	//Postfix block:
	//smtpd_block := logparser.Logfile{Filename: "mail.log",
	//	Searchreg: "SASL LOGIN authentication failed",
	//	Filterreg: "\\d+\\.\\d+.\\d+\\.\\d+",
	//}
	//banSTMPD := smtpd_block.Run()

	//inicio la limpieza de ips
	go cleanupRemoteIP()

	//inicio ban de IP de los otros servidores
	go banInformedIP()

	for {
		select {
		case ip := <-banSSH:
			log.Println("Invalid Auth Ip: ", ip)

			if !allowIP(ip) {
				//Block IP!

				log.Println("Blocking ip:", ip)
				err = redisclient.Publish(hostname + " " + ip)
				if err != nil {
					log.Fatalf("Error tratando de publicar", err)
				}

				runcmd(ip)

			}
		}

	}

}

// Chequea la cantidad de intentos de la IP,
// incrementa el valor de vistas y setea "lastseen"
func allowIP(ip string) bool {
	mu.Lock()
	defer mu.Unlock()

	v, exists := remoteips[ip]
	if !exists {
		remoteips[ip] = &remoteIP{1, time.Now()}
		return true
	}

	v.lastSeen = time.Now()
	v.attemps++
	log.Printf("Ip: %s  Attempts: %d", ip, v.attemps)
	if v.attemps > cfg.Application.MaxAttempts {
		return false
	}
	return true
}

// Funcion encargada de borrar viejas IPs que ya no se
// vieron durante un intervalo de tiempo MAX_INTERVAL_SCAN
func cleanupRemoteIP() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, v := range remoteips {
			if time.Since(v.lastSeen) > time.Duration(cfg.Application.MaxIntervalScan)*time.Minute {
				log.Printf("Down ip: %s", ip)
				delete(remoteips, ip)
			}
		}
		mu.Unlock()
	}
}

func banInformedIP() {
	//recibo data al canal
	sub := redisclient.Subscribe()
	for msg := range sub {
		var host, ip string
		_, err := fmt.Sscanf(msg, "%s %s", &host, &ip)
		if err != nil {
			log.Println("Error parseando info: ", err)
		}
		if host == "INIT" {
			//INIT= servidores iniciando la applicaion
			log.Printf(msg)
		} else {
			log.Printf("Banned external Host: %s   Banned IP: %s \n", host, ip)
			runcmd(ip)
		}
	}
}

func runcmd(bannedIP string) {
	var c *exec.Cmd
	cmdline := strings.Replace(cfg.Application.Command, "{}", bannedIP, -1)
	log.Println("command:", cmdline)
	switch runtime.GOOS {
	case "windows":
		c = exec.Command("cmd", "/C", cmdline)
	default: //Mac & Linux
		c = exec.Command("bash", "-c", cmdline)
	}
	if err := c.Run(); err != nil {
		log.Println("Error: ", err)
	}
}
