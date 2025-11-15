package command

import (
	"fmt"
	"mangahub/cmd/cli/authentication"
	"mangahub/cmd/cli/command/client"
	"time"

	"github.com/spf13/cobra"
)

// auth.go handles authentication commands for the mangahubCLI application.
// auth login, register, logout, and token management commands will be implemented here.

// authCmd represents the auth command for authentication related subcommands
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  `Authenticate with the MangaHub API server. Supports login, registration, logout.`,
}

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new MangaHub account",
	RunE: func(cmd *cobra.Command, args []string) error {
		// get data from flags
		var c client.RegisterRequest
		c.Username, _ = cmd.Flags().GetString("username")
		c.Password, _ = cmd.Flags().GetString("password")
		c.Email, _ = cmd.Flags().GetString("email")

		// call API to register user
		htppClient := client.NewHTTPClient(apiURL) // create new HTTP client
		response, err := htppClient.Register(&c)
		if err != nil {
			return fmt.Errorf("registration process failed: %w", err)
		}

		// return confirmation message
		fmt.Println("✓ Registration successful! Please login to continue.")
		fmt.Printf("Username: %s\n", response.Username)
		return nil
	},
}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your MangaHub account",
	RunE: func(cmd *cobra.Command, args []string) error {
		// get data from flags
		var c client.LoginRequest
		c.Username, _ = cmd.Flags().GetString("username")
		c.Password, _ = cmd.Flags().GetString("password")

		// call API to login user
		htppClient := client.NewHTTPClient(apiURL) // create new HTTP client
		response, err := htppClient.Login(&c)
		if err != nil {
			return fmt.Errorf("login process failed: %w", err)
		}

		// store tokens in keyring
		now := time.Now().Unix()
		err = authentication.StoreTokens(&authentication.StoredCredentials{
			AccessToken:  response.AccessToken,
			RefreshToken: response.RefreshToken,
			Username:     c.Username,
			ExpiresAt:    now + response.ExpiresIn,
		})
		if err != nil {
			return fmt.Errorf("failed to store tokens: %w", err)
		}

		// return confirmation message
		fmt.Println("✓ Successfully logged in!")
		return nil
	},
}

var autologinCmd = &cobra.Command{
	Use:   "autologin",
	Short: "Automatically login using stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		// check if token is on the keyring already
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahub auth login'")
		}
		httpClient := client.NewHTTPClient(apiURL) // create new HTTP client
		req := &client.RefreshTokenRequest{RefreshToken: creds.RefreshToken}
		// if token is expired, refresh it
		now := time.Now().Unix()
		if now >= creds.ExpiresAt {
			refreshResp, err := httpClient.RefreshToken(req)
			if err != nil {
				return fmt.Errorf("session expired, please login again: %w", err)
			}

			// if refresh successful, store new tokens
			_ = authentication.StoreTokens(&authentication.StoredCredentials{
				AccessToken:  refreshResp.AccessToken,
				RefreshToken: refreshResp.RefreshToken,
				Username:     creds.Username,
				ExpiresAt:    now + refreshResp.ExpiresIn,
			})
		}
		return nil
	},
}

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from your MangaHub account",
	Run: func(cmd *cobra.Command, args []string) {
		// clear token from config
		err := authentication.DeleteTokens()
		if err != nil {
			fmt.Printf("✗ Logout failed: %v\n", err)
			return
		}
		// return confirmation message
		fmt.Println("✓ Successfully logged out.")
	},
}

// tokenCmd represents the token management command || implement later

// init function to add auth commands to root command
func init() {
	// add subcommands to authCmd
	authCmd.AddCommand(registerCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(autologinCmd)

	// add flags for register command
	registerCmd.Flags().StringP("username", "u", "", "Username for the new account")
	registerCmd.Flags().StringP("password", "p", "", "Password for the new account")
	registerCmd.Flags().StringP("email", "e", "", "Email address for the new account")
	registerCmd.MarkFlagRequired("username")
	registerCmd.MarkFlagRequired("password")
	registerCmd.MarkFlagRequired("email")

	// add flags for login command
	loginCmd.Flags().StringP("username", "u", "", "Username for the account")
	loginCmd.Flags().StringP("password", "p", "", "Password for the account")
	loginCmd.MarkFlagRequired("username")
	loginCmd.MarkFlagRequired("password")

	// add flags for logout command
	logoutCmd.Flags() // no flags for logout
}
