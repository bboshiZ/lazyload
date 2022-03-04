package module

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	bootconfig "slime.io/slime/framework/apis/config/v1alpha1"
	"slime.io/slime/framework/bootstrap"
	"slime.io/slime/framework/util"
)

type InitCallbacks struct {
	AddStartup func(func(ctx context.Context))
}

type Module interface {
	Name() string
	Config() proto.Message
	InitScheme(scheme *runtime.Scheme) error
	InitManager(mgr manager.Manager, env bootstrap.Environment, cbs InitCallbacks) error
}

func Main(bundle string, modules []Module) {
	fatal := func() {
		os.Exit(1)
	}

	type configH struct {
		config      *bootconfig.Config
		generalJson []byte
	}

	config, rawCfg, generalJson, err := bootstrap.GetModuleConfig("")
	if err != nil {
		panic(err)
	}
	if config == nil {
		panic(fmt.Errorf("module config nil for %s", bundle))
	}
	err = util.InitLog(config.Global.Log)
	if err != nil {
		panic(err)
	}

	log.Infof("load module config of %s: %s", bundle, string(rawCfg))
	modConfigs := map[string]configH{}
	isBundle := config.Bundle != nil
	if isBundle {
		for _, mod := range config.Bundle.Modules {
			modConfig, modRawCfg, modGeneralJson, err := bootstrap.GetModuleConfig(mod.Name)
			if err != nil {
				panic(err)
			}
			if config == nil {
				panic(fmt.Errorf("module config nil for %s", mod.Name))
			}

			if config.Global != nil {
				if modConfig.Global == nil {
					modConfig.Global = &bootconfig.Global{}
				}
				proto.Merge(config.Global, modConfig.Global)
			}
			log.Infof("load module config of bundle item %s: %s", mod.Name, string(modRawCfg))
			modConfigs[mod.Name] = configH{modConfig, modGeneralJson}
		}
	}

	var (
		scheme   = runtime.NewScheme()
		modNames []string
	)
	for _, mod := range modules {
		modNames = append(modNames, mod.Name())
		if err := mod.InitScheme(scheme); err != nil {
			log.Errorf("mod %s InitScheme met err %v", mod.Name(), err)
			fatal()
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: config.Global.Misc["metrics-addr"],
		Port:               9443,
		LeaderElection:     config.Global.Misc["enable-leader-election"] == "on",
		LeaderElectionID:   bundle,
	})
	if err != nil {
		log.Errorf("unable to start manager %s, %+v", bundle, err)
		fatal()
	}

	clientSet, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		log.Errorf("create a new clientSet failed, %+v", err)
		os.Exit(1)
	}

	remoteClients, err := getRemoteK8s(clientSet)
	if err != nil {
		log.Errorf("create a new remote k8s clientSet failed, %+v", err)
		os.Exit(1)
	}

	dynamicClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		log.Errorf("create a new dynamic client failed, %+v", err)
		os.Exit(1)
	}

	var startups []func(ctx context.Context)
	cbs := InitCallbacks{
		AddStartup: func(f func(ctx context.Context)) {
			startups = append(startups, f)
		},
	}

	ctx := context.Background()

	ph := bootstrap.NewPathHandler()

	for _, mod := range modules {
		var modCfg *bootconfig.Config
		var modGeneralJson []byte
		if isBundle {
			if h, ok := modConfigs[mod.Name()]; ok {
				modCfg, modGeneralJson = h.config, h.generalJson
			}
		} else {
			modCfg, modGeneralJson = config, generalJson
		}

		if modCfg != nil && modCfg.General != nil {
			modSelfCfg := mod.Config()
			if modSelfCfg != nil {
				if len(modCfg.General.XXX_unrecognized) > 0 {
					if err := proto.Unmarshal(modCfg.General.XXX_unrecognized, modSelfCfg); err != nil {
						log.Errorf("unmarshal for mod %s XXX_unrecognized (%v) met err %v", mod.Name(), modCfg.General.XXX_unrecognized, err)
						fatal()
					}
				} else if len(modGeneralJson) > 0 {
					if err := jsonpb.Unmarshal(bytes.NewBuffer(modGeneralJson), modSelfCfg); err != nil {
						log.Errorf("unmarshal for mod %s modGeneralJson (%v) met err %v", mod.Name(), modGeneralJson, err)
						fatal()
					}
				}
			}
		}

		env := bootstrap.Environment{
			Config:           modCfg,
			K8SClient:        clientSet,
			K8SRemoteClients: remoteClients,
			DynamicClient:    dynamicClient,
			HttpPathHandler: bootstrap.PrefixPathHandlerManager{
				Prefix:      mod.Name(),
				PathHandler: ph,
			},
			Stop: ctx.Done(),
		}

		if err := mod.InitManager(mgr, env, cbs); err != nil {
			log.Errorf("mod %s InitManager met err %v", mod.Name(), err)
			fatal()
		}
	}

	go bootstrap.AuxiliaryHttpServerStart(ph, config.Global.Misc["aux-addr"])

	for _, startup := range startups {
		startup(ctx)
	}

	log.Infof("starting manager bundle %s with modules %v", bundle, modNames)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Errorf("problem running manager, %+v", err)
		os.Exit(1)
	}
}

type RemoteSecretCluster struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}

type RemoteSecretClusterListItem struct {
	Cluster RemoteSecretCluster `yaml:"cluster"`
	Name    string              `yaml:"name"`
}

type RemoteSecretUser struct {
	Name string                `yaml:"name"`
	User RemoteSecretUserToken `yaml:"user"`
}

type RemoteSecretUserToken struct {
	Token string `yaml:"token"`
}
type RemoteSecret struct {
	APIVersion string                        `yaml:"apiVersion"`
	Clusters   []RemoteSecretClusterListItem `yaml:"clusters"`
	Contexts   []struct {
		Context struct {
			Cluster string `yaml:"cluster"`
			User    string `yaml:"user"`
		} `yaml:"context"`
		Name string `yaml:"name"`
	} `yaml:"contexts"`
	CurrentContext string `yaml:"current-context"`
	Kind           string `yaml:"kind"`
	Preferences    struct {
	} `yaml:"preferences"`
	Users []RemoteSecretUser `yaml:"users"`
}

// Point the k8s client to a remote cluster's API server
func UseRemoteCreds(remoteSecret *RemoteSecret) (*rest.Config, error) {
	caData := remoteSecret.Clusters[0].Cluster.CertificateAuthorityData
	rootCaDecoded, err := base64.StdEncoding.DecodeString(caData)
	if err != nil {
		return nil, err
	}
	// Basically implement rest.InClusterConfig() with the remote creds
	tlsClientConfig := rest.TLSClientConfig{
		CAData: []byte(rootCaDecoded),
	}

	serverParse := strings.Split(remoteSecret.Clusters[0].Cluster.Server, ":")
	if len(serverParse) != 3 && len(serverParse) != 2 {
		return nil, errors.New("Invalid remote API server URL")
	}
	host := strings.TrimPrefix(serverParse[1], "//")

	port := "443"
	if len(serverParse) == 3 {
		port = serverParse[2]
	}

	if !strings.EqualFold(serverParse[0], "https") {
		return nil, errors.New("Only HTTPS protocol is allowed in remote API server URL")
	}

	// There's no need to add the BearerToken because it's ignored later on
	return &rest.Config{
		Host:            "https://" + net.JoinHostPort(host, port),
		TLSClientConfig: tlsClientConfig,
	}, nil
}

func ParseRemoteSecretBytes(secretBytes []byte) (*RemoteSecret, error) {
	secret := &RemoteSecret{}
	err := yaml.Unmarshal(secretBytes, &secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func getRemoteK8s(clientSet *kubernetes.Clientset) ([]*kubernetes.Clientset, error) {
	var clientSets []*kubernetes.Clientset

	secrets, err := clientSet.CoreV1().Secrets("").List(metav1.ListOptions{LabelSelector: "istio/remoteKiali=true"})
	if err != nil {
		return nil, errors.New("failed to get remote k8s Secrets")

	}

	// log.Errorf("Error getRemoteK8s: %v", secrets)
	for _, secret := range secrets.Items {
		clusterName, ok := secret.Annotations["networking.istio.io/cluster"]
		if !ok {
			continue
		}

		kubeconfigFile, ok := secret.Data[clusterName]
		if !ok {
			log.Errorf("Error get secret: %v, %v", secret.Data, clusterName)
			continue
		}
		remoteSecret, parseErr := ParseRemoteSecretBytes(kubeconfigFile)
		if parseErr != nil {
			log.Errorf("Error ParseRemoteSecretBytes: %v", parseErr)
			continue
		}

		restConfig, restConfigErr := UseRemoteCreds(remoteSecret)
		if restConfigErr != nil {
			log.Errorf("Error using remote creds: %v", restConfigErr)
			continue
		}

		restConfig.Timeout = 15 * time.Second
		restConfig.BearerToken = remoteSecret.Users[0].User.Token

		clientSet, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			log.Errorf("Error using remote creds: %v", err)
			continue

		}
		clientSets = append(clientSets, clientSet)

	}

	return clientSets, nil
}
