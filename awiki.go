/* Application Personal Wiki

 */
package main

import (
	"log"
	"os"

	page "github.com/akruszewski/awiki/page/git"
	settings "github.com/akruszewski/awiki/settings"
	"github.com/akruszewski/awiki/webservice"
	awikiSSH "github.com/akruszewski/awiki/webservice/auth/ssh"
	"github.com/urfave/cli"
	//    "gopkg.in/yaml.v2"
)

// TODO: move it to separate yaml file
var WikiPath = os.Getenv("WIKIPATH")

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Initialize Wiki git repository.",
			Action: func(c *cli.Context) error {
				log.Print("Initializing git repo in ", WikiPath)
				_, err := page.Init(WikiPath)
				if err != nil {
					os.Exit(-1)
				}
				return nil
			},
		},
		{
			Name:    "generate-key",
			Aliases: []string{"g"},
			Usage:   "Generate SSH keys pair.",
			Action: func(c *cli.Context) error {
				log.Print("Generating ssh public and private keys.")
				err := awikiSSH.MakeSSHKeyPair(
					settings.PrivateKeyPath,
					settings.PublicKeyPath,
				)
				if err != nil {
					os.Exit(-1)
				}
				return nil
			},
		},
		{
			Name:    "runserver",
			Aliases: []string{"r"},
			Usage:   "Run Wiki server.",
			Action: func(c *cli.Context) error {
				webservice.RunServer()
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
