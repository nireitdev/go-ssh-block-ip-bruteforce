//Simple polling del archivo de log,
//En cada segundo verifico si se agregaron nuevas lineas
//
// Deberia usar orientado a eventos en el filesystem:
//@TODO: utilizar a futuro alguna libreria como fsnotify, go-tailer, etc
//

package logparser

import (
	"bufio"
	"io"
	"log"
	"os"
	"regexp"
	"time"
)

type Logfile struct {
	Filename  string
	Searchreg string
	Filterreg string
	lines     chan string
}

func (lf *Logfile) Run() (listadoIPs chan string) {
	lf.lines = make(chan string)

	go lf.tail()
	return lf.lines
}
func (lf *Logfile) tail() {

	f, err := os.Open(lf.Filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	info, err := f.Stat()
	if err != nil {
		log.Fatalln("Error leyendo Log. ", err)
	}
	oldSize := info.Size()

	for {
		// Leo el contenido del archivo hasta el EOF:
		for line, err := r.ReadString('\n'); err != io.EOF; line, err = r.ReadString('\n') {

			matched, _ := regexp.MatchString(lf.Searchreg, line)
			if matched {
				pattern := regexp.MustCompile(lf.Filterreg)
				firstMatchSubstring := pattern.FindString(line)
				lf.lines <- firstMatchSubstring
			}

		}

		pos, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			log.Fatalln(err)
		}

		//Polling cada 1 seg:
		for {
			time.Sleep(time.Second)
			newinfo, err := f.Stat()
			if err != nil {
				panic(err)
			}
			newSize := newinfo.Size()
			if newSize != oldSize {
				//el archivo cambio de tamaño:
				if newSize < oldSize {
					f.Seek(0, 0)
				} else {
					f.Seek(pos, io.SeekStart)
				}
				r = bufio.NewReader(f)
				oldSize = newSize
				//y salgo de este bucle para leer lo nuevo en el bucle for() superior:
				break
			}
		}
	}
}
