package db

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
)

type Redis struct {
	Addr      string
	User      string
	Password  string
	ctx       context.Context
	rds       *redis.Client
	NroServer int64
}

const (
	KEY_SERVER_NRO = "SERVERNRO"
	CHANNEL_NAME   = "BANNEDIP"
)

func (r *Redis) Open(ctx context.Context) error {
	r.ctx = ctx
	r.rds = redis.NewClient(&redis.Options{
		Addr:     r.Addr,
		Username: r.User,
		Password: r.Password,
		DB:       0, //default db????
	})
	//Redis responde con un pong:="PONG" y err:=<nil>
	//_ , err := client.Ping(ctx).Result()
	err := r.rds.Ping(r.ctx).Err()
	if err != nil {
		log.Fatal("Error conectando con servidor Redis: ", err)
	}
	//obtengo mi ID unico en redis, utilizando una key con valores incrementables
	//como no existe SERVERNRO, redis la crea y la setea en 1. Luego, ya existe para los demas srvs
	myid, err := r.rds.Incr(r.ctx, KEY_SERVER_NRO).Result()
	if err != nil {
		log.Fatal("Imposible obtener ID unico de Servidor: ", err)
	}
	r.NroServer = myid

	return err //@TODO: mejorar la salida del error con wrapping de los msg

}

func (r *Redis) Publish(msg string) error {

	//envio data al canal
	err := r.rds.Publish(r.ctx, CHANNEL_NAME, msg).Err()
	if err != nil {
		log.Fatalf("Error tratando de publicar", err)
	}

	return err
}

func (r *Redis) Subscribe() chan string {
	//recibo data al canal
	mesg := make(chan string)

	go func() {
		sub := r.rds.Subscribe(r.ctx, CHANNEL_NAME)
		defer sub.Close()
		chsub := sub.Channel()
		for msg := range chsub {
			mesg <- msg.Payload
		}
	}()
	return mesg
}
func (r *Redis) Close() {
	r.rds.Close()
}
