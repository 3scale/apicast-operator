module github.com/3scale/apicast-operator

go 1.13

require (
	github.com/RHsyseng/operator-utils v0.0.0-20200213165520-1a022eb07a43
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.3.0
	github.com/go-playground/validator/v10 v10.4.0
	github.com/google/uuid v1.1.1
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/stretchr/testify v1.5.1
	k8s.io/api v0.19.14
	k8s.io/apimachinery v0.19.14
	k8s.io/client-go v0.19.14
	sigs.k8s.io/controller-runtime v0.7.2
)

// security release to address CVE-2020-14040
replace golang.org/x/text => golang.org/x/text v0.3.3

// security release to address CVE-2020-9283. First version
// that addresses the CVE is golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
// but we replace to the most recent version that appeared on go.sum before
// this change
replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
