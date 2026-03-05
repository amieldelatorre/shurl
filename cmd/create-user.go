/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/amieldelatorre/shurl/internal"
	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	createUserCmdEmailInput        string
	createUserCmdUsernameInput     string
	createUserCmdPasswordInput     string
	createUserCmdPasswordFromStdin bool
)

// createUserCmd represents the createUser command
var createUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "Create a user without enabling registration",
	Long:  `Create a user without enabling registration`,
	Run: func(cmd *cobra.Command, args []string) {
		tempLogger := utils.NewCustomJsonLogger(os.Stdout, slog.LevelDebug)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(15)*time.Second)
		defer cancel()
		ctx = context.WithValue(ctx, utils.RequestIdName, "create-user")

		createUserCmdEmailInput = strings.TrimSpace(createUserCmdEmailInput)
		createUserCmdUsernameInput = strings.TrimSpace(createUserCmdUsernameInput)
		createUserCmdPasswordInput = strings.TrimSpace(createUserCmdPasswordInput)

		var password string
		var err error
		if createUserCmdPasswordInput == "" && !createUserCmdPasswordFromStdin { //
			tempLogger.ErrorExit(ctx, "--password is not provided and --password-stdin is not set. Please use one of the two.")
		} else if createUserCmdPasswordInput != "" && !createUserCmdPasswordFromStdin {
			password = createUserCmdPasswordInput
		} else {
			// if createUserCmdPasswordInput == "" && createUserCmdPasswordFromStdin
			// and createUserCmdPasswordInput != "" && createUserCmdPasswordFromStdin

			password, err = getPasswordFromStdin()
			if err != nil {
				tempLogger.ErrorExit(ctx, err.Error())
			}
		}

		configFilePath = strings.TrimSpace(configFilePath)
		app := internal.NewApp(configFilePath)

		req := handlers.PostUserRequest{
			Username:        createUserCmdUsernameInput,
			Email:           createUserCmdEmailInput,
			Password:        password,
			ConfirmPassword: password,
		}
		idempotencyKey, err := uuid.NewV7()
		if err != nil {
			tempLogger.ErrorExit(ctx, err.Error())
		}

		newUser, err := handlers.CreateUser(ctx, app.DbContext, req, idempotencyKey)
		if err != nil {
			tempLogger.ErrorExit(ctx, err.Error())
		}

		fmt.Printf("A new user, with email `%s` and username `%s` has been created\n", newUser.Email, newUser.Username)
	},
}

func init() {
	rootCmd.AddCommand(createUserCmd)
	createUserCmd.Flags().StringVar(&createUserCmdEmailInput, "email", "", "Email of the user to be created")
	err := createUserCmd.MarkFlagRequired("email")
	if err != nil {
		panic(err)
	}
	createUserCmd.Flags().StringVar(&createUserCmdUsernameInput, "username", "", "Username of the user to be created")
	err = createUserCmd.MarkFlagRequired("username")
	if err != nil {
		panic(err)
	}
	createUserCmd.Flags().StringVar(&createUserCmdPasswordInput, "password", "", "Password of the user to be created. WARNING: leading and trailing spaces will be trimmed")
	createUserCmd.Flags().BoolVar(&createUserCmdPasswordFromStdin, "password-stdin", false, "Take password from stdin, this takes precedence if both this and the password flag is given. WARNING: leading and trailing spaces will be trimmed")
}

func getPasswordFromStdin() (string, error) {
	var password string
	// if interactive terminal

	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Print("Enter password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Print("\n")

		if err != nil {
			return "", err
		}

		password = string(passwordBytes)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			password = scanner.Text()
		}
		err := scanner.Err()
		if err != nil {
			return "", err
		}
	}

	password = strings.TrimSpace(password)
	if password == "" {
		return "", errors.New("empty input from stdin for password")
	}

	return password, nil
}
