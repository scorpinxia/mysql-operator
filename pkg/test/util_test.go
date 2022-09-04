package test

import (
	"fmt"
	"github.com/scorpinxia/mysql-operator/pkg/util"
	"testing"
)

func TestMakeStatefulSet(t *testing.T) {
	ss := util.MakeStatefulSetMysql("vke-system", "8.0.27")
	fmt.Println(ss)
}

func TestMakeSecret(t *testing.T) {
	secret := util.MakeSecretMysql("vke-system")
	fmt.Println(secret)
}

func TestMakeService(t *testing.T) {
	service := util.MakeServiceMysql("vke-system")
	fmt.Println(service)
}
