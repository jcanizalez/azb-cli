package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/kataras/tablewriter"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	apiclient "azb/client/client"
	"net/http"
	neturl "net/url"
)

var cfgFile string
var output string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "azb",
	Short: "azb command line tool",
	Long: `
azb is a CLI to handle remote terraform workspace and modules in organizations 
and handle all the lifecycle (plan, apply, destroy).`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.azb-cli.yaml)")
	rootCmd.PersistentFlags().StringVar(&output, "output", "json", "Use json, table, tsv or none to format CLI output")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	cobra.AddTemplateFunc("StyleHeading", color.New(color.FgCyan).SprintFunc())
	usageTemplate := rootCmd.UsageTemplate()
	usageTemplate = strings.NewReplacer(
		`Usage:`, `{{StyleHeading "Usage:"}}`,
		`Aliases:`, `{{StyleHeading "Aliases:"}}`,
		`Available Commands:`, `{{StyleHeading "Commands:"}}`,
		`Global Flags:`, `{{StyleHeading "Global Flags:"}}`,
		`Examples:`, `{{StyleHeading "Examples:"}}`,
	).Replace(usageTemplate)
	re := regexp.MustCompile(`(?m)^Flags:\s*$`)
	usageTemplate = re.ReplaceAllLiteralString(usageTemplate, `{{StyleHeading "Flags:"}}`)
	rootCmd.SetUsageTemplate(usageTemplate)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".azb-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".azb-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func newClient() apiclient.Client {
	home, err := homedir.Dir()
	cobra.CheckErr(err)
	viper.AddConfigPath(home)
	viper.SetConfigName("azb.yml")
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}
	url := viper.Get("Server").(string)
	token := viper.Get("Token").(string)
	baseUrl := &neturl.URL{
		Scheme: "http",
		Host:   url,
	}
	client := apiclient.NewClient(&http.Client{}, token, baseUrl)
	return *client
}

func renderOutput(result interface{}, format string) {
	switch {
	case format == "json":
		printJSON, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate json", err)
		}
		fmt.Printf("%s\n", string(printJSON))
	case format == "tsv":
		data, _ := splitInterface(result)
		for _, v := range data {
			fmt.Println(strings.Join(v[:], "\t"))
		}
	case format == "table":
		data, header := splitInterface(result)
		if len(data) > 0 {
			table := tablewriter.NewWriter(os.Stdout)
			table.AppendBulk(data)
			table.SetHeader(header)
			table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
			table.SetCaption(true, " ")
			table.SetCenterSeparator("|")
			table.Render()
		}
	case format == "none":

	}
}

func splitInterface(input interface{}) ([][]string, []string) {
	reflectData := reflect.ValueOf(input)
	headers := make([]string, 0)
	headers = append(headers, "ID")
	result := make([][]string, 0)
	if reflectData.Kind() == reflect.Slice {
		for i := 0; i < reflectData.Len(); i++ {
			data := reflectData.Index(i).Interface()
			d := reflect.Indirect(reflect.ValueOf(data))

			row := make([]string, 0)
			id := d.FieldByName("ID").String()
			row = append(row, id)

			attr := reflect.Indirect(reflect.ValueOf(d.FieldByName("Attributes").Interface()))
			for j := 0; j < attr.NumField(); j++ {
				if i == 0 {
					headers = append(headers, attr.Type().Field(j).Name)
				}
				row = append(row, attr.Field(j).String())
			}
			result = append(result, row)
		}
	} else {
		d := reflect.Indirect(reflectData)
		row := make([]string, 0)
		id := d.FieldByName("ID").String()
		row = append(row, id)

		attr := reflect.Indirect(reflect.ValueOf(d.FieldByName("Attributes").Interface()))
		for j := 0; j < attr.NumField(); j++ {
			headers = append(headers, attr.Type().Field(j).Name)
			row = append(row, attr.Field(j).String())
		}
		result = append(result, row)

	}
	return result, headers
}
