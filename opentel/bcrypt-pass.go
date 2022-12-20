package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	var (
		cost int
		err  error
	)

	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)

	generateCmd.IntVar(&cost, "cost", 10, "bcrypt hash cost")

	if len(os.Args) < 2 {
		fmt.Println("generate or verify subcommand is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		generateCmd.Parse(os.Args[2:])
		err = generatePass(cost)
	case "verify":
		verifyCmd.Parse(os.Args[2:])
		err = verifyPass()
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err != nil {
		log.Fatalln(err)
	}
}

func generatePass(cost int) (err error) {
	var (
		username []byte
		password []byte
		bts      []byte
		reader   *bufio.Reader
	)

	reader = bufio.NewReader(os.Stdin)
	fmt.Print(">>> Username: ")
	if username, err = reader.ReadBytes('\n'); err != nil {
		return err
	}
	username = bytes.TrimSpace(username)

	fmt.Print(">>> Password: ")
	if password, err = reader.ReadBytes('\n'); err != nil {
		return err
	}
	password = bytes.TrimSpace(password)

	if bts, err = bcrypt.GenerateFromPassword(password, cost); err != nil {
		return err
	}

	fmt.Printf("%s: %s\n", username, bts)
	return nil
}

func verifyPass() (err error) {
	var (
		password []byte
		bha      []byte
		reader   *bufio.Reader
	)

	reader = bufio.NewReader(os.Stdin)
	fmt.Print(">>> Bcrypt adaptive hashing: ")
	if bha, err = reader.ReadBytes('\n'); err != nil {
		return err
	}
	bha = bytes.TrimSpace(bha)

	fmt.Print(">>> Password: ")
	if password, err = reader.ReadBytes('\n'); err != nil {
		return err
	}
	password = bytes.TrimSpace(password)

	if err = bcrypt.CompareHashAndPassword(bha, password); err != nil {
		fmt.Println("password don't match")
		return nil
	}

	fmt.Println("password verified")
	return nil
}
