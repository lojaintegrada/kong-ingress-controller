//+build integration_tests

package integration

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters/addons/knative"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters/addons/kong"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters/addons/metallb"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters/types/gke"
	"github.com/kong/kubernetes-testing-framework/pkg/clusters/types/kind"
	"github.com/kong/kubernetes-testing-framework/pkg/environments"
	"github.com/kong/kubernetes-testing-framework/pkg/utils/kubernetes/generators"
	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kong/kubernetes-ingress-controller/internal/cmd/rootcmd"
	"github.com/kong/kubernetes-ingress-controller/internal/manager"
)

// -----------------------------------------------------------------------------
// Testing Main
// -----------------------------------------------------------------------------

func TestMain(m *testing.M) {
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("INFO: setting up test environment")
	kongbuilder := kong.NewBuilder()
	kongAdminPwd := ""
	if enterpriseEnablement == "on" {
		if enterpriseRepo != "" && enterpriseTag != "" {
			licenseJSON := os.Getenv("KONG_ENTERPRISE_LICENSE")
			if licenseJSON == "" {
				exitOnErr(fmt.Errorf(("enterprise can not be installed w/o a license json.")))
			}
			password, err := password.Generate(10, 5, 0, false, false)
			if err != nil {
				kongAdminPwd = KongTestPassword
			} else {
				kongAdminPwd = password
			}
			kongbuilder = kongbuilder.WithEnterprise().
				WithImage(enterpriseRepo, enterpriseTag).
				WithEnterpriseLicense(licenseJSON).
				WithKongAdminPassword(kongAdminPwd).
				WithAdminServiceTypeLoadBalancer()
		} else {
			exitOnErr(fmt.Errorf(("enterprise repo and tag is not configured.")))
		}
	}

	if dbmode == "postgres" {
		kongbuilder = kongbuilder.WithPostgreSQL()
	}

	kongbuilder.WithControllerDisabled()
	kongAddon := kongbuilder.Build()
	builder := environments.NewBuilder().WithAddons(kongAddon, knative.New())

	fmt.Println("INFO: configuring cluster for testing environment")
	if existingCluster != "" {
		if clusterVersionStr != "" {
			exitOnErrWithCode(fmt.Errorf("can't flag cluster version & provide an existing cluster at the same time"), ExitCodeIncompatibleOptions)
		}
		clusterParts := strings.Split(existingCluster, ":")
		if len(clusterParts) != 2 {
			exitOnErrWithCode(fmt.Errorf("existing cluster in wrong format (%s): format is <TYPE>:<NAME> (e.g. kind:test-cluster)", existingCluster), ExitCodeCantUseExistingCluster)
		}
		clusterType, clusterName := clusterParts[0], clusterParts[1]

		fmt.Printf("INFO: using existing %s cluster %s\n", clusterType, clusterName)
		switch clusterType {
		case string(kind.KindClusterType):
			cluster, err := kind.NewFromExisting(clusterName)
			exitOnErr(err)
			builder.WithExistingCluster(cluster)
			builder.WithAddons(metallb.New())
		case string(gke.GKEClusterType):
			cluster, err := gke.NewFromExistingWithEnv(ctx, clusterName)
			exitOnErr(err)
			builder.WithExistingCluster(cluster)
		default:
			exitOnErrWithCode(fmt.Errorf("unknown cluster type: %s", clusterType), ExitCodeCantUseExistingCluster)
		}
	} else {
		fmt.Println("INFO: no existing cluster found, deploying using Kubernetes In Docker (KIND)")
		builder.WithAddons(metallb.New())
	}
	if clusterVersionStr != "" {
		clusterVersion, err := semver.Parse(strings.TrimPrefix(clusterVersionStr, "v"))
		exitOnErr(err)
		cluster, err := kind.NewBuilder().WithClusterVersion(clusterVersion).Build(ctx)
		exitOnErr(err)
		builder.WithExistingCluster(cluster)
	}

	fmt.Println("INFO: building test environment")
	var err error
	env, err = builder.Build(ctx)
	exitOnErr(err)

	fmt.Printf("INFO: reconfiguring the kong admin service as LoadBalancer type\n")
	svc, err := env.Cluster().Client().CoreV1().Services(kongAddon.Namespace()).Get(ctx, kong.DefaultAdminServiceName, metav1.GetOptions{})
	exitOnErr(err)
	svc.Spec.Type = corev1.ServiceTypeLoadBalancer
	_, err = env.Cluster().Client().CoreV1().Services(kongAddon.Namespace()).Update(ctx, svc, metav1.UpdateOptions{})
	exitOnErr(err)

	fmt.Printf("INFO: waiting for cluster %s and all addons to become ready\n", env.Cluster().Name())
	exitOnErr(<-env.WaitForReady(ctx))

	fmt.Println("INFO: collecting urls from the kong proxy deployment")
	proxyURL, err = kongAddon.ProxyURL(ctx, env.Cluster())
	exitOnErr(err)
	proxyAdminURL, err = kongAddon.ProxyAdminURL(ctx, env.Cluster())
	exitOnErr(err)
	proxyUDPURL, err = kongAddon.ProxyUDPURL(ctx, env.Cluster())
	exitOnErr(err)

	fmt.Println("INFO: generating unique namespaces for each test case")
	testCases, err := identifyTestCasesForDir("./")
	exitOnErr(err)
	for _, testCase := range testCases {
		namespaceForTestCase, err := generators.GenerateNamespace(ctx, env.Cluster(), testCase)
		exitOnErr(err)
		namespaces[testCase] = namespaceForTestCase
		watchNamespaces = fmt.Sprintf("%s,%s", watchNamespaces, namespaceForTestCase.Name)
	}

	if v := os.Getenv("KONG_BRING_MY_OWN_KIC"); v == "true" {
		fmt.Println("WARNING: caller indicated that they will manage their own controller")
	} else {
		exitOnErr(deployControllers(ctx, controllerNamespace, enterpriseEnablement, kongAdminPwd))
	}

	fmt.Println("INFO: running final testing environment checks")
	clusterVersion, err = env.Cluster().Version()
	exitOnErr(err)

	fmt.Printf("INFO: testing environment is ready KUBERNETES_VERSION=(%v): running tests\n", clusterVersion)
	code := m.Run()

	if keepTestCluster == "" && existingCluster == "" {
		ctx, cancel := context.WithTimeout(context.Background(), environmentCleanupTimeout)
		defer cancel()
		fmt.Printf("INFO: cluster %s is being deleted\n", env.Cluster().Name())
		exitOnErr(env.Cleanup(ctx))
	}

	os.Exit(code)
}

// -----------------------------------------------------------------------------
// Testing Main - Controller Deployment
// -----------------------------------------------------------------------------

var crds = []string{
	"../../config/crd/bases/configuration.konghq.com_udpingresses.yaml",
	"../../config/crd/bases/configuration.konghq.com_tcpingresses.yaml",
	"../../config/crd/bases/configuration.konghq.com_kongplugins.yaml",
	"../../config/crd/bases/configuration.konghq.com_kongingresses.yaml",
	"../../config/crd/bases/configuration.konghq.com_kongconsumers.yaml",
	"../../config/crd/bases/configuration.konghq.com_kongclusterplugins.yaml",
}

// deployControllers ensures that relevant CRDs and controllers are deployed to the test cluster
func deployControllers(ctx context.Context, namespace string, enterprise, adminPassword string) error {
	// ensure the controller namespace is created
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
	if _, err := env.Cluster().Client().CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{}); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}

	// we'll wait until the controller has started before returning
	wg := sync.WaitGroup{}
	wg.Add(1)

	// run the controller in the background
	go func() {
		// convert the cluster rest.Config into a kubeconfig
		yaml, err := generators.NewKubeConfigForRestConfig(env.Cluster().Name(), env.Cluster().Config())
		exitOnErr(err)

		// create a tempfile to hold the cluster kubeconfig that will be used for the controller
		kubeconfig, err := os.CreateTemp(os.TempDir(), "kubeconfig-")
		exitOnErr(err)
		defer os.Remove(kubeconfig.Name())
		defer kubeconfig.Close()

		// dump the kubeconfig from kind into the tempfile
		c, err := kubeconfig.Write(yaml)
		exitOnErr(err)
		if c != len(yaml) {
			exitOnErr(fmt.Errorf("could not write entire kubeconfig file (%d/%d bytes)", c, len(yaml)))
		}
		kubeconfig.Close()

		// deploy our CRDs to the cluster
		for _, crd := range crds {
			cmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig.Name(), "apply", "-f", crd) //nolint:gosec
			stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			if err := cmd.Run(); err != nil {
				exitOnErr(fmt.Errorf("%s: %w", stderr.String(), err))
			}
		}

		config := manager.Config{}
		flags := config.FlagSet()
		basicParams := []string{
			fmt.Sprintf("--kong-admin-url=http://%s:8001", proxyAdminURL.Hostname()),
			fmt.Sprintf("--kubeconfig=%s", kubeconfig.Name()),
			"--election-id=integrationtests.konghq.com",
			"--publish-service=kong-system/ingress-controller-kong-proxy",
			fmt.Sprintf("--watch-namespace=%s", watchNamespaces),
			fmt.Sprintf("--ingress-class=%s", ingressClass),
			"--log-level=trace",
			"--log-format=text",
			"--debug-log-reduce-redundancy",
			"--admission-webhook-listen=172.17.0.1:49023",
			fmt.Sprintf("--admission-webhook-cert=%s", admissionWebhookCert),
			fmt.Sprintf("--admission-webhook-key=%s", admissionWebhookKey),
			"--profiling",
			"--dump-config",
		}

		if enterpriseEnablement == "on" {
			adminURL := fmt.Sprintf("http://%s:8001", proxyAdminURL.Hostname())

			workspace := "non-default-workspace"
			if err := createNonDefaultWorkspace(adminURL, workspace, adminPassword); err != nil {
				panic("failed creating non-default workspace through kong admin api.")
			}

			userName := "non-admin-usr"
			userToken := "non-admin-usr"
			if err := createNonAdminUser(adminURL, workspace, userName, userToken, adminPassword); err != nil {
				panic("failed creating non-admin user through kong admin api.")
			}

			role := "non-admin-role"
			if err := createwsRoleAndPermission(adminURL, workspace, role, adminPassword); err != nil {
				panic("failed creating role and endpoints through kong admin api.")
			}

			if err := addUserIntoRole(adminURL, workspace, userName, role, adminPassword); err != nil {
				panic("failed adding user into role kong admin api.")
			}

			enterpriseParams := []string{
				fmt.Sprintf("--kong-admin-token=%s", adminPassword),
				fmt.Sprintf("--kong-workspace=%s", workspace),
				fmt.Sprintf("--kong-non-admin-token=%s", userToken),
			}
			basicParams = append(basicParams, enterpriseParams...)
		}
		exitOnErr(flags.Parse(basicParams))
		fmt.Fprintf(os.Stderr, "INFO: Starting Controller Manager with Configuration: %+v\n", config)
		wg.Done()
		exitOnErr(rootcmd.Run(ctx, &config))
	}()

	wg.Wait()
	return nil
}

func createNonDefaultWorkspace(adminURL, workspaceName, adminPassword string) error {
	url := adminURL + "/workspaces"
	fmt.Println("add non-default workspace URL:>", url)

	var jsonStr = []byte(`{"name":"` + workspaceName + `"}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("kong-admin-token", adminPassword)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if 201 == resp.StatusCode {
		fmt.Printf("successfully created workspace %s through admin api.", body)
		return nil
	}
	return fmt.Errorf("failed creating workspace %s ", body)
}

func createNonAdminUser(adminURL, workspace, user, userToken, adminPassword string) error {
	url := adminURL + "/" + workspace + "/rbac/users"
	fmt.Println("add rbac user URL:>", url)

	str := `{"enabled": true , "name":"` + user + `","user_token":"` + userToken + `"}`
	fmt.Printf("json %s", str)
	var jsonStr = []byte(str)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("kong-admin-token", adminPassword)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if 201 == resp.StatusCode {
		fmt.Printf("successfully created rbac user %s through admin api.", body)
		return nil
	}

	return fmt.Errorf("failed creating rbac user %s.", body)
}

func createwsRoleAndPermission(adminURL, workspace, role, adminPassword string) error {
	url := adminURL + "/" + workspace + "/rbac/roles"
	fmt.Println("add role URL:>", url)

	var jsonStr = []byte(`{"name": "` + role + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("kong-admin-token", adminPassword)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if 201 == resp.StatusCode {
		fmt.Printf("successfully created workspace role, %s", body)
	} else {
		return fmt.Errorf("failed creating role %s.", body)
	}

	url = adminURL + "/" + workspace + "/rbac/roles/" + role + "/endpoints/"
	fmt.Println("add role endpoints URL:>", url)
	jsonStr = []byte(`{"endpoint": "*", "workspace": "` + workspace + `",  "actions":"*"}`)
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("kong-admin-token", adminPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ = ioutil.ReadAll(resp.Body)
	if 201 == resp.StatusCode {
		fmt.Printf("successfully added role permssions, %s", body)
		return nil
	}

	return fmt.Errorf("failed creating role permssions %s.", body)
}

func addUserIntoRole(adminURL, workspace, user, role, adminPassword string) error {
	url := adminURL + "/" + workspace + "/rbac/users/" + user + "/roles/"
	fmt.Println("add user into role endpoints URL:>", url)

	var jsonStr = []byte(`{"roles": "` + role + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("kong-admin-token", adminPassword)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if 201 == resp.StatusCode {
		fmt.Printf("successfully added role permssions, %s", body)
		return nil
	}

	return fmt.Errorf("failed adding role permssions %s.", body)
}
