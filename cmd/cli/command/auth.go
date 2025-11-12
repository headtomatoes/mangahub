package command

import (
	"fmt"
	"mangahub/cmd/cli/command/client"

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
		fmt.Printf("UserID: %s\n", response.UserID)
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

		// save token to config
		saveToken(response.AccessToken, response.RefreshToken)

		// return confirmation message
		fmt.Println("✓ Successfully logged in!")
		return nil
	},
}

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from your MangaHub account",
	Run: func(cmd *cobra.Command, args []string) {
		// clear token from config
		saveToken("", "")
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

// saveToken saves the authentication token to config file || implement later
func saveToken(accessToken, refreshToken string) {
	// TODO: Implement token persistence
	// For now, store in global variable
	token = accessToken
	fmt.Printf("Access Token: %s\n", accessToken)
	if refreshToken != "" {
		fmt.Printf("Refresh Token: %s\n", refreshToken)
	}
	// Future: Save to ~/.mangahub/config.json
}
