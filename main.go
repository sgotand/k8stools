package main

import (
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// const (
// 	saDir     = "/var/run/secrets/kubernetes.io/serviceaccount/"
// 	tokenPath = saDir + "token"
// 	nsPath    = saDir + "namespace"
// 	caCrtPath = saDir + "ca.crt"
// )

func main() {
	err := Cmd().Run(os.Args)
	if err != nil {
		logrus.Fatalf("%+v", err)
	}
}

func Cmd() *cli.App {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{
		{
			Name:  "dump-in-cluster",
			Usage: "create kubeconfig from in-cluster config",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "out",
					Aliases: []string{"o"},
					Usage:   "output kubeconfig path",
					Value:   "kubeconfig",
				},
				&cli.StringFlag{
					Name:  "cluster",
					Usage: "target cluster name",
					Value: "default",
				},
				&cli.StringFlag{
					Name:  "user",
					Usage: "target user name",
					Value: "default",
				},
				&cli.StringFlag{
					Name:  "context",
					Usage: "target context name",
					Value: "default",
				},
			},
			Action: CmdCreateKubeConfigFileFromInClusterConfig,
		},
	}
	return app
}

type createKubeConfigOpts struct {
	ClusterName string
	UserName    string
	ContextName string
}

func CmdCreateKubeConfigFileFromInClusterConfig(c *cli.Context) error {
	opts := &createKubeConfigOpts{
		ClusterName: c.String("cluster"),
		UserName:    c.String("user"),
		ContextName: c.String("context"),
	}
	return createKubeConfigFileFromInClusterConfig(c.String("out"), opts)
}

func createKubeConfigFileFromInClusterConfig(filePath string, opts *createKubeConfigOpts) error {
	clusterName := "default"
	userName := "default"
	contextName := "default"
	if opts != nil && opts.ClusterName != "" {
		clusterName = opts.ClusterName
	}
	if opts != nil && opts.UserName != "" {
		userName = opts.UserName
	}
	if opts != nil && opts.ContextName != "" {
		contextName = opts.ContextName
	}

	icConfig, err := rest.InClusterConfig()
	if err != nil {
		return xerrors.Errorf("Failed to get InClusterConfig: %v", err)
	}

	clusters := make(map[string]*api.Cluster)
	authinfos := make(map[string]*api.AuthInfo)
	contexts := make(map[string]*api.Context)

	caCrtBytes, err := readFile(icConfig.CAFile)
	if err != nil {
		return xerrors.Errorf("failed to read CAFile: %v", err)
	}
	clusters[clusterName] = &api.Cluster{
		Server:                   icConfig.Host,
		CertificateAuthorityData: caCrtBytes,
	}

	tokenBytes, err := readFile(icConfig.BearerTokenFile)
	if err != nil {
		return xerrors.Errorf("failed to read BearerTokenFile: %v", err)
	}
	authinfos[userName] = &api.AuthInfo{
		Token: string(tokenBytes),
	}

	contexts[contextName] = &api.Context{
		Cluster:  clusterName,
		AuthInfo: userName,
	}

	config := api.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		AuthInfos:      authinfos,
		CurrentContext: contextName,
	}

	err = clientcmd.WriteToFile(config, filePath)
	if err != nil {
		return xerrors.Errorf("Failed to write Kubeconfig to %s: %v", filePath, err)
	}
	return nil
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, xerrors.Errorf("failed to open %s: %v", path, err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, xerrors.Errorf("failed to read %s: %v", path, err)
	}
	return data, nil
}
