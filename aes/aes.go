package main

/*
#include <termios.h>
#include <stdio.h>
#include <errno.h>

struct termios tio;

void noecho(int fd) {
	struct termios tio;
	tcgetattr(fd, &tio);
	tio.c_lflag &= ~ECHO;
	tcsetattr(fd, TCSANOW, &tio);
}

void savetermios(int fd) {
	tcgetattr(fd, &tio);
}

void resettermios(int fd) {
	tcsetattr(fd, TCSANOW, &tio);
}
*/
import "C"

import (
	"fmt"
	"os"
	"flag"
	"crypto/aes"
	"crypto/cipher"
)

const (
	CipherBlockSize = 32
)

func ReadPassword() []byte {
	var n int
	var err error
	
	tty, err := os.Open("/dev/tty")
	if err != nil {
		panic(err)
	}
	
	C.savetermios(C.int(tty.Fd()))
	C.noecho(C.int(tty.Fd()))
	
	data := make([]byte, CipherBlockSize)
	fmt.Fprint(os.Stderr, "Enter pass: ")
	n, err = tty.Read(data)
	tty.Close()
	if err != nil {
		panic(err)
	}
	
	C.resettermios(C.int(tty.Fd()))
	fmt.Fprint(os.Stderr, "\n")

	clean := false
	for i := 0; i < n; i++ {
		if data[i] == '\n' {
			clean = true
		}
		
		if clean {
			data[i] = 0
		}
	}

	return data
}

func clean(a []byte, n int) {
	for i := n; i < len(a); i++ {
		a[i] = 0
	}
}

func Encrypt(block cipher.Block, clear, encrypt *os.File) error {
	in := make([]byte, block.BlockSize())
	out := make([]byte, block.BlockSize())
	
	for {
		n, err := clear.Read(in)
		if n == 0 {
			break
		} else if err != nil {
			return err
		}
		
		clean(in, n)
		
		block.Encrypt(out, in)
		
		_, err = encrypt.Write(out)
		if err != nil {
			return err
		}
	}	
	
	return nil
}

func Decrypt(block cipher.Block, encrypt, clear *os.File) error {
	in := make([]byte, block.BlockSize())
	out := make([]byte, block.BlockSize())
	
	for {
		n, err := encrypt.Read(in)
		if n != block.BlockSize() {
			break
		} else if err != nil {
			return err
		}
		
		block.Decrypt(out, in)
		
		var i int
		for i = 0; i < n; i++ {
			if out[i] == 0 {
				break
			}
		}
		_, err = clear.Write(out[:i])
		if err != nil {
			return err
		}
	}	
	
	return nil
}

func main() {
	var output *os.File
	var err error
	
	decrypt := flag.Bool("d", false, "Decrypt given files to output")
	encrypt := flag.Bool("e", false, "Encrypt given files to output")
	outputPath := flag.String("o", "", "Redirect output to file")
	
	flag.Parse()
		if *decrypt && *encrypt || !*decrypt && !*encrypt {
		fmt.Errorf("Please specify one of d or e\n")
		os.Exit(1)
	}
	
	if *outputPath == "" {
		output = os.Stdout
	} else {
		output, err = os.Create(*outputPath)
		if err != nil {
			panic(err)
		}
	}
	
	pass := ReadPassword()
	block, err := aes.NewCipher(pass)
	if err != nil {
		panic(err)
	}
	
	for _, arg := range flag.Args() {
		var file *os.File
		
		if arg == "-" {
			file = os.Stdin
		} else {
			file, err = os.Open(arg)
			if err != nil {
				panic(err)
			}
		}
		
		if *encrypt {
			Encrypt(block, file, output)
		} else if *decrypt {
			Decrypt(block, file, output)
		}
	}
}
