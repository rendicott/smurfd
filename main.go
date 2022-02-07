/* Smurfdykt
"Something about secrets and do you keep them?"
Simple AWS Secrets Manager CLI application
Pulls secrets down from AWS secrets manager and returns them in
the desired formatting.
*/

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// make the config obj avail to this package globablly
var sc smurfConfig

// keyVal is a simple struct so we can parse
// the response from the secretmanager api
type keyVal struct {
	Key   string
	Value string
}

// parseSecret takes the secret string from the getsecret response
// and parses it for the desired tag key, returning just the value
// as a string
func parseSecret(tagname, secret string) (result string, err error) {
	var kvs []keyVal
	ss := []byte(secret)
	err = json.Unmarshal(ss, &kvs)
	if err != nil {
		return result, err
	}
	for _, tag := range kvs {
		if tag.Key == tagname {
			return tag.Value, err
		}
	}
	return result, err
}

// grabSecret takes the session and a configuration object and makes the API call
// to grab the secret. It then passes that to the parser and returns a string as
// a response.
func grabSecret(sc smurfConfig, smgr *secretsmanager.Client) (result string, err error) {
	gsvInput := secretsmanager.GetSecretValueInput{
		SecretId: &sc.SecretName,
	}
	response, err := smgr.GetSecretValue(context.TODO(), &gsvInput)
	if err != nil {
		return result, err
	}
	if !sc.Raw {
		result, err = parseSecret(sc.Tag, *response.SecretString)
	} else {
		return *response.SecretString, err
	}
	if err != nil {
		return result, err
	}
	return result, err
}

// smurfConfig is an internal struct for storing
// configuration needed to run this application
type smurfConfig struct {
	Profile    string `yaml:"profile"`
	SecretName string `yaml:"secret_name"`
	Tag        string `yaml:"tag"`
	Raw        bool   `yaml:"raw"`
}

// ParseConfigFile takes a yaml filename as input and
// attempts to parse it into a config object.
func (sc *smurfConfig) ParseConfigFile(filename string) (err error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, sc)
	return err
}

func main() {
	// build some config object junk
	var configFile string
	var err error
	flag.StringVar(&configFile, "config", "", "Filename of YAML configuration file. Contents overrides all parameters. Leave blank to use parameters only. ")
	flag.StringVar(&sc.Profile, "profile", "", "AWS session credentials profile, if blank default or instance profile will be attempted.")
	flag.StringVar(&sc.Tag, "tag", "username", "the tag key name to grab from the secret. The value of this key will be returned")
	flag.StringVar(&sc.SecretName, "secretname", "foo", "name of secret to retrieve")
	flag.BoolVar(&sc.Raw, "raw", false, "pull down the raw value of the secret instead of tag parsing.")

	flag.Parse()
	// honor the config file as an override.
	if configFile != "" {
		err = sc.ParseConfigFile(configFile)
		if err != nil {
			log.Printf("Error parsing config file: '%s'.  Continuing with parameter defaults", err.Error())
		}
	}
	// set up a session based on profile entry or instance profile
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(sc.Profile),
	)
	if err != nil {
		log.Println("Error loading default or shared profile credentials")
		os.Exit(1)
	}
	svc := secretsmanager.NewFromConfig(cfg)
	result, err := grabSecret(sc, svc)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", result)
	os.Exit(0)
}
